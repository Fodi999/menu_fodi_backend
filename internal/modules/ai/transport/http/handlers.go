package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/middleware"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/ai/dto"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/ai/prompts"
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

// AddMissingIngredients godoc
// @Summary Add missing ingredients from recipe to fridge
// @Description Add ingredientsMissing from AI recipe to user's fridge (finds ingredients in catalog by name)
// @Tags ai
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.AddMissingIngredientsRequest true "Missing ingredients from recipe"
// @Success 200 {object} map[string]interface{} "success: true, added: count, skipped: []string"
// @Failure 400 {object} httpx.ErrorResponse
// @Failure 401 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /api/ai/add-missing-ingredients [post]
func (h *AIHandlers) AddMissingIngredients(w http.ResponseWriter, r *http.Request) {
	// 1. Authenticate user
	userIDPtr := middleware.GetUserID(r)
	if userIDPtr == nil {
		logger.Error("user ID not found in context")
		httpx.Unauthorized(w, "unauthorized")
		return
	}
	userID := userIDPtr.String()

	// 2. Parse request
	var req dto.AddMissingIngredientsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("failed to decode add missing ingredients request", zap.Error(err))
		httpx.BadRequest(w, "invalid request body")
		return
	}

	if len(req.Ingredients) == 0 {
		httpx.BadRequest(w, "ingredients list cannot be empty")
		return
	}

	logger.Info("Adding missing ingredients to fridge",
		zap.String("user_id", userID),
		zap.Int("count", len(req.Ingredients)))

	// 3. Find ingredients in catalog and add to fridge
	var added int
	var skipped []string

	for _, ing := range req.Ingredients {
		// Find ingredient in catalog by name (try Polish, English, Russian)
		var catalogIngredient models.Ingredient
		err := h.db.Where("LOWER(name) = LOWER(?)", ing.Name).
			First(&catalogIngredient).Error

		if err != nil {
			// Ingredient not found in catalog - skip it
			logger.Warn("ingredient not found in catalog",
				zap.String("user_id", userID),
				zap.String("ingredient", ing.Name),
				zap.Error(err))
			skipped = append(skipped, ing.Name)
			continue
		}

		// Check if user already has this ingredient
		var existingItem models.UserFridgeItem
		err = h.db.Where("user_id = ? AND ingredient_id = ?", userID, catalogIngredient.ID).
			First(&existingItem).Error

		if err == nil {
			// User already has this ingredient - update quantity instead of creating duplicate
			existingItem.Quantity += ing.Quantity
			if err := h.db.Save(&existingItem).Error; err != nil {
				logger.Error("failed to update existing ingredient quantity",
					zap.String("user_id", userID),
					zap.String("ingredient", ing.Name),
					zap.Error(err))
				skipped = append(skipped, ing.Name)
				continue
			}
			logger.Info("updated existing ingredient quantity",
				zap.String("user_id", userID),
				zap.String("ingredient", ing.Name),
				zap.Float64("added_quantity", ing.Quantity),
				zap.Float64("new_total", existingItem.Quantity))
			added++
			continue
		}

		// Create new fridge item
		expiresAt := time.Now().AddDate(0, 0, 14) // Default: 14 days for pantry items
		newItem := &models.UserFridgeItem{
			ID:           uuid.New().String(),
			UserID:       userID,
			IngredientID: catalogIngredient.ID,
			Quantity:     ing.Quantity,
			Unit:         ing.Unit,
			ExpiresAt:    &expiresAt,
		}

		if err := h.db.Create(newItem).Error; err != nil {
			logger.Error("failed to create fridge item",
				zap.String("user_id", userID),
				zap.String("ingredient", ing.Name),
				zap.Error(err))
			skipped = append(skipped, ing.Name)
			continue
		}

		logger.Info("added missing ingredient to fridge",
			zap.String("user_id", userID),
			zap.String("ingredient", ing.Name),
			zap.Float64("quantity", ing.Quantity),
			zap.String("unit", ing.Unit))
		added++
	}

	// 4. Return result
	httpx.Success(w, map[string]interface{}{
		"success": true,
		"added":   added,
		"skipped": skipped,
		"message": fmt.Sprintf("Added %d ingredients to fridge", added),
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

		// Get price per unit from cache
		var pricePerUnit *float64
		currency := "PLN"
		if item.CurrentPricePerUnit != nil && *item.CurrentPricePerUnit > 0 {
			pricePerUnit = item.CurrentPricePerUnit
			if item.CurrentPriceCurrency != "" {
				currency = item.CurrentPriceCurrency
			}
		}

		aiItems = append(aiItems, dto.FridgeItemDTO{
			Name:         item.Ingredient.Name,
			Category:     item.Ingredient.Category,
			Quantity:     item.Quantity,
			Unit:         item.Unit,
			DaysLeft:     daysLeft,
			Status:       status,
			PricePerUnit: pricePerUnit,
			Currency:     currency,
		})
	}

	logger.Info("AI analyzing fridge",
		zap.String("user_id", userID),
		zap.String("goal", req.Goal),
		zap.Int("items_count", len(aiItems)))

	// Нормализуем язык
	language := prompts.NormalizeLanguage(req.Language)

	// 3️⃣ Если холодильник пустой - не зовём AI
	if len(aiItems) == 0 {
		logger.Info("empty fridge - returning default message",
			zap.String("user_id", userID),
			zap.String("goal", req.Goal),
			zap.String("language", language))
		
		emptyMessages := map[string]string{
			"pl": "Twoja lodówka jest pusta. Dodaj produkty, aby otrzymać rekomendacje AI!",
			"en": "Your fridge is empty. Add some products to get AI recommendations!",
			"ru": "Твой холодильник пуст. Добавь продукты, чтобы получить рекомендации AI!",
		}
		
		httpx.Success(w, map[string]string{
			"result": emptyMessages[language],
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
			zap.String("language", language),
			zap.Error(err))
		
		// Возвращаем fallback вместо 500
		errorMessages := map[string]string{
			"pl": "Przepraszamy, AI jest chwilowo niedostępne. Spróbuj ponownie później.",
			"en": "Sorry, AI is temporarily unavailable. Please try again later.",
			"ru": "Извините, AI временно недоступен. Попробуйте позже.",
		}
		
		httpx.Success(w, map[string]string{
			"result": errorMessages[language],
		})
		return
	}

	// 🛡️ КРИТИЧЕСКАЯ ЗАЩИТА: гарантируем непустой result
	if strings.TrimSpace(result) == "" {
		logger.Warn("AI returned empty result - using fallback",
			zap.String("user_id", userID),
			zap.String("goal", req.Goal),
			zap.String("language", language))
		
		fallbackMessages := map[string]string{
			"pl": "AI nie wygenerowało odpowiedzi. Spróbuj ponownie za chwilę lub wybierz inny cel.",
			"en": "AI did not generate a response. Please try again in a moment or choose a different goal.",
			"ru": "AI не сгенерировал ответ. Попробуйте снова через минуту или выберите другую цель.",
		}
		result = fallbackMessages[language]
	}

	httpx.Success(w, map[string]string{
		"result": result,
	})
}

// CreateRecipeFromFridge godoc
// @Summary Create restaurant-grade recipe from fridge products
// @Description Generate a professional recipe using only available fridge items, prioritizing expiry dates
// @Tags ai
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.CreateRecipeFromFridgeRequest true "Recipe request"
// @Success 200 {object} dto.CreateRecipeFromFridgeResponse
// @Failure 400 {object} httpx.ErrorResponse
// @Failure 401 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /api/ai/create-recipe-from-fridge [post]
func (h *AIHandlers) CreateRecipeFromFridge(w http.ResponseWriter, r *http.Request) {
	// 1. Get authenticated user ID
	userIDPtr := middleware.GetUserID(r)
	if userIDPtr == nil {
		httpx.Unauthorized(w, "authentication required")
		return
	}
	userID := userIDPtr.String()

	// 2. Parse request
	var req dto.CreateRecipeFromFridgeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("failed to decode request", zap.Error(err))
		httpx.BadRequest(w, "invalid request body")
		return
	}

	// Normalize language
	language := prompts.NormalizeLanguage(req.Language)

	// 3. Get fridge items from database
	var fridgeItems []models.UserFridgeItem
	if err := h.db.Preload("Ingredient").Where("user_id = ?", userID).
		Find(&fridgeItems).Error; err != nil {
		logger.Error("failed to fetch fridge items",
			zap.String("user_id", userID),
			zap.Error(err))
		httpx.InternalError(w, "failed to load fridge data")
		return
	}

	// DEBUG: Log price data from database
	logger.Info("Loaded fridge items with prices",
		zap.String("user_id", userID),
		zap.Int("total_items", len(fridgeItems)))
	for _, item := range fridgeItems {
		priceStr := "NULL"
		if item.CurrentPricePerUnit != nil {
			priceStr = fmt.Sprintf("%.4f %s", *item.CurrentPricePerUnit, item.CurrentPriceCurrency)
		}
		logger.Info("Fridge item price",
			zap.String("user_id", userID),
			zap.String("ingredient_id", item.IngredientID),
			zap.String("ingredient_name", func() string {
				if item.Ingredient != nil {
					return item.Ingredient.Name
				}
				return "unknown"
			}()),
			zap.Float64("quantity", item.Quantity),
			zap.String("unit", item.Unit),
			zap.String("current_price_per_unit", priceStr))
	}

	// 4. Convert to DTO format with expiry calculation
	aiItems := make([]dto.FridgeItemDTO, 0, len(fridgeItems))
	for _, item := range fridgeItems {
		if item.Ingredient == nil {
			continue
		}

		// Calculate days left until expiry
		var daysLeft *int
		var status string
		if item.ExpiresAt != nil {
			now := time.Now()
			diff := item.ExpiresAt.Sub(now)
			days := int(diff.Hours() / 24)
			daysLeft = &days

			// Determine status based on days left
			switch {
			case days < 0:
				status = "expired"
			case days <= 2:
				status = "critical"
			case days <= 5:
				status = "warning"
			default:
				status = "ok"
			}
		} else {
			status = "ok"
		}

		// Get unit or use default "szt."
		unit := item.Unit
		if unit == "" {
			unit = "szt."
		}

		// Get price per unit from current cache
		var pricePerUnit *float64
		currency := "PLN" // default currency
		if item.CurrentPricePerUnit != nil && *item.CurrentPricePerUnit > 0 {
			pricePerUnit = item.CurrentPricePerUnit
			if item.CurrentPriceCurrency != "" {
				currency = item.CurrentPriceCurrency
			}
			logger.Info("Price data found for item",
				zap.String("user_id", userID),
				zap.String("name", item.Ingredient.Name),
				zap.Float64("price_per_unit", *pricePerUnit),
				zap.String("currency", currency))
		} else {
			logger.Warn("No price data for item",
				zap.String("user_id", userID),
				zap.String("name", item.Ingredient.Name),
				zap.Any("current_price_per_unit", item.CurrentPricePerUnit))
		}

		aiItems = append(aiItems, dto.FridgeItemDTO{
			Name:         item.Ingredient.Name,
			Quantity:     item.Quantity,
			Unit:         unit,
			Status:       status,
			DaysLeft:     daysLeft,
			PricePerUnit: pricePerUnit,
			Currency:     currency,
		})
	}

	// 5. Call AI service to create recipe
	response, err := h.service.CreateRecipeFromFridge(userID, language, aiItems)
	if err != nil {
		logger.Error("AI recipe creation failed",
			zap.String("user_id", userID),
			zap.Int("items_count", len(aiItems)),
			zap.String("language", language),
			zap.Error(err))
		httpx.InternalError(w, "failed to generate recipe")
		return
	}

	// 6. Return response directly (it already has proper structure)
	// Response structure:
	// Success case: {"success": true, "data": {"recipe": {...}, "usedProducts": [...]}}
	// Error case: {"success": false, "data": {"message": "..."}}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	if !response.Success {
		// Error case: empty fridge, no valid products, AI error
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"data": map[string]interface{}{
				"message": response.Message,
			},
		})
		return
	}
	
	// Success case: recipe generated
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"recipe":       response.Recipe,
			"usedProducts": response.UsedProducts,
		},
	})
}
