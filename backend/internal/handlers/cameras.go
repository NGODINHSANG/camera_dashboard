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
	cameraRepo  *repository.CameraRepository
	projectRepo *repository.ProjectRepository
}

func NewCameraHandler(cameraRepo *repository.CameraRepository, projectRepo *repository.ProjectRepository) *CameraHandler {
	return &CameraHandler{
		cameraRepo:  cameraRepo,
		projectRepo: projectRepo,
	}
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

	// Check project access
	_, err, status := h.checkProjectAccess(claims, projectID)
	if err != nil {
		if status == http.StatusNotFound {
			response.NotFound(w, "Project not found")
		} else if status == http.StatusForbidden {
			response.Forbidden(w, "Access denied")
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

	projectIDStr := chi.URLParam(r, "projectId")
	projectID, err := strconv.ParseInt(projectIDStr, 10, 64)
	if err != nil {
		response.ValidationError(w, "Invalid project ID")
		return
	}

	// Check project access
	_, err, status := h.checkProjectAccess(claims, projectID)
	if err != nil {
		if status == http.StatusNotFound {
			response.NotFound(w, "Project not found")
		} else if status == http.StatusForbidden {
			response.Forbidden(w, "Access denied")
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

	// Check project access
	_, err, status := h.checkProjectAccess(claims, projectID)
	if err != nil {
		if status == http.StatusNotFound {
			response.NotFound(w, "Project not found")
		} else if status == http.StatusForbidden {
			response.Forbidden(w, "Access denied")
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

	camera.Name = req.Name
	camera.Location = req.Location
	camera.StreamURL = req.StreamURL
	camera.IsRecording = req.IsRecording

	if err := h.cameraRepo.Update(camera); err != nil {
		response.InternalError(w, "Failed to update camera")
		return
	}

	response.JSON(w, http.StatusOK, camera)
}

func (h *CameraHandler) Delete(w http.ResponseWriter, r *http.Request) {
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

	cameraIDStr := chi.URLParam(r, "cameraId")
	cameraID, err := strconv.ParseInt(cameraIDStr, 10, 64)
	if err != nil {
		response.ValidationError(w, "Invalid camera ID")
		return
	}

	// Check project access
	_, err, status := h.checkProjectAccess(claims, projectID)
	if err != nil {
		if status == http.StatusNotFound {
			response.NotFound(w, "Project not found")
		} else if status == http.StatusForbidden {
			response.Forbidden(w, "Access denied")
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
