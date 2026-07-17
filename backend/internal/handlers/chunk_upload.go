package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"camera-dashboard-backend/internal/middleware"
	"camera-dashboard-backend/internal/models"
	"camera-dashboard-backend/internal/repository"
	"camera-dashboard-backend/internal/services"
	"camera-dashboard-backend/pkg/response"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

const (
	MaxChunkSize = 10 << 20 // 10 MB max per chunk
)

type ChunkUploadHandler struct {
	db             *sqlx.DB
	chunkRepo      *repository.ChunkRepository
	uploadRepo     *repository.UploadRepository
	cameraRepo     *repository.CameraRepository
	projectRepo    *repository.ProjectRepository
	permRepo       *repository.ProjectPermissionRepository
	videoProcessor *services.VideoProcessor
}

func NewChunkUploadHandler(db *sqlx.DB, cameraRepo *repository.CameraRepository, projectRepo *repository.ProjectRepository, permRepo *repository.ProjectPermissionRepository, videoProcessor *services.VideoProcessor) *ChunkUploadHandler {
	return &ChunkUploadHandler{
		db:             db,
		chunkRepo:      repository.NewChunkRepository(db),
		uploadRepo:     repository.NewUploadRepository(db),
		cameraRepo:     cameraRepo,
		projectRepo:    projectRepo,
		permRepo:       permRepo,
		videoProcessor: videoProcessor,
	}
}

func (h *ChunkUploadHandler) isProjectAdmin(claims *middleware.Claims, projectID int64) bool {
	return claims.Role == "admin" || h.permRepo.HasPermission(claims.UserID, projectID)
}

// InitUpload initializes a chunked upload session
func (h *ChunkUploadHandler) InitUpload(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		response.Unauthorized(w, "Unauthorized")
		return
	}

	// Parse project and camera IDs
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

	// Verify project and camera
	_, err = h.projectRepo.GetByID(projectID)
	if err != nil {
		response.NotFound(w, "Project not found")
		return
	}

	camera, err := h.cameraRepo.GetByID(cameraID)
	if err != nil || camera.ProjectID != projectID {
		response.NotFound(w, "Camera not found in this project")
		return
	}

	// Parse multipart form
	if err := r.ParseMultipartForm(1 << 20); err != nil { // 1MB for form data
		log.Printf("[ChunkUpload] Failed to parse form: %v", err)
		response.ValidationError(w, "Failed to parse form data")
		return
	}

	// Parse request body
	filename := r.FormValue("filename")
	totalSizeStr := r.FormValue("totalSize")
	totalChunksStr := r.FormValue("totalChunks")
	chunkSizeStr := r.FormValue("chunkSize")

	log.Printf("[ChunkUpload] Init request: filename=%s, totalSize=%s, totalChunks=%s, chunkSize=%s",
		filename, totalSizeStr, totalChunksStr, chunkSizeStr)

	if filename == "" || totalSizeStr == "" || totalChunksStr == "" || chunkSizeStr == "" {
		response.ValidationError(w, "Missing required fields: filename, totalSize, totalChunks, chunkSize")
		return
	}

	totalSize, _ := strconv.ParseInt(totalSizeStr, 10, 64)
	totalChunks, _ := strconv.Atoi(totalChunksStr)
	chunkSize, _ := strconv.ParseInt(chunkSizeStr, 10, 64)

	// Validate file extension
	ext := strings.ToLower(filepath.Ext(filename))
	if !allowedExtensions[ext] {
		response.ValidationError(w, "Invalid file type. Allowed: mp4, avi, mkv, mov")
		return
	}

	// Validate size (max 2GB)
	if totalSize > MaxUploadSize {
		response.ValidationError(w, "File too large. Maximum size is 2GB")
		return
	}

	// Generate upload ID
	uploadID := uuid.New().String()

	// Create chunks directory
	chunksDir := filepath.Join(h.videoProcessor.GetTempPath(), uploadID)
	if err := os.MkdirAll(chunksDir, 0755); err != nil {
		log.Printf("[ChunkUpload] Failed to create chunks dir: %v", err)
		response.InternalError(w, "Failed to initialize upload")
		return
	}

	// Create chunk upload session
	chunk := &models.ChunkUpload{
		ID:               uploadID,
		CameraID:         cameraID,
		ProjectID:        projectID,
		OriginalFilename: filename,
		TotalSize:        totalSize,
		TotalChunks:      totalChunks,
		UploadedChunks:   0,
		ChunkSize:        chunkSize,
		Status:           models.ChunkStatusUploading,
	}

	if err := h.chunkRepo.Create(chunk); err != nil {
		os.RemoveAll(chunksDir)
		log.Printf("[ChunkUpload] Failed to create session: %v", err)
		response.InternalError(w, "Failed to create upload session")
		return
	}

	log.Printf("[ChunkUpload] Initialized upload %s: %s (%d bytes, %d chunks)", uploadID, filename, totalSize, totalChunks)

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"uploadId":    uploadID,
		"totalChunks": totalChunks,
		"chunkSize":   chunkSize,
		"status":      "initialized",
	})
}

