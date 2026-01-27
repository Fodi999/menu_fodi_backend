package http

import (
	"fmt"
	"net/http"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/database"
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
	Details         []struct {
		IngredientName string  `json:"ingredientName"`
		Quantity       float64 `json:"quantity"`
		Unit           string  `json:"unit"`
		PricePerUnit   float64 `json:"pricePerUnit,omitempty"`
		ItemCost       float64 `json:"itemCost"`
		Optional       bool    `json:"optional"`
	} `json:"details"`
}

// CalculateRecipeCost - GET /api/admin/recipes/{recipeId}/cost
// Рассчитывает себестоимость рецепта БЕЗ создания блюда
func (h *AdminHandlers) CalculateRecipeCost(w http.ResponseWriter, r *http.Request) {
	// 🛡️ Защита от panic
	defer func() {
		if rec := recover(); rec != nil {
			fmt.Printf("🚨 PANIC in CalculateRecipeCost: %v\n", rec)
			utils.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		}
	}()

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

	fmt.Printf("🎯 CalculateRecipeCost: recipeId=%s\n", recipeID.String())

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

	// Рассчитываем стоимость
	var totalCost float64
	missingPrice := false
	var details []struct {
		IngredientName string  `json:"ingredientName"`
		Quantity       float64 `json:"quantity"`
		Unit           string  `json:"unit"`
		PricePerUnit   float64 `json:"pricePerUnit,omitempty"`
		ItemCost       float64 `json:"itemCost"`
		Optional       bool    `json:"optional"`
	}

	for _, catalogIng := range recipe.Ingredients {
		ingredient := catalogIng.Ingredient

		// Получаем цену за единицу из DefaultPricePerUnit
		var pricePerUnit float64
		if ingredient.DefaultPricePerUnit != nil {
			pricePerUnit = *ingredient.DefaultPricePerUnit
		} else {
			fmt.Printf("⚠️ No price for ingredient: %s\n", ingredient.ID)
			missingPrice = true
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

		details = append(details, struct {
			IngredientName string  `json:"ingredientName"`
			Quantity       float64 `json:"quantity"`
			Unit           string  `json:"unit"`
			PricePerUnit   float64 `json:"pricePerUnit,omitempty"`
			ItemCost       float64 `json:"itemCost"`
			Optional       bool    `json:"optional"`
		}{
			IngredientName: ingredientName,
			Quantity:       catalogIng.Quantity,
			Unit:           catalogIng.Unit,
			PricePerUnit:   pricePerUnit,
			ItemCost:       itemCost,
			Optional:       catalogIng.Optional,
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

	fmt.Printf("✅ Recipe cost calculated: %.2f PLN\n", totalCost)
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
