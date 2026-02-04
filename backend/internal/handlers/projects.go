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

type ProjectHandler struct {
	projectRepo *repository.ProjectRepository
	cameraRepo  *repository.CameraRepository
}

func NewProjectHandler(projectRepo *repository.ProjectRepository, cameraRepo *repository.CameraRepository) *ProjectHandler {
	return &ProjectHandler{
		projectRepo: projectRepo,
		cameraRepo:  cameraRepo,
	}
}

func (h *ProjectHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		response.Unauthorized(w, "Authentication required")
		return
	}

	projects, err := h.projectRepo.GetByUserID(claims.UserID)
	if err != nil {
		response.InternalError(w, "Failed to fetch projects")
		return
	}

	// Load cameras for each project
	for i := range projects {
		cameras, err := h.cameraRepo.GetByProjectID(projects[i].ID)
		if err != nil {
			response.InternalError(w, "Failed to fetch cameras")
			return
		}
		projects[i].Cameras = cameras
	}

	response.JSON(w, http.StatusOK, projects)
}

func (h *ProjectHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		response.Unauthorized(w, "Authentication required")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.ValidationError(w, "Invalid project ID")
		return
	}

	project, err := h.projectRepo.GetByID(id)
	if err != nil {
		if err == repository.ErrProjectNotFound {
			response.NotFound(w, "Project not found")
			return
		}
		response.InternalError(w, "Failed to fetch project")
		return
	}

	// Check ownership
	if project.UserID != claims.UserID && claims.Role != "admin" {
		response.Forbidden(w, "Access denied")
		return
	}

	// Load cameras
	cameras, err := h.cameraRepo.GetByProjectID(project.ID)
	if err != nil {
		response.InternalError(w, "Failed to fetch cameras")
		return
	}
	project.Cameras = cameras

	response.JSON(w, http.StatusOK, project)
}

func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		response.Unauthorized(w, "Authentication required")
		return
	}

	var req models.CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.ValidationError(w, "Invalid request body")
		return
	}

	if req.Name == "" {
		response.ValidationError(w, "Project name is required")
		return
	}

	project := &models.Project{
		UserID: claims.UserID,
		Name:   req.Name,
	}

	if err := h.projectRepo.Create(project); err != nil {
		response.InternalError(w, "Failed to create project")
		return
	}

	response.JSON(w, http.StatusCreated, project)
}

func (h *ProjectHandler) Update(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		response.Unauthorized(w, "Authentication required")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.ValidationError(w, "Invalid project ID")
		return
	}

	// Check ownership
	project, err := h.projectRepo.GetByID(id)
	if err != nil {
		if err == repository.ErrProjectNotFound {
			response.NotFound(w, "Project not found")
			return
		}
		response.InternalError(w, "Failed to fetch project")
		return
	}

	if project.UserID != claims.UserID && claims.Role != "admin" {
		response.Forbidden(w, "Access denied")
		return
	}

	var req models.UpdateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.ValidationError(w, "Invalid request body")
		return
	}

	if req.Name == "" {
		response.ValidationError(w, "Project name is required")
		return
	}

	project.Name = req.Name
	if err := h.projectRepo.Update(project); err != nil {
		response.InternalError(w, "Failed to update project")
		return
	}

	response.JSON(w, http.StatusOK, project)
}

func (h *ProjectHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		response.Unauthorized(w, "Authentication required")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.ValidationError(w, "Invalid project ID")
		return
	}

	// Check ownership
	project, err := h.projectRepo.GetByID(id)
	if err != nil {
		if err == repository.ErrProjectNotFound {
			response.NotFound(w, "Project not found")
			return
		}
		response.InternalError(w, "Failed to fetch project")
		return
	}

	if project.UserID != claims.UserID && claims.Role != "admin" {
		response.Forbidden(w, "Access denied")
		return
	}

	if err := h.projectRepo.Delete(id); err != nil {
		response.InternalError(w, "Failed to delete project")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "Project deleted successfully"})
}
