package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
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
	var fridgeItemsDTO []dto.AvailableIngredientDTO
	if req.UseFridge && userIDPtr != nil {
		var fridgeItems []models.UserFridgeItem
		if err := h.db.Preload("Ingredient").Where("user_id = ?", userIDPtr.String()).
			Find(&fridgeItems).Error; err != nil {
			logger.Error("failed to get fridge items", zap.Error(err))
		}

		// Конвертируем модели в DTO для AI слоя
		fridgeItemsDTO = make([]dto.AvailableIngredientDTO, len(fridgeItems))
		for i, item := range fridgeItems {
			name := ""
			if item.Ingredient != nil {
				name = item.Ingredient.Name
			}
			quantityStr := fmt.Sprintf("%.0f %s", item.Quantity, item.Unit)
			fridgeItemsDTO[i] = dto.NewAvailableIngredientDTO(name, quantityStr, item.ExpiresAt)
		}
	}

	response, err := h.service.GenerateMealPlan(req, userIDPtr, fridgeItemsDTO)
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
	var fridgeItems []models.UserFridgeItem
	if err := h.db.Preload("Ingredient").Where("user_id = ?", userID.String()).
		Find(&fridgeItems).Error; err != nil {
		logger.Error("failed to get fridge items", zap.Error(err), zap.String("user_id", userID.String()))
		httpx.InternalError(w, "failed to get fridge items")
		return
	}

	// Конвертируем модели в DTO для AI слоя
	fridgeItemsDTO := make([]dto.AvailableIngredientDTO, len(fridgeItems))
	for i, item := range fridgeItems {
		name := ""
		if item.Ingredient != nil {
			name = item.Ingredient.Name
		}
		quantityStr := fmt.Sprintf("%.0f %s", item.Quantity, item.Unit)
		fridgeItemsDTO[i] = dto.NewAvailableIngredientDTO(name, quantityStr, item.ExpiresAt)
	}

	recommendations, err := h.service.GetFridgeRecommendations(req, fridgeItemsDTO)
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
		expiresAt := time.Now().AddDate(0, 0, 7) // По умолчанию 7 дней
		fridgeItem := &models.UserFridgeItem{
			ID:           uuid.New().String(),
			UserID:       userID.String(),
			IngredientID: ingredient.Name, // TODO: нужно найти ID по имени из каталога
			Quantity:     ingredient.Amount,
			Unit:         ingredient.Unit,
			ExpiresAt:    &expiresAt,
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
		"success": true,
		"message": "ingredients saved to fridge",
		"count":   len(req.Ingredients),
	})
}

// AnalyzeFridge godoc
// @Summary AI Fridge Analysis (Smart Kitchen)
// @Description Analyze user's fridge and provide AI-powered recommendations based on goal
// @Tags ai
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.FridgeAnalyzeRequest true "Fridge analysis request"
// @Success 200 {object} map[string]string "result: AI recommendations"
// @Failure 400 {object} httpx.ErrorResponse
// @Failure 401 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /api/ai/fridge/analyze [post]
func (h *AIHandlers) AnalyzeFridge(w http.ResponseWriter, r *http.Request) {
	// Получаем User ID
	userIDPtr := middleware.GetUserID(r)
	if userIDPtr == nil {
		logger.Error("user ID not found in context")
		httpx.Unauthorized(w, "unauthorized")
		return
	}
	userID := userIDPtr.String()

	// Парсим запрос
	var req dto.FridgeAnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("failed to decode AI fridge analyze request", zap.Error(err))
		httpx.BadRequest(w, "invalid request body")
		return
	}

	// Валидация goal
	validGoals := map[string]bool{
		"today_meals":   true,
		"3_days_plan":   true,
		"reduce_waste":  true,
		"budget_review": true,
	}
	if !validGoals[req.Goal] {
		logger.Error("invalid goal received",
			zap.String("user_id", userID),
			zap.String("goal", req.Goal))
		httpx.BadRequest(w, "invalid goal: must be today_meals, 3_days_plan, reduce_waste, or budget_review")
		return
	}

	logger.Info("AI fridge analyze request",
		zap.String("user_id", userID),
		zap.String("goal", req.Goal),
		zap.String("language", req.Language),
		zap.String("time_preference", req.Preferences.Time),
		zap.String("budget_preference", req.Preferences.Budget))

	// Загружаем холодильник пользователя
	var fridgeItems []models.UserFridgeItem
	if err := h.db.Preload("Ingredient").Where("user_id = ?", userID).Find(&fridgeItems).Error; err != nil {
		logger.Error("failed to load fridge items",
			zap.String("user_id", userID),
			zap.Error(err))
		httpx.InternalError(w, "failed to load fridge")
		return
	}

	// Конвертируем в DTO для AI (без ID, user_id)
	aiItems := make([]dto.FridgeItemDTO, 0, len(fridgeItems))
	for _, item := range fridgeItems {
		if item.Ingredient == nil {
			continue
		}

		// Вычисляем daysLeft
		var daysLeft *int
		if item.ExpiresAt != nil {
			days := int(time.Until(*item.ExpiresAt).Hours() / 24)
			daysLeft = &days
		}

		// Определяем status
		status := "ok"
		if daysLeft != nil {
			if *daysLeft < 0 {
				status = "expired"
			} else if *daysLeft <= 2 {
				status = "critical"
			} else if *daysLeft <= 5 {
				status = "warning"
			}
		}

		aiItems = append(aiItems, dto.FridgeItemDTO{
			Name:       item.Ingredient.Name,
			Category:   item.Ingredient.Category,
			Quantity:   item.Quantity,
			Unit:       item.Unit,
			DaysLeft:   daysLeft,
			Status:     status,
			TotalPrice: item.CurrentPricePerUnit, // Используем цену если есть
			Currency:   item.CurrentPriceCurrency,
		})
	}

	logger.Info("AI analyzing fridge",
		zap.String("user_id", userID),
		zap.String("goal", req.Goal),
		zap.Int("items_count", len(aiItems)))

	// 3️⃣ Если холодильник пустой - не зовём AI
	if len(aiItems) == 0 {
		logger.Info("empty fridge - returning default message",
			zap.String("user_id", userID),
			zap.String("goal", req.Goal))
		
		httpx.Success(w, map[string]string{
			"result": "Twoja lodówka jest pusta. Dodaj produkty, aby otrzymać rekomendacje AI!",
		})
		return
	}

	// 2️⃣ Анализируем через AI service (с safe error handling)
	result, err := h.service.AnalyzeFridge(userID, req, aiItems)
	if err != nil {
		logger.Error("AI fridge analysis failed",
			zap.String("user_id", userID),
			zap.String("goal", req.Goal),
			zap.Int("items_count", len(aiItems)),
			zap.Error(err))
		
		// Возвращаем fallback вместо 500
		httpx.Success(w, map[string]string{
			"result": "Przepraszamy, AI jest chwilowo niedostępne. Spróbuj ponownie później.",
		})
		return
	}

	httpx.Success(w, map[string]string{
		"result": result,
	})
}
