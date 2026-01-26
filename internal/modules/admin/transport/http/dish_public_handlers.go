package http

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/admin/service"
	"github.com/dmitrijfomin/menu-fodifood/backend/pkg/utils"
	"github.com/go-chi/chi/v5"
)

// ===========================
// Public Dish Handlers (Marketplace)
// ===========================

// GetPublishedDishes - GET /api/marketplace/dishes
// Публичный endpoint для получения опубликованных блюд
// Авторизация опциональна (для персонализации, если нужно)
func (h *AdminHandlers) GetPublishedDishes(w http.ResponseWriter, r *http.Request) {
	// Парсим параметры
	params := service.GetDishesParams{
		Limit:  20, // default
		Offset: 0,
	}

	// Фильтруем только published блюда
	publishedStatus := "published"
	params.Status = &publishedStatus

	// Recipe ID filter (если нужно показать блюда для конкретного рецепта)
	if recipeID := r.URL.Query().Get("recipeId"); recipeID != "" {
		params.RecipeID = &recipeID
	}

	// Pagination
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			params.Limit = limit
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			params.Offset = offset
		}
	}

	// Получаем блюда
	dishes, _, err := h.service.GetDishes(params)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get dishes: %v", err))
		return
	}

	// Фильтруем только доступные блюда
	availableDishes := make([]interface{}, 0)
	for _, dish := range dishes {
		if dish.IsAvailable {
			// Формируем публичный DTO (без внутренних данных)
			availableDishes = append(availableDishes, map[string]interface{}{
				"id":           dish.ID,
				"title":        dish.Title,
				"description":  dish.Description,
				"imageUrl":     dish.ImageURL,
				"price":        dish.Price,
				"isAvailable":  dish.IsAvailable,
				"recipe": map[string]interface{}{
					"id":          dish.Recipe.ID,
					"category":    dish.Recipe.Category,
					"difficulty":  dish.Recipe.Difficulty,
					"timeMinutes": dish.Recipe.TimeMinutes,
				},
			})
		}
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data":  availableDishes,
		"total": len(availableDishes),
		"limit": params.Limit,
		"offset": params.Offset,
	})
}

// GetPublishedDishByID - GET /api/marketplace/dishes/{id}
// Публичный endpoint для получения блюда по ID
func (h *AdminHandlers) GetPublishedDishByID(w http.ResponseWriter, r *http.Request) {
	dishID := chi.URLParam(r, "id")
	if dishID == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Dish ID is required")
		return
	}

	dish, err := h.service.GetDishByID(dishID)
	if err != nil {
		if err.Error() == "dish not found" {
			utils.RespondWithError(w, http.StatusNotFound, "Dish not found")
			return
		}
		utils.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get dish: %v", err))
		return
	}

	// Проверка: возвращаем только published и available
	if dish.Status != models.DishStatusPublished {
		utils.RespondWithError(w, http.StatusNotFound, "Dish not found")
		return
	}

	if !dish.IsAvailable {
		utils.RespondWithError(w, http.StatusConflict, "Dish is not available")
		return
	}

	// Формируем публичный DTO
	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"id":           dish.ID,
			"title":        dish.Title,
			"description":  dish.Description,
			"imageUrl":     dish.ImageURL,
			"price":        dish.Price,
			"isAvailable":  dish.IsAvailable,
			"recipe": map[string]interface{}{
				"id":          dish.Recipe.ID,
				"title":       dish.Recipe.Title,
				"category":    dish.Recipe.Category,
				"difficulty":  dish.Recipe.Difficulty,
				"timeMinutes": dish.Recipe.TimeMinutes,
				"servings":    dish.Recipe.Servings,
			},
		},
	})
}
