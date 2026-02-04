package router

import (
	"net/http"

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
	recordingHandler := handlers.NewRecordingHandler(db, recorderService)

	// Routes
	r.Route("/api", func(r chi.Router) {
		// Health check (public)
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok"}`))
		})
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
				r.Get("/camera/{cameraId}/status", recordingHandler.GetCameraRecordingStatus)
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