// UploadChunk handles uploading a single chunk
func (h *ChunkUploadHandler) UploadChunk(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		response.Unauthorized(w, "Unauthorized")
		return
	}

	// Limit chunk size
	r.Body = http.MaxBytesReader(w, r.Body, MaxChunkSize)

	uploadID := chi.URLParam(r, "uploadId")
	chunkIndexStr := r.FormValue("chunkIndex")

	if uploadID == "" || chunkIndexStr == "" {
		response.ValidationError(w, "Missing uploadId or chunkIndex")
		return
	}

	chunkIndex, err := strconv.Atoi(chunkIndexStr)
	if err != nil {
		response.ValidationError(w, "Invalid chunk index")
		return
	}

	// Get upload session
	session, err := h.chunkRepo.GetByID(uploadID)
	if err != nil {
		response.NotFound(w, "Upload session not found")
		return
	}

	if session.Status != models.ChunkStatusUploading {
		response.ValidationError(w, "Upload session is not in uploading state")
		return
	}

	// Get chunk file
	file, _, err := r.FormFile("chunk")
	if err != nil {
		response.ValidationError(w, "No chunk data provided")
		return
	}
	defer file.Close()

	// Save chunk to temp directory
	chunksDir := filepath.Join(h.videoProcessor.GetTempPath(), uploadID)
	chunkPath := filepath.Join(chunksDir, fmt.Sprintf("chunk_%05d", chunkIndex))

	chunkFile, err := os.Create(chunkPath)
	if err != nil {
		log.Printf("[ChunkUpload] Failed to create chunk file: %v", err)
		response.InternalError(w, "Failed to save chunk")
		return
	}
	defer chunkFile.Close()

	written, err := io.Copy(chunkFile, file)
	if err != nil {
		os.Remove(chunkPath)
		log.Printf("[ChunkUpload] Failed to write chunk: %v", err)
		response.InternalError(w, "Failed to save chunk")
		return
	}

	// Count uploaded chunks by counting files in directory (no DB needed, avoids SQLITE_BUSY)
	files, _ := os.ReadDir(chunksDir)
	uploadedCount := 0
	for _, f := range files {
		if strings.HasPrefix(f.Name(), "chunk_") {
			uploadedCount++
		}
	}

	log.Printf("[ChunkUpload] Chunk %d/%d uploaded for %s (%d bytes)", chunkIndex+1, session.TotalChunks, uploadID, written)

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"chunkIndex":     chunkIndex,
		"uploadedChunks": uploadedCount,
		"totalChunks":    session.TotalChunks,
		"status":         "uploaded",
	})
}

