package services

import (
	"errors"

	"camera-dashboard-backend/internal/models"
	"camera-dashboard-backend/internal/repository"
	"camera-dashboard-backend/pkg/jwt"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailExists        = errors.New("email already exists")
)

type AuthService struct {
	userRepo  *repository.UserRepository
	jwtSecret string
}

func NewAuthService(userRepo *repository.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

type AuthResult struct {
	User      models.UserResponse `json:"user"`
	Token     string              `json:"token"`
	ExpiresAt string              `json:"expiresAt"`
}

func (s *AuthService) Register(req *models.UserRegisterRequest) (*AuthResult, error) {
	// Check if email exists
	exists, err := s.userRepo.EmailExists(req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrEmailExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return nil, err
	}

	// Check if this is the first user - make them admin
	role := models.RoleUser
	userCount, _ := s.userRepo.Count()
	if userCount == 0 {
		role = models.RoleAdmin
	}

	// Create user
	user := &models.User{
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Name:         req.Name,
		Role:         role,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	// Generate token
	token, expiresAt, err := jwt.GenerateToken(user.ID, user.Email, string(user.Role), s.jwtSecret)
	if err != nil {
		return nil, err
	}

	return &AuthResult{
		User:      user.ToResponse(),
		Token:     token,
		ExpiresAt: expiresAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

func (s *AuthService) Login(req *models.UserLoginRequest) (*AuthResult, error) {
	// Find user
	user, err := s.userRepo.GetByEmail(req.Email)
	if err != nil {
		if err == repository.ErrUserNotFound {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Generate token
	token, expiresAt, err := jwt.GenerateToken(user.ID, user.Email, string(user.Role), s.jwtSecret)
	if err != nil {
		return nil, err
	}

	return &AuthResult{
		User:      user.ToResponse(),
		Token:     token,
		ExpiresAt: expiresAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

func (s *AuthService) GetUserByID(id int64) (*models.UserResponse, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	response := user.ToResponse()
	return &response, nil
}
