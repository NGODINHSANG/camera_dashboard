package handlers

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"camera-dashboard-backend/internal/middleware"
	"camera-dashboard-backend/internal/models"
	"camera-dashboard-backend/internal/repository"
	"camera-dashboard-backend/pkg/response"
	"camera-dashboard-backend/internal/services"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

const (
	MaxUploadSize = 2 << 30 // 2 GB
)

// Allowed video extensions
var allowedExtensions = map[string]bool{
	".mp4": true,
	".avi": true,
	".mkv": true,
	".mov": true,
}

type UploadHandler struct {
	db             *sqlx.DB
	uploadRepo     *repository.UploadRepository
	cameraRepo     *repository.CameraRepository
	projectRepo    *repository.ProjectRepository
	permRepo       *repository.ProjectPermissionRepository
	videoProcessor *services.VideoProcessor
}

func NewUploadHandler(db *sqlx.DB, cameraRepo *repository.CameraRepository, projectRepo *repository.ProjectRepository, permRepo *repository.ProjectPermissionRepository, videoProcessor *services.VideoProcessor) *UploadHandler {
	return &UploadHandler{
		db:             db,
		uploadRepo:     repository.NewUploadRepository(db),
		cameraRepo:     cameraRepo,
		projectRepo:    projectRepo,
		permRepo:       permRepo,
		videoProcessor: videoProcessor,
	}
}

func (h *UploadHandler) isProjectAdmin(claims *middleware.Claims, projectID int64) bool {
	return claims.Role == "admin" || h.permRepo.HasPermission(claims.UserID, projectID)
}

// Upload handles video file upload
func (h *UploadHandler) Upload(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		response.Unauthorized(w, "Unauthorized")
		return
	}

	// Parse project and camera IDs from URL
	projectIDStr := chi.URLParam(r, "projectId")
	cameraIDStr := chi.URLParam(r, "cameraId")

	projectID, err := strconv.ParseInt(projectIDStr, 10, 64)
	if err != nil {
		response.ValidationError(w, "Invalid project ID")
		return
	}

	cameraID, err := strconv.ParseInt(cameraIDStr, 10, 64)
	if err != nil {
		response.ValidationError(w, "Invalid camera ID")
		return
	}

	// Check project admin permission
	if !h.isProjectAdmin(claims, projectID) {
		response.Forbidden(w, "Only admin can upload videos")
		return
	}

	// Verify project access
	project, err := h.projectRepo.GetByID(projectID)
	if err != nil {
		response.NotFound(w, "Project not found")
		return
	}

	// Verify camera belongs to project
	camera, err := h.cameraRepo.GetByID(cameraID)
	if err != nil || camera.ProjectID != projectID {
		response.NotFound(w, "Camera not found in this project")
		return
	}

	// Limit request size
	r.Body = http.MaxBytesReader(w, r.Body, MaxUploadSize)

	// Parse multipart form
	if err := r.ParseMultipartForm(MaxUploadSize); err != nil {
		if strings.Contains(err.Error(), "too large") {
			response.ValidationError(w, "File too large. Maximum size is 2GB")
			return
		}
		response.ValidationError(w, "Failed to parse form: "+err.Error())
		return
	}

	// Get uploaded file
	file, header, err := r.FormFile("video")
	if err != nil {
		response.ValidationError(w, "No video file provided")
		return
	}
	defer file.Close()

	// Validate file extension
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !allowedExtensions[ext] {
		response.ValidationError(w, "Invalid file type. Allowed: mp4, avi, mkv, mov")
		return
	}

	// Generate unique filename
	jobID := uuid.New().String()
	tempFilename := jobID + ext

	// Save to temp directory
	tempPath := filepath.Join(h.videoProcessor.GetTempPath(), tempFilename)
	tempFile, err := os.Create(tempPath)
	if err != nil {
		log.Printf("[Upload] Failed to create temp file: %v", err)
		response.InternalError(w, "Failed to save file")
		return
	}
	defer tempFile.Close()

	// Copy file content
	written, err := io.Copy(tempFile, file)
	if err != nil {
		os.Remove(tempPath)
		log.Printf("[Upload] Failed to write file: %v", err)
		response.InternalError(w, "Failed to save file")
		return
	}

	log.Printf("[Upload] Saved file: %s (%d bytes)", tempPath, written)

	// Create upload job
	job := &models.UploadJob{
		ID:               jobID,
		CameraID:         cameraID,
		ProjectID:        projectID,
		Filename:         tempFilename,
		OriginalFilename: header.Filename,
		Status:           models.UploadStatusUploading,
		Progress:         0,
		FileSize:         written,
	}

	if err := h.uploadRepo.Create(job); err != nil {
		os.Remove(tempPath)
		log.Printf("[Upload] Failed to create job: %v", err)
		response.InternalError(w, "Failed to create upload job")
		return
	}

	// Start processing in background
	go func() {
		if err := h.videoProcessor.ProcessVideo(job, project.Name, camera.Name); err != nil {
			log.Printf("[Upload] Processing failed for job %s: %v", jobID, err)
		}
	}()

	// Return job ID for status polling
	response.JSON(w, http.StatusAccepted, map[string]interface{}{
		"job_id":   jobID,
		"filename": header.Filename,
		"size":     written,
		"status":   "uploading",
		"message":  "File uploaded, processing started",
	})
}

// GetJobStatus returns the status of an upload job
func (h *UploadHandler) GetJobStatus(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		response.Unauthorized(w, "Unauthorized")
		return
	}

	jobID := chi.URLParam(r, "jobId")
	if jobID == "" {
		response.ValidationError(w, "Job ID required")
		return
	}

	job, err := h.uploadRepo.GetByID(jobID)
	if err != nil {
		response.NotFound(w, "Job not found")
		return
	}

	response.JSON(w, http.StatusOK, job)
}

// CancelJob cancels an upload job
func (h *UploadHandler) CancelJob(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		response.Unauthorized(w, "Unauthorized")
		return
	}

	jobID := chi.URLParam(r, "jobId")
	if jobID == "" {
		response.ValidationError(w, "Job ID required")
		return
	}

	job, err := h.uploadRepo.GetByID(jobID)
	if err != nil {
		response.NotFound(w, "Job not found")
		return
	}

	// Can only cancel uploading or processing jobs
	if job.Status != models.UploadStatusUploading && job.Status != models.UploadStatusProcessing {
		response.ValidationError(w, "Cannot cancel completed or failed job")
		return
	}

	// Remove temp file
	tempPath := filepath.Join(h.videoProcessor.GetTempPath(), job.Filename)
	os.Remove(tempPath)

	// Delete job
	if err := h.uploadRepo.Delete(jobID); err != nil {
		log.Printf("[Upload] Failed to delete job: %v", err)
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "Job cancelled",
	})
}

// GetCameraJobs returns all upload jobs for a camera
func (h *UploadHandler) GetCameraJobs(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		response.Unauthorized(w, "Unauthorized")
		return
	}

	cameraIDStr := chi.URLParam(r, "cameraId")
	cameraID, err := strconv.ParseInt(cameraIDStr, 10, 64)
	if err != nil {
		response.ValidationError(w, "Invalid camera ID")
		return
	}

	jobs, err := h.uploadRepo.GetByCameraID(cameraID)
	if err != nil {
		log.Printf("[Upload] Failed to get jobs: %v", err)
		response.InternalError(w, "Failed to get jobs")
		return
	}

	response.JSON(w, http.StatusOK, jobs)
}