// CompleteUpload merges chunks and starts processing
func (h *ChunkUploadHandler) CompleteUpload(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		response.Unauthorized(w, "Unauthorized")
		return
	}

	uploadID := chi.URLParam(r, "uploadId")
	if uploadID == "" {
		response.ValidationError(w, "Missing uploadId")
		return
	}

	// Get upload session
	session, err := h.chunkRepo.GetByID(uploadID)
	if err != nil {
		response.NotFound(w, "Upload session not found")
		return
	}

	if session.Status != models.ChunkStatusUploading {
		response.ValidationError(w, "Upload session is not in uploading state")
		return
	}

	// Count uploaded chunks from filesystem
	chunksDir := filepath.Join(h.videoProcessor.GetTempPath(), uploadID)
	files, err := os.ReadDir(chunksDir)
	if err != nil {
		response.ValidationError(w, "Failed to read chunks directory")
		return
	}

	uploadedChunks := 0
	for _, f := range files {
		if strings.HasPrefix(f.Name(), "chunk_") {
			uploadedChunks++
		}
	}

	// Verify all chunks uploaded
	if uploadedChunks < session.TotalChunks {
		response.ValidationError(w, fmt.Sprintf("Missing chunks: %d/%d uploaded", uploadedChunks, session.TotalChunks))
		return
	}

	// Get project and camera info
	project, err := h.projectRepo.GetByID(session.ProjectID)
	if err != nil {
		response.NotFound(w, "Project not found")
		return
	}

	camera, err := h.cameraRepo.GetByID(session.CameraID)
	if err != nil {
		response.NotFound(w, "Camera not found")
		return
	}

	// Update status to merging
	h.chunkRepo.UpdateStatus(uploadID, models.ChunkStatusMerging)

	// Start merge and process in background
	go func() {
		if err := h.mergeAndProcess(session, project.Name, camera.Name); err != nil {
			log.Printf("[ChunkUpload] Merge/process failed for %s: %v", uploadID, err)
			h.chunkRepo.UpdateError(uploadID, err.Error())
		}
	}()

	response.JSON(w, http.StatusAccepted, map[string]interface{}{
		"uploadId": uploadID,
		"status":   "merging",
		"message":  "Chunks uploaded, merging and processing started",
	})
}

// mergeAndProcess merges chunks and processes the video
func (h *ChunkUploadHandler) mergeAndProcess(session *models.ChunkUpload, projectName, cameraName string) error {
	log.Printf("[ChunkUpload] Starting merge for %s", session.ID)

	chunksDir := filepath.Join(h.videoProcessor.GetTempPath(), session.ID)
	ext := strings.ToLower(filepath.Ext(session.OriginalFilename))
	mergedPath := filepath.Join(h.videoProcessor.GetTempPath(), session.ID+ext)

	// Get sorted list of chunks
	files, err := os.ReadDir(chunksDir)
	if err != nil {
		return fmt.Errorf("failed to read chunks directory: %v", err)
	}

	// Sort chunks by name (chunk_00000, chunk_00001, ...)
	var chunkFiles []string
	for _, f := range files {
		if strings.HasPrefix(f.Name(), "chunk_") {
			chunkFiles = append(chunkFiles, filepath.Join(chunksDir, f.Name()))
		}
	}
	sort.Strings(chunkFiles)

	if len(chunkFiles) != session.TotalChunks {
		return fmt.Errorf("chunk count mismatch: found %d, expected %d", len(chunkFiles), session.TotalChunks)
	}

	// Create merged file
	mergedFile, err := os.Create(mergedPath)
	if err != nil {
		return fmt.Errorf("failed to create merged file: %v", err)
	}

	// Merge all chunks
	var totalWritten int64
	for _, chunkPath := range chunkFiles {
		chunkFile, err := os.Open(chunkPath)
		if err != nil {
			mergedFile.Close()
			os.Remove(mergedPath)
			return fmt.Errorf("failed to open chunk %s: %v", chunkPath, err)
		}

		written, err := io.Copy(mergedFile, chunkFile)
		chunkFile.Close()
		if err != nil {
			mergedFile.Close()
			os.Remove(mergedPath)
			return fmt.Errorf("failed to copy chunk: %v", err)
		}
		totalWritten += written
	}
	mergedFile.Close()

	log.Printf("[ChunkUpload] Merged %d chunks into %s (%d bytes)", len(chunkFiles), mergedPath, totalWritten)

	// Clean up chunks directory
	os.RemoveAll(chunksDir)

	// Update status to processing
	h.chunkRepo.UpdateStatus(session.ID, models.ChunkStatusProcessing)

	// Create upload job for video processing
	job := &models.UploadJob{
		ID:               session.ID,
		CameraID:         session.CameraID,
		ProjectID:        session.ProjectID,
		Filename:         session.ID + ext,
		OriginalFilename: session.OriginalFilename,
		Status:           models.UploadStatusProcessing,
		Progress:         0,
		FileSize:         totalWritten,
	}

	if err := h.uploadRepo.Create(job); err != nil {
		log.Printf("[ChunkUpload] Warning: Failed to create upload job: %v", err)
	}

	// Process video
	if err := h.videoProcessor.ProcessVideo(job, projectName, cameraName); err != nil {
		h.chunkRepo.UpdateError(session.ID, err.Error())
		return err
	}

	// Update chunk session status
	h.chunkRepo.UpdateStatus(session.ID, models.ChunkStatusCompleted)

	log.Printf("[ChunkUpload] Completed processing for %s", session.ID)
	return nil
}

