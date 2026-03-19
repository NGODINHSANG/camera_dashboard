package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"camera-dashboard-backend/internal/middleware"
	"camera-dashboard-backend/internal/models"
	"camera-dashboard-backend/internal/repository"
	"camera-dashboard-backend/pkg/response"

	"github.com/go-chi/chi/v5"
)

type CameraHandler struct {
	cameraRepo          *repository.CameraRepository
	projectRepo         *repository.ProjectRepository
	onAutoRecordEnable  func(cameraID int64) // Callback when auto_record is enabled
	onAutoRecordDisable func(cameraID int64) // Callback when auto_record is disabled
}

func NewCameraHandler(cameraRepo *repository.CameraRepository, projectRepo *repository.ProjectRepository) *CameraHandler {
	return &CameraHandler{
		cameraRepo:  cameraRepo,
		projectRepo: projectRepo,
	}
}

// SetAutoRecordCallback sets the callback function called when auto_record is enabled
func (h *CameraHandler) SetAutoRecordCallback(callback func(cameraID int64)) {
	h.onAutoRecordEnable = callback
}

// SetAutoRecordDisableCallback sets the callback function called when auto_record is disabled
func (h *CameraHandler) SetAutoRecordDisableCallback(callback func(cameraID int64)) {
	h.onAutoRecordDisable = callback
}

func (h *CameraHandler) checkProjectAccess(claims *middleware.Claims, projectID int64) (*models.Project, error, int) {
	project, err := h.projectRepo.GetByID(projectID)
	if err != nil {
		if err == repository.ErrProjectNotFound {
			return nil, err, http.StatusNotFound
		}
		return nil, err, http.StatusInternalServerError
	}

	if project.UserID != claims.UserID && claims.Role != "admin" {
		return nil, repository.ErrProjectNotFound, http.StatusForbidden
	}

	return project, nil, 0
}

func (h *CameraHandler) GetByProject(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		response.Unauthorized(w, "Authentication required")
		return
	}

	projectIDStr := chi.URLParam(r, "projectId")
	projectID, err := strconv.ParseInt(projectIDStr, 10, 64)
	if err != nil {
		response.ValidationError(w, "Invalid project ID")
		return
	}

	// All users can VIEW cameras (surveillance dashboard)
	// Edit/Delete permissions are checked in those handlers

	// Check project exists
	_, err = h.projectRepo.GetByID(projectID)
	if err != nil {
		if err == repository.ErrProjectNotFound {
			response.NotFound(w, "Project not found")
		} else {
			response.InternalError(w, "Failed to fetch project")
		}
		return
	}

	cameras, err := h.cameraRepo.GetByProjectID(projectID)
	if err != nil {
		response.InternalError(w, "Failed to fetch cameras")
		return
	}

	response.JSON(w, http.StatusOK, cameras)
}

func (h *CameraHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		response.Unauthorized(w, "Authentication required")
		return
	}

	// Only admin can create cameras
	if claims.Role != "admin" {
		response.Forbidden(w, "Only admin can create cameras")
		return
	}

	projectIDStr := chi.URLParam(r, "projectId")
	projectID, err := strconv.ParseInt(projectIDStr, 10, 64)
	if err != nil {
		response.ValidationError(w, "Invalid project ID")
		return
	}

	// Check project exists
	_, err = h.projectRepo.GetByID(projectID)
	if err != nil {
		if err == repository.ErrProjectNotFound {
			response.NotFound(w, "Project not found")
		} else {
			response.InternalError(w, "Failed to fetch project")
		}
		return
	}

	var req models.CreateCameraRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.ValidationError(w, "Invalid request body")
		return
	}

	if req.Name == "" {
		response.ValidationError(w, "Camera name is required")
		return
	}
	if req.StreamURL == "" {
		response.ValidationError(w, "Stream URL is required")
		return
	}

	camera := &models.Camera{
		ProjectID:   projectID,
		Name:        req.Name,
		Location:    req.Location,
		StreamURL:   req.StreamURL,
		IsRecording: req.IsRecording,
		AutoRecord:  req.AutoRecord,
	}

	if err := h.cameraRepo.Create(camera); err != nil {
		response.InternalError(w, "Failed to create camera")
		return
	}

	response.JSON(w, http.StatusCreated, camera)
}

