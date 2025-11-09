package http

import (
	"encoding/json"
	"net/http"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/middleware"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/auth/dto"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/auth/service"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/platform/httpx"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/platform/logger"
	"go.uber.org/zap"
)

// AuthHandler handles HTTP requests for authentication
type AuthHandler struct {
	service *service.AuthService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{
		service: svc,
	}
}

// Register handles POST /api/auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.BadRequest(w, "Invalid request body")
		return
	}

	response, err := h.service.Register(req)
	if err != nil {
		logger.Error("Registration failed", zap.Error(err), zap.String("email", req.Email))

		if err == service.ErrUserExists {
			httpx.BadRequest(w, "User already exists")
			return
		}

		if err == service.ErrWeakPassword {
			httpx.BadRequest(w, "Password must be at least 6 characters")
			return
		}

		httpx.InternalError(w, "Registration failed")
		return
	}

	logger.Info("User registered successfully",
		zap.String("userId", response.User.ID),
		zap.String("email", response.User.Email))

	httpx.Created(w, response)
}

// Login handles POST /api/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.BadRequest(w, "Invalid request body")
		return
	}

	response, err := h.service.Login(req)
	if err != nil {
		logger.Error("Login failed", zap.Error(err), zap.String("email", req.Email))

		if err == service.ErrInvalidCredentials {
			httpx.Unauthorized(w, "Invalid credentials")
			return
		}

		httpx.InternalError(w, "Login failed")
		return
	}

	logger.Info("User logged in successfully",
		zap.String("userId", response.User.ID),
		zap.String("email", response.User.Email))

	httpx.Success(w, response)
}

// VerifyToken handles POST /api/auth/verify
func (h *AuthHandler) VerifyToken(w http.ResponseWriter, r *http.Request) {
	var req dto.VerifyTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.BadRequest(w, "Invalid request body")
		return
	}

	response, err := h.service.VerifyToken(req)
	if err != nil {
		httpx.InternalError(w, "Token verification failed")
		return
	}

	httpx.Success(w, response)
}

// GetCurrentUser handles GET /api/auth/me
func (h *AuthHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	userIDPtr := middleware.GetUserID(r)
	if userIDPtr == nil {
		httpx.Unauthorized(w, "Unauthorized")
		return
	}
	userID := *userIDPtr

	response, err := h.service.GetCurrentUser(userID)
	if err != nil {
		logger.Error("Failed to get current user", zap.Error(err), zap.String("userId", userID.String()))

		if err == service.ErrUserNotFound {
			httpx.NotFound(w, "User not found")
			return
		}

		httpx.InternalError(w, "Failed to get user")
		return
	}

	httpx.Success(w, response)
}