// GetUploadStatus returns the status of a chunked upload
func (h *ChunkUploadHandler) GetUploadStatus(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		response.Unauthorized(w, "Unauthorized")
		return
	}

	uploadID := chi.URLParam(r, "uploadId")
	if uploadID == "" {
		response.ValidationError(w, "Missing uploadId")
		return
	}

	session, err := h.chunkRepo.GetByID(uploadID)
	if err != nil {
		response.NotFound(w, "Upload session not found")
		return
	}

	// Count uploaded chunks from filesystem (more reliable than DB)
	uploadedChunks := 0
	if session.Status == models.ChunkStatusUploading {
		chunksDir := filepath.Join(h.videoProcessor.GetTempPath(), uploadID)
		files, _ := os.ReadDir(chunksDir)
		for _, f := range files {
			if strings.HasPrefix(f.Name(), "chunk_") {
				uploadedChunks++
			}
		}
	} else {
		uploadedChunks = session.TotalChunks // Already completed uploading
	}

	// Get processing progress if applicable
	var progress int
	if session.Status == models.ChunkStatusProcessing {
		job, err := h.uploadRepo.GetByID(uploadID)
		if err == nil {
			progress = job.Progress
		}
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"uploadId":       session.ID,
		"status":         session.Status,
		"uploadedChunks": uploadedChunks,
		"totalChunks":    session.TotalChunks,
		"progress":       progress,
		"error":          session.Error,
	})
}

// CancelUpload cancels a chunked upload
func (h *ChunkUploadHandler) CancelUpload(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		response.Unauthorized(w, "Unauthorized")
		return
	}

	uploadID := chi.URLParam(r, "uploadId")
	if uploadID == "" {
		response.ValidationError(w, "Missing uploadId")
		return
	}

	session, err := h.chunkRepo.GetByID(uploadID)
	if err != nil {
		response.NotFound(w, "Upload session not found")
		return
	}

	// Clean up chunks directory
	chunksDir := filepath.Join(h.videoProcessor.GetTempPath(), uploadID)
	os.RemoveAll(chunksDir)

	// Clean up merged file if exists
	ext := strings.ToLower(filepath.Ext(session.OriginalFilename))
	mergedPath := filepath.Join(h.videoProcessor.GetTempPath(), uploadID+ext)
	os.Remove(mergedPath)

	// Delete session
	h.chunkRepo.Delete(uploadID)

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "Upload cancelled",
	})
}
