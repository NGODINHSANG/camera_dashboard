package router

import (
	"net/http"
	"time"

	"camera-dashboard-backend/internal/config"
	"camera-dashboard-backend/internal/handlers"
	"camera-dashboard-backend/internal/middleware"
	"camera-dashboard-backend/internal/repository"
	"camera-dashboard-backend/internal/services"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jmoiron/sqlx"
)

func Setup(db *sqlx.DB, cfg *config.Config) *chi.Mux {
	r := chi.NewRouter()

	// Global middleware
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.RequestID)
	r.Use(middleware.CORSMiddleware(cfg.CORSOrigins))

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	cameraRepo := repository.NewCameraRepository(db)
	permRepo := repository.NewProjectPermissionRepository(db)
	chatRepo := repository.NewChatRepository(db)

	// Initialize services
	authService := services.NewAuthService(userRepo, cfg.JWTSecret)
	recorderService := services.NewRecorderService(db)

	// Initialize video processor for uploads
	videoProcessor := services.NewVideoProcessor(db, cfg.RecordingsPath)

	// Start chat hub
	go services.GlobalChatHub.Run()

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService, permRepo)
	projectHandler := handlers.NewProjectHandler(projectRepo, cameraRepo)
	cameraHandler := handlers.NewCameraHandler(cameraRepo, projectRepo, permRepo)
	adminHandler := handlers.NewAdminHandler(userRepo, projectRepo, cameraRepo, permRepo)
	recordingHandler := handlers.NewRecordingHandler(db, recorderService, cfg.RecordingsPath)
	webhookHandler := handlers.NewWebhookHandler(db, recorderService, cfg.RecordingsPath)
	uploadHandler := handlers.NewUploadHandler(db, cameraRepo, projectRepo, permRepo, videoProcessor)
	chunkUploadHandler := handlers.NewChunkUploadHandler(db, cameraRepo, projectRepo, permRepo, videoProcessor)
	chatHandler := handlers.NewChatHandler(chatRepo, cfg.RecordingsPath, services.GlobalChatHub, cfg.JWTSecret)

	// Wire up auto-record callbacks
	cameraHandler.SetAutoRecordCallback(webhookHandler.TriggerAutoRecordForCamera)
	cameraHandler.SetAutoRecordDisableCallback(webhookHandler.StopRecordingForCamera)

	// Initialize HLS cache (local disk for segments, avoids SMB reads)
	services.InitHLSCache(cfg.HLSCachePath, cfg.RecordingsPath)

	// Start HLS converter worker (scans every 2 minutes)
	hlsConverter := services.NewHLSConverter(cfg.RecordingsPath, db)
	hlsConverter.StartWorker(2 * time.Minute)

	// Start periodic auto-record check (every 30 seconds)
	// This ensures cameras with auto_record=1 and active streams are always recording
	go func() {
		time.Sleep(10 * time.Second) // Wait for MediaMTX to be fully ready
		webhookHandler.StartPeriodicCheck(30 * time.Second)
	}()

	// Routes
	r.Route("/api", func(r chi.Router) {
		// Health check (public)
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok"}`))
		})

		// Internal webhook routes (from MediaMTX - no auth required)
		r.Route("/internal/webhooks", func(r chi.Router) {
			r.Post("/stream-ready", webhookHandler.HandleStreamReady)
			r.Post("/stream-not-ready", webhookHandler.HandleStreamNotReady)
		})

		// Public video streaming routes (no auth required for video element)
		r.Get("/recordings/files/stream", recordingHandler.StreamVideoFile)
		r.Get("/recordings/files/download", recordingHandler.DownloadVideoFile)
		r.Get("/recordings/hls", recordingHandler.ServeHLSFile)
		r.Get("/recordings/restream", recordingHandler.LiveRestream)

		// Chat WebSocket — must be outside AuthMiddleware; ServeWS authenticates via query param token
		r.Get("/chat/ws", chatHandler.ServeWS)
		r.Get("/chat/attachments/{filename}", chatHandler.ServeAttachment)

		// Auth routes (public)
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)

			// Protected auth routes
			r.Group(func(r chi.Router) {
				r.Use(middleware.AuthMiddleware(cfg.JWTSecret))
				r.Get("/me", authHandler.GetMe)
			})
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthMiddleware(cfg.JWTSecret))

			// Projects routes
			r.Route("/projects", func(r chi.Router) {
				r.Get("/", projectHandler.GetAll)
				r.Post("/", projectHandler.Create)
				r.Get("/{id}", projectHandler.GetByID)
				r.Put("/{id}", projectHandler.Update)
				r.Delete("/{id}", projectHandler.Delete)

				// Cameras nested under projects
				r.Route("/{projectId}/cameras", func(r chi.Router) {
					r.Get("/", cameraHandler.GetByProject)
					r.Post("/", cameraHandler.Create)
					r.Put("/{cameraId}", cameraHandler.Update)
					r.Delete("/{cameraId}", cameraHandler.Delete)

					// Upload video to camera (global admin or project admin)
					r.Post("/{cameraId}/upload", uploadHandler.Upload)
					r.Get("/{cameraId}/upload-jobs", uploadHandler.GetCameraJobs)
				})

				// Upload job status
				r.Get("/upload-jobs/{jobId}", uploadHandler.GetJobStatus)
				r.Delete("/upload-jobs/{jobId}", uploadHandler.CancelJob)

				// Chunked upload routes (global admin or project admin)
				r.Route("/chunked-upload", func(r chi.Router) {
					r.Post("/init/{projectId}/{cameraId}", chunkUploadHandler.InitUpload)
					r.Post("/chunk/{uploadId}", chunkUploadHandler.UploadChunk)
					r.Post("/complete/{uploadId}", chunkUploadHandler.CompleteUpload)
					r.Get("/status/{uploadId}", chunkUploadHandler.GetUploadStatus)
					r.Delete("/cancel/{uploadId}", chunkUploadHandler.CancelUpload)
				})
			})

			// Recordings routes
			r.Route("/recordings", func(r chi.Router) {
				r.Get("/", recordingHandler.GetRecordings)
				r.Get("/active", recordingHandler.GetActiveRecordings)
				r.Post("/start", recordingHandler.StartRecording)
				r.Get("/browse", recordingHandler.BrowseDirectory)
				r.Post("/validate-path", recordingHandler.ValidatePath)
				r.Get("/{id}", recordingHandler.GetRecording)
				r.Post("/{id}/stop", recordingHandler.StopRecording)
				r.Delete("/{id}", recordingHandler.DeleteRecording)
				r.Get("/{id}/download", recordingHandler.DownloadRecording)
				r.Get("/{id}/stream", recordingHandler.StreamRecording)
				r.Get("/camera/{cameraId}/status", recordingHandler.GetCameraRecordingStatus)
				r.Get("/camera/{cameraId}/list", recordingHandler.GetCameraRecordings)

				// File-based recording routes (new)
				r.Get("/files/{projectName}/{cameraName}", recordingHandler.ListCameraRecordingFiles)
				r.Get("/files/stream", recordingHandler.StreamVideoFile)
				r.Get("/files/download", recordingHandler.DownloadVideoFile)
				r.Delete("/files/delete", recordingHandler.DeleteVideoFile)
			})

			// Chat routes
			r.Route("/chat", func(r chi.Router) {
				r.Get("/online", chatHandler.GetOnlineUsers)
				r.Get("/users/search", chatHandler.SearchUsers)
				r.Get("/conversations", chatHandler.GetConversations)
				r.Post("/conversations", chatHandler.CreateConversation)
				r.Put("/conversations/{id}/name", chatHandler.RenameGroup)
				r.Post("/conversations/{id}/members/{userId}", chatHandler.AddMember)
				r.Delete("/conversations/{id}/members/{userId}", chatHandler.RemoveMember)
				r.Get("/conversations/{id}/messages", chatHandler.GetMessages)
				r.Post("/conversations/{id}/messages", chatHandler.SendMessage)
				r.Post("/conversations/{id}/attachments", chatHandler.UploadAttachment)
				r.Post("/conversations/{id}/read", chatHandler.MarkRead)
			})

			// Admin routes
			r.Route("/admin", func(r chi.Router) {
				r.Use(middleware.AdminOnlyMiddleware)

				r.Get("/users", adminHandler.GetAllUsers)
				r.Get("/users/{id}", adminHandler.GetUser)
				r.Delete("/users/{id}", adminHandler.DeleteUser)
				r.Put("/users/{id}/role", adminHandler.UpdateUserRole)
				r.Get("/users/{id}/project-permissions", adminHandler.GetUserProjectPermissions)
				r.Get("/projects", adminHandler.GetAllProjects)

				// Project permission management
				r.Get("/projects/{projectId}/members", adminHandler.GetProjectMembers)
				r.Post("/projects/{projectId}/members/{userId}", adminHandler.GrantProjectPermission)
				r.Delete("/projects/{projectId}/members/{userId}", adminHandler.RevokeProjectPermission)
			})
		})
	})

	return r
}
