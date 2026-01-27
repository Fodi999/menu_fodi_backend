package http

import (
	"fmt"
	"net/http"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/database"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/middleware"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/dmitrijfomin/menu-fodifood/backend/pkg/utils"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// RecipeCostResponse - DTO для ответа при расчете стоимости
type RecipeCostResponse struct {
	RecipeID        string  `json:"recipeId"`
	RecipeTitle     string  `json:"recipeTitle"`
	TotalCost       float64 `json:"totalCost"`       // PLN, общая стоимость на одну порцию
	CostPerServing  float64 `json:"costPerServing"`  // PLN, (alias для TotalCost для ясности)
	IngredientsCount int    `json:"ingredientsCount"` // Количество ингредиентов
	MissingPrice    bool    `json:"missingPrice"`     // true если некоторые ингредиенты без цены
	Details         []RecipeCostDetail `json:"details"`
}

// RecipeCostDetail детали расчета стоимости по ингредиентам
type RecipeCostDetail struct {
	IngredientName string  `json:"ingredientName"`
	Quantity       float64 `json:"quantity"`
	Unit           string  `json:"unit"`
	PricePerUnit   float64 `json:"pricePerUnit,omitempty"`
	ItemCost       float64 `json:"itemCost"`
	Optional       bool    `json:"optional"`
	Source         string  `json:"source"` // "fridge" или "default" или "missing"
}

// CalculateRecipeCost - GET /api/admin/recipes/{recipeId}/calculate-cost
// Рассчитывает себестоимость рецепта из холодильника админа БЕЗ создания блюда
func (h *AdminHandlers) CalculateRecipeCost(w http.ResponseWriter, r *http.Request) {
	// 🛡️ Защита от panic
	defer func() {
		if rec := recover(); rec != nil {
			fmt.Printf("🚨 PANIC in CalculateRecipeCost: %v\n", rec)
			utils.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		}
	}()

	// Получаем админа из контекста (обязательно авторизован)
	claims := middleware.GetUserFromContext(r)
	if claims == nil || claims.Subject == "" {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	adminID := claims.Subject
	fmt.Printf("🔑 Admin ID: %s\n", adminID)

	// Получаем recipeId из path параметра
	recipeIDStr := chi.URLParam(r, "recipeId")
	if recipeIDStr == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "recipeId is required")
		return
	}

	// Валидируем UUID
	recipeID, err := uuid.Parse(recipeIDStr)
	if err != nil {
		fmt.Printf("❌ Invalid recipeId format: %s\n", recipeIDStr)
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid recipeId format")
		return
	}

	fmt.Printf("🎯 CalculateRecipeCost: recipeId=%s for admin=%s\n", recipeID.String(), adminID)

	// Получаем рецепт из каталога
	var recipe models.RecipeCatalog
	if err := database.DB.
		Preload("Ingredients.Ingredient").
		Where("id = ?", recipeID).
		First(&recipe).Error; err != nil {
		fmt.Printf("❌ Recipe not found: %v\n", err)
		utils.RespondWithError(w, http.StatusNotFound, "Recipe not found")
		return
	}

	fmt.Printf("✅ Found recipe: %s (title=%s)\n", recipe.ID, recipe.Title)

	// Получаем все items холодильника админа для быстрого O(1) lookup
	var fridgeItems []models.UserFridgeItem
	if err := database.DB.
		Where("user_id = ?", adminID).
		Find(&fridgeItems).Error; err != nil {
		fmt.Printf("⚠️ Error loading admin fridge: %v\n", err)
	}

	// Создаем map для быстрого поиска по ингредиентам
	fridgeMap := make(map[string]*models.UserFridgeItem)
	for i := range fridgeItems {
		fridgeMap[fridgeItems[i].IngredientID] = &fridgeItems[i]
	}

	fmt.Printf("📦 Loaded %d fridge items for admin\n", len(fridgeItems))

	// Рассчитываем стоимость
	var totalCost float64
	missingPrice := false
	var details []RecipeCostDetail

	for _, catalogIng := range recipe.Ingredients {
		ingredient := catalogIng.Ingredient

		var pricePerUnit float64
		var source string

		// 1️⃣ СНАЧАЛА ищем цену в холодильнике админа
		if fridgeItem, exists := fridgeMap[ingredient.ID]; exists && fridgeItem.CurrentPricePerUnit != nil {
			pricePerUnit = *fridgeItem.CurrentPricePerUnit
			source = "fridge"
			fmt.Printf("✅ Using fridge price for %s: %.2f\n", ingredient.ID, pricePerUnit)
		} else {
			// 2️⃣ FALLBACK: используем цену из каталога (DefaultPricePerUnit)
			if ingredient.DefaultPricePerUnit != nil {
				pricePerUnit = *ingredient.DefaultPricePerUnit
				source = "default"
				fmt.Printf("⚠️ Using default price for %s: %.2f\n", ingredient.ID, pricePerUnit)
			} else {
				// 3️⃣ ПОСЛЕДНИЙ вариант: нет цены совсем
				fmt.Printf("❌ No price found for ingredient: %s\n", ingredient.ID)
				missingPrice = true
				source = "missing"
			}
		}

		// Рассчитываем стоимость для этого ингредиента
		itemCost := catalogIng.Quantity * pricePerUnit

		// Добавляем в общую стоимость только если это не опциональный ингредиент
		if !catalogIng.Optional {
			totalCost += itemCost
		}

		// Получаем локализованное название ингредиента
		ingredientName := ingredient.ID // Fallback
		if ingredient.NamePL != nil && *ingredient.NamePL != "" {
			ingredientName = *ingredient.NamePL
		} else if ingredient.NameEN != nil && *ingredient.NameEN != "" {
			ingredientName = *ingredient.NameEN
		} else if ingredient.NameRU != nil && *ingredient.NameRU != "" {
			ingredientName = *ingredient.NameRU
		}

		details = append(details, RecipeCostDetail{
			IngredientName: ingredientName,
			Quantity:       catalogIng.Quantity,
			Unit:           catalogIng.Unit,
			PricePerUnit:   pricePerUnit,
			ItemCost:       itemCost,
			Optional:       catalogIng.Optional,
			Source:         source,
		})
	}

	// Нормализуем до 2 знаков после запятой
	totalCost = normalizeFloat(totalCost, 2)

	response := RecipeCostResponse{
		RecipeID:         recipe.ID.String(),
		RecipeTitle:      recipe.Title,
		TotalCost:        totalCost,
		CostPerServing:   totalCost,
		IngredientsCount: len(recipe.Ingredients),
		MissingPrice:     missingPrice,
		Details:          details,
	}

	fmt.Printf("✅ Recipe cost calculated: %.2f PLN (from admin fridge)\n", totalCost)
	utils.RespondWithJSON(w, http.StatusOK, response)
}

// normalizeFloat - вспомогательная функция для нормализации чисел
func normalizeFloat(value float64, decimals int) float64 {
	multiplier := 1.0
	for i := 0; i < decimals; i++ {
		multiplier *= 10
	}
	return float64(int64(value*multiplier+0.5)) / multiplier
}