func (h *CameraHandler) Update(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		response.Unauthorized(w, "Authentication required")
		return
	}

	// Only admin can update cameras
	if claims.Role != "admin" {
		response.Forbidden(w, "Only admin can update cameras")
		return
	}

	projectIDStr := chi.URLParam(r, "projectId")
	projectID, err := strconv.ParseInt(projectIDStr, 10, 64)
	if err != nil {
		response.ValidationError(w, "Invalid project ID")
		return
	}

	cameraIDStr := chi.URLParam(r, "cameraId")
	cameraID, err := strconv.ParseInt(cameraIDStr, 10, 64)
	if err != nil {
		response.ValidationError(w, "Invalid camera ID")
		return
	}

	// Check project exists
	_, err = h.projectRepo.GetByID(projectID)
	if err != nil {
		if err == repository.ErrProjectNotFound {
			response.NotFound(w, "Project not found")
		} else {
			response.InternalError(w, "Failed to fetch project")
		}
		return
	}

	// Check camera belongs to project
	belongs, err := h.cameraRepo.BelongsToProject(cameraID, projectID)
	if err != nil {
		response.InternalError(w, "Failed to verify camera")
		return
	}
	if !belongs {
		response.NotFound(w, "Camera not found in this project")
		return
	}

	var req models.UpdateCameraRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.ValidationError(w, "Invalid request body")
		return
	}

	camera, err := h.cameraRepo.GetByID(cameraID)
	if err != nil {
		response.NotFound(w, "Camera not found")
		return
	}

	// Check if auto_record is being toggled
	autoRecordJustEnabled := !camera.AutoRecord && req.AutoRecord
	autoRecordJustDisabled := camera.AutoRecord && !req.AutoRecord

	camera.Name = req.Name
	camera.Location = req.Location
	camera.StreamURL = req.StreamURL
	camera.IsRecording = req.IsRecording
	camera.AutoRecord = req.AutoRecord

	if err := h.cameraRepo.Update(camera); err != nil {
		response.InternalError(w, "Failed to update camera")
		return
	}

	// If auto_record was just enabled, trigger stream check
	if autoRecordJustEnabled && h.onAutoRecordEnable != nil {
		go h.onAutoRecordEnable(cameraID)
	}

	// If auto_record was just disabled, stop recording
	if autoRecordJustDisabled && h.onAutoRecordDisable != nil {
		go h.onAutoRecordDisable(cameraID)
	}

	response.JSON(w, http.StatusOK, camera)
}

func (h *CameraHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		response.Unauthorized(w, "Authentication required")
		return
	}

	// Only admin can delete cameras
	if claims.Role != "admin" {
		response.Forbidden(w, "Only admin can delete cameras")
		return
	}

	projectIDStr := chi.URLParam(r, "projectId")
	projectID, err := strconv.ParseInt(projectIDStr, 10, 64)
	if err != nil {
		response.ValidationError(w, "Invalid project ID")
		return
	}

	cameraIDStr := chi.URLParam(r, "cameraId")
	cameraID, err := strconv.ParseInt(cameraIDStr, 10, 64)
	if err != nil {
		response.ValidationError(w, "Invalid camera ID")
		return
	}

	// Check project exists
	_, err = h.projectRepo.GetByID(projectID)
	if err != nil {
		if err == repository.ErrProjectNotFound {
			response.NotFound(w, "Project not found")
		} else {
			response.InternalError(w, "Failed to fetch project")
		}
		return
	}

	// Check camera belongs to project
	belongs, err := h.cameraRepo.BelongsToProject(cameraID, projectID)
	if err != nil {
		response.InternalError(w, "Failed to verify camera")
		return
	}
	if !belongs {
		response.NotFound(w, "Camera not found in this project")
		return
	}

	if err := h.cameraRepo.Delete(cameraID); err != nil {
		response.InternalError(w, "Failed to delete camera")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "Camera deleted successfully"})
}
