package http

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/middleware"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/ai/dto"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/ai/service"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/platform/httpx"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/platform/logger"
)

// AIHandlers contains AI HTTP handlers
type AIHandlers struct {
	service service.AIService
	db      *gorm.DB
}

// NewAIHandlers creates new AI handlers
func NewAIHandlers(service service.AIService, db *gorm.DB) *AIHandlers {
	return &AIHandlers{
		service: service,
		db:      db,
	}
}

// ChefMentor godoc
// @Summary Chef Mentor interactive assistant
// @Description Interactive AI chef that helps create recipes step-by-step
// @Tags ai
// @Accept json
// @Produce json
// @Param request body dto.ChefMentorRequest true "Chef mentor request"
// @Success 200 {object} dto.ChefMentorResponse
// @Failure 400 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /api/ai/chef-mentor [post]
func (h *AIHandlers) ChefMentor(w http.ResponseWriter, r *http.Request) {
	var req dto.ChefMentorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("failed to decode request", zap.Error(err))
		httpx.BadRequest(w, "invalid request body")
		return
	}

	response, err := h.service.ChefMentor(req)
	if err != nil {
		switch err {
		case service.ErrEmptyMessage:
			httpx.BadRequest(w, err.Error())
		default:
			logger.Error("chef mentor error", zap.Error(err))
			httpx.InternalError(w, "failed to process request")
		}
		return
	}

	httpx.Success(w, response)
}

// GenerateMealPlan godoc
// @Summary Generate meal plan
// @Description Generate AI-powered meal plan for specified days and calories
// @Tags ai
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.MealPlanRequest true "Meal plan request"
// @Success 200 {object} dto.MealPlanResponse
// @Failure 400 {object} httpx.ErrorResponse
// @Failure 401 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /api/ai/meal-plan [post]
func (h *AIHandlers) GenerateMealPlan(w http.ResponseWriter, r *http.Request) {
	var req dto.MealPlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("failed to decode request", zap.Error(err))
		httpx.BadRequest(w, "invalid request body")
		return
	}

	// Get user ID (optional for this endpoint)
	userIDPtr := middleware.GetUserID(r)

	// Get fridge items if requested and user is authenticated
	var fridgeItems []models.UserFridge
	if req.UseFridge && userIDPtr != nil {
		if err := h.db.Where("user_id = ? AND available = ?", *userIDPtr, true).
			Find(&fridgeItems).Error; err != nil {
			logger.Error("failed to get fridge items", zap.Error(err))
		}
	}

	response, err := h.service.GenerateMealPlan(req, userIDPtr, fridgeItems)
	if err != nil {
		switch err {
		case service.ErrInvalidDays, service.ErrInvalidCalories:
			httpx.BadRequest(w, err.Error())
		default:
			logger.Error("meal plan generation error", zap.Error(err))
			httpx.InternalError(w, "failed to generate meal plan")
		}
		return
	}

	httpx.Success(w, response)
}

// GenerateRecipe godoc
// @Summary Generate recipe
// @Description Generate complete recipe from dish title using AI
// @Tags ai
// @Accept json
// @Produce json
// @Param request body dto.RecipeGenerationRequest true "Recipe generation request"
// @Success 200 {object} dto.GeneratedRecipe
// @Failure 400 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /api/ai/recipe-generator [post]
func (h *AIHandlers) GenerateRecipe(w http.ResponseWriter, r *http.Request) {
	var req dto.RecipeGenerationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("failed to decode request", zap.Error(err))
		httpx.BadRequest(w, "invalid request body")
		return
	}

	response, err := h.service.GenerateRecipe(req)
	if err != nil {
		switch err {
		case service.ErrEmptyTitle:
			httpx.BadRequest(w, err.Error())
		default:
			logger.Error("recipe generation error", zap.Error(err))
			httpx.InternalError(w, "failed to generate recipe")
		}
		return
	}

	httpx.Success(w, response)
}

// GetFridgeRecommendations godoc
// @Summary Get fridge-based recommendations
// @Description Get recipe recommendations based on available fridge items
// @Tags ai
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.FridgeRecommendationsRequest true "Recommendations request"
// @Success 200 {array} dto.FridgeRecommendation
// @Failure 401 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /api/ai/fridge-recommendations [post]
func (h *AIHandlers) GetFridgeRecommendations(w http.ResponseWriter, r *http.Request) {
	userIDPtr := middleware.GetUserID(r)
	if userIDPtr == nil {
		logger.Error("user ID not found in context")
		httpx.Unauthorized(w, "unauthorized")
		return
	}
	userID := *userIDPtr

	var req dto.FridgeRecommendationsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("failed to decode request", zap.Error(err))
		httpx.BadRequest(w, "invalid request body")
		return
	}

	// Get user's fridge items
	var fridgeItems []models.UserFridge
	if err := h.db.Where("user_id = ? AND available = ?", userID, true).
		Find(&fridgeItems).Error; err != nil {
		logger.Error("failed to get fridge items", zap.Error(err), zap.String("user_id", userID.String()))
		httpx.InternalError(w, "failed to get fridge items")
		return
	}

	recommendations, err := h.service.GetFridgeRecommendations(req, fridgeItems)
	if err != nil {
		logger.Error("fridge recommendations error", zap.Error(err), zap.String("user_id", userID.String()))
		httpx.InternalError(w, "failed to get recommendations")
		return
	}

	httpx.Success(w, map[string]interface{}{
		"success":         true,
		"recommendations": recommendations,
		"count":           len(recommendations),
	})
}

// SaveRecipeIngredientsToFridge godoc
// @Summary Save recipe ingredients to fridge
// @Description Add all ingredients from a recipe to user's fridge
// @Tags ai
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.SaveIngredientsRequest true "Save ingredients request"
// @Success 200 {object} httpx.MessageResponse
// @Failure 400 {object} httpx.ErrorResponse
// @Failure 401 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /api/ai/save-ingredients [post]
func (h *AIHandlers) SaveRecipeIngredientsToFridge(w http.ResponseWriter, r *http.Request) {
	userIDPtr := middleware.GetUserID(r)
	if userIDPtr == nil {
		logger.Error("user ID not found in context")
		httpx.Unauthorized(w, "unauthorized")
		return
	}
	userID := *userIDPtr

	var req dto.SaveIngredientsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("failed to decode request", zap.Error(err))
		httpx.BadRequest(w, "invalid request body")
		return
	}

	if len(req.Ingredients) == 0 {
		httpx.BadRequest(w, "ingredients list cannot be empty")
		return
	}

	// Save each ingredient to fridge
	for _, ingredient := range req.Ingredients {
		fridgeItem := &models.UserFridge{
			UserID:    userID,
			Product:   ingredient.Name,
			Quantity:  ingredient.Amount,
			Unit:      ingredient.Unit,
			Available: true,
		}

		if err := h.db.Create(fridgeItem).Error; err != nil {
			logger.Error("failed to save ingredient to fridge", 
				zap.Error(err), 
				zap.String("user_id", userID.String()),
				zap.String("product", ingredient.Name))
			httpx.InternalError(w, "failed to save ingredients")
			return
		}
	}

	httpx.Success(w, map[string]interface{}{
		"success":  true,
		"message":  "ingredients saved to fridge",
		"count":    len(req.Ingredients),
	})
}
