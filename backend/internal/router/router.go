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

	// Initialize services
	authService := services.NewAuthService(userRepo, cfg.JWTSecret)
	recorderService := services.NewRecorderService(db)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	projectHandler := handlers.NewProjectHandler(projectRepo, cameraRepo)
	cameraHandler := handlers.NewCameraHandler(cameraRepo, projectRepo)
	adminHandler := handlers.NewAdminHandler(userRepo, projectRepo, cameraRepo)
	recordingHandler := handlers.NewRecordingHandler(db, recorderService, cfg.RecordingsPath)

	// Initialize HLS cache (local disk for segments, avoids SMB reads)
	services.InitHLSCache(cfg.HLSCachePath, cfg.RecordingsPath)

	// Start HLS converter worker (scans every 2 minutes)
	hlsConverter := services.NewHLSConverter(cfg.RecordingsPath, db)
	hlsConverter.StartWorker(2 * time.Minute)

	// Routes
	r.Route("/api", func(r chi.Router) {
		// Health check (public)
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok"}`))
		})

		// Public video streaming routes (no auth required for video element)
		r.Get("/recordings/files/stream", recordingHandler.StreamVideoFile)
		r.Get("/recordings/files/download", recordingHandler.DownloadVideoFile)
		r.Get("/recordings/hls", recordingHandler.ServeHLSFile)
		r.Get("/recordings/restream", recordingHandler.LiveRestream)

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

			// Admin routes
			r.Route("/admin", func(r chi.Router) {
				r.Use(middleware.AdminOnlyMiddleware)

				r.Get("/users", adminHandler.GetAllUsers)
				r.Get("/users/{id}", adminHandler.GetUser)
				r.Delete("/users/{id}", adminHandler.DeleteUser)
				r.Put("/users/{id}/role", adminHandler.UpdateUserRole)
				r.Get("/projects", adminHandler.GetAllProjects)
			})
		})
	})

	return r
}
