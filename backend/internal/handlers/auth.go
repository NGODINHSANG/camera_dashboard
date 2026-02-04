package handlers

import (
	"encoding/json"
	"net/http"

	"camera-dashboard-backend/internal/middleware"
	"camera-dashboard-backend/internal/models"
	"camera-dashboard-backend/internal/services"
	"camera-dashboard-backend/pkg/response"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.UserRegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.ValidationError(w, "Invalid request body")
		return
	}

	// Validation
	if req.Email == "" {
		response.ValidationError(w, "Email is required")
		return
	}
	if req.Password == "" {
		response.ValidationError(w, "Password is required")
		return
	}
	if len(req.Password) < 6 {
		response.ValidationError(w, "Password must be at least 6 characters")
		return
	}
	if req.Name == "" {
		response.ValidationError(w, "Name is required")
		return
	}

	result, err := h.authService.Register(&req)
	if err != nil {
		if err == services.ErrEmailExists {
			response.ValidationError(w, "Email already exists")
			return
		}
		response.InternalError(w, "Failed to register user")
		return
	}

	response.JSON(w, http.StatusCreated, result)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.UserLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.ValidationError(w, "Invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		response.ValidationError(w, "Email and password are required")
		return
	}

	result, err := h.authService.Login(&req)
	if err != nil {
		if err == services.ErrInvalidCredentials {
			response.Unauthorized(w, "Invalid email or password")
			return
		}
		response.InternalError(w, "Failed to login")
		return
	}

	response.JSON(w, http.StatusOK, result)
}

func (h *AuthHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		response.Unauthorized(w, "Authentication required")
		return
	}

	user, err := h.authService.GetUserByID(claims.UserID)
	if err != nil {
		response.NotFound(w, "User not found")
		return
	}

	response.JSON(w, http.StatusOK, user)
}
