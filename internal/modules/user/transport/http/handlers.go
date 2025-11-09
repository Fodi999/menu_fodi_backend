package http

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/middleware"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/user/dto"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/user/service"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/platform/httpx"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/platform/logger"
)

// UserHandlers contains user HTTP handlers
type UserHandlers struct {
	service service.UserService
}

// NewUserHandlers creates new user handlers
func NewUserHandlers(service service.UserService) *UserHandlers {
	return &UserHandlers{service: service}
}

// GetProfile godoc
// @Summary Get user profile
// @Description Get current user profile with stats
// @Tags user
// @Security BearerAuth
// @Produce json
// @Success 200 {object} dto.UserProfileResponse
// @Failure 401 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /api/user/profile [get]
func (h *UserHandlers) GetProfile(w http.ResponseWriter, r *http.Request) {
	userIDPtr := middleware.GetUserID(r)
	if userIDPtr == nil {
		logger.Error("user ID not found in context")
		httpx.Unauthorized(w, "unauthorized")
		return
	}
	userID := *userIDPtr

	profile, err := h.service.GetProfile(userID)
	if err != nil {
		logger.Error("failed to get profile", zap.Error(err), zap.String("user_id", userID.String()))
		httpx.InternalError(w, "failed to get profile")
		return
	}

	httpx.Success(w, profile)
}

// UpdateProfile godoc
// @Summary Update user profile
// @Description Update user profile information
// @Tags user
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.UpdateProfileRequest true "Update profile request"
// @Success 200 {object} httpx.MessageResponse
// @Failure 400 {object} httpx.ErrorResponse
// @Failure 401 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /api/user/profile [put]
func (h *UserHandlers) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userIDPtr := middleware.GetUserID(r)
	if userIDPtr == nil {
		logger.Error("user ID not found in context")
		httpx.Unauthorized(w, "unauthorized")
		return
	}
	userID := *userIDPtr

	var req dto.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("failed to decode request", zap.Error(err))
		httpx.BadRequest(w, "invalid request body")
		return
	}

	if err := h.service.UpdateProfile(userID, req); err != nil {
		if err == service.ErrInvalidUpdateData {
			httpx.BadRequest(w, err.Error())
			return
		}
		logger.Error("failed to update profile", zap.Error(err), zap.String("user_id", userID.String()))
		httpx.InternalError(w, "failed to update profile")
		return
	}

	httpx.Success(w, map[string]string{"message": "profile updated successfully"})
}

// GetProgress godoc
// @Summary Get user progress
// @Description Get user's course progress
// @Tags user
// @Security BearerAuth
// @Produce json
// @Success 200 {array} dto.UserProgressResponse
// @Failure 401 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /api/user/progress [get]
func (h *UserHandlers) GetProgress(w http.ResponseWriter, r *http.Request) {
	userIDPtr := middleware.GetUserID(r)
	if userIDPtr == nil {
		logger.Error("user ID not found in context")
		httpx.Unauthorized(w, "unauthorized")
		return
	}
	userID := *userIDPtr

	progress, err := h.service.GetUserProgress(userID)
	if err != nil {
		logger.Error("failed to get progress", zap.Error(err), zap.String("user_id", userID.String()))
		httpx.InternalError(w, "failed to get progress")
		return
	}

	httpx.Success(w, progress)
}

// GetDashboard godoc
// @Summary Get user dashboard
// @Description Get comprehensive user dashboard data
// @Tags user
// @Security BearerAuth
// @Produce json
// @Success 200 {object} dto.DashboardResponse
// @Failure 401 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /api/user/dashboard [get]
func (h *UserHandlers) GetDashboard(w http.ResponseWriter, r *http.Request) {
	userIDPtr := middleware.GetUserID(r)
	if userIDPtr == nil {
		logger.Error("user ID not found in context")
		httpx.Unauthorized(w, "unauthorized")
		return
	}
	userID := *userIDPtr

	dashboard, err := h.service.GetDashboard(userID)
	if err != nil {
		logger.Error("failed to get dashboard", zap.Error(err), zap.String("user_id", userID.String()))
		httpx.InternalError(w, "failed to get dashboard")
		return
	}

	httpx.Success(w, dashboard)
}

// GetAchievements godoc
// @Summary Get user achievements
// @Description Get all unlocked achievements for user
// @Tags user
// @Security BearerAuth
// @Produce json
// @Success 200 {array} dto.AchievementResponse
// @Failure 401 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /api/user/achievements [get]
func (h *UserHandlers) GetAchievements(w http.ResponseWriter, r *http.Request) {
	userIDPtr := middleware.GetUserID(r)
	if userIDPtr == nil {
		logger.Error("user ID not found in context")
		httpx.Unauthorized(w, "unauthorized")
		return
	}
	userID := *userIDPtr

	achievements, err := h.service.GetAchievements(userID)
	if err != nil {
		logger.Error("failed to get achievements", zap.Error(err), zap.String("user_id", userID.String()))
		httpx.InternalError(w, "failed to get achievements")
		return
	}

	httpx.Success(w, achievements)
}
