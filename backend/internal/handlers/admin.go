package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"camera-dashboard-backend/internal/models"
	"camera-dashboard-backend/internal/repository"
	"camera-dashboard-backend/pkg/response"

	"github.com/go-chi/chi/v5"
)

type AdminHandler struct {
	userRepo    *repository.UserRepository
	projectRepo *repository.ProjectRepository
	cameraRepo  *repository.CameraRepository
}

func NewAdminHandler(userRepo *repository.UserRepository, projectRepo *repository.ProjectRepository, cameraRepo *repository.CameraRepository) *AdminHandler {
	return &AdminHandler{
		userRepo:    userRepo,
		projectRepo: projectRepo,
		cameraRepo:  cameraRepo,
	}
}

func (h *AdminHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.userRepo.GetAll()
	if err != nil {
		response.InternalError(w, "Failed to fetch users")
		return
	}

	response.JSON(w, http.StatusOK, users)
}

func (h *AdminHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.ValidationError(w, "Invalid user ID")
		return
	}

	user, err := h.userRepo.GetByID(id)
	if err != nil {
		if err == repository.ErrUserNotFound {
			response.NotFound(w, "User not found")
			return
		}
		response.InternalError(w, "Failed to fetch user")
		return
	}

	response.JSON(w, http.StatusOK, user.ToResponse())
}

func (h *AdminHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.ValidationError(w, "Invalid user ID")
		return
	}

	if err := h.userRepo.Delete(id); err != nil {
		if err == repository.ErrUserNotFound {
			response.NotFound(w, "User not found")
			return
		}
		response.InternalError(w, "Failed to delete user")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "User deleted successfully"})
}

type UpdateRoleRequest struct {
	Role string `json:"role"`
}

func (h *AdminHandler) UpdateUserRole(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.ValidationError(w, "Invalid user ID")
		return
	}

	var req UpdateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.ValidationError(w, "Invalid request body")
		return
	}

	if req.Role != "admin" && req.Role != "user" {
		response.ValidationError(w, "Role must be 'admin' or 'user'")
		return
	}

	if err := h.userRepo.UpdateRole(id, models.Role(req.Role)); err != nil {
		if err == repository.ErrUserNotFound {
			response.NotFound(w, "User not found")
			return
		}
		response.InternalError(w, "Failed to update user role")
		return
	}

	user, err := h.userRepo.GetByID(id)
	if err != nil {
		response.InternalError(w, "Failed to fetch user")
		return
	}

	response.JSON(w, http.StatusOK, user.ToResponse())
}

func (h *AdminHandler) GetAllProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := h.projectRepo.GetAll()
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
