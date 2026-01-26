package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/middleware"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/admin/service"
	"github.com/dmitrijfomin/menu-fodifood/backend/pkg/utils"
	"github.com/go-chi/chi/v5"
)

// ===========================
// Dish API Handlers
// ===========================

// GenerateDishFromRecipe - POST /api/admin/dishes/generate-from-recipe
// Генерирует карточку блюда из рецепта через AI
func (h *AdminHandlers) GenerateDishFromRecipe(w http.ResponseWriter, r *http.Request) {
	// Получаем admin ID из контекста
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	adminID := claims.Subject

	// Парсим запрос
	var req service.GenerateDishRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Валидация
	if req.RecipeID == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Recipe ID is required")
		return
	}
	if req.TargetMargin < 0 || req.TargetMargin > 100 {
		utils.RespondWithError(w, http.StatusBadRequest, "Target margin must be between 0 and 100")
		return
	}

	// Генерируем блюдо
	dish, err := h.service.GenerateDishWithAI(req, adminID)
	if err != nil {
		if err.Error() == "recipe not found" {
			utils.RespondWithError(w, http.StatusNotFound, "Recipe not found")
			return
		}
		if err.Error() == "cannot create dish: missing ingredients" {
			utils.RespondWithError(w, http.StatusConflict, "Cannot create dish: some ingredients are missing")
			return
		}
		utils.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to generate dish: %v", err))
		return
	}

	utils.RespondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "Dish generated successfully",
		"dish":    dish,
	})
}

// GetDishes - GET /api/admin/dishes
// Возвращает список блюд с фильтрацией
func (h *AdminHandlers) GetDishes(w http.ResponseWriter, r *http.Request) {
	// Парсим параметры
	params := service.GetDishesParams{
		Limit:  20, // default
		Offset: 0,
	}

	// Status filter
	if status := r.URL.Query().Get("status"); status != "" {
		params.Status = &status
	}

	// Recipe ID filter
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
	dishes, total, err := h.service.GetDishes(params)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get dishes: %v", err))
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data":  dishes,
		"total": total,
		"limit": params.Limit,
		"offset": params.Offset,
	})
}

// GetDishByID - GET /api/admin/dishes/{id}
// Возвращает блюдо по ID
func (h *AdminHandlers) GetDishByID(w http.ResponseWriter, r *http.Request) {
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

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": dish,
	})
}

// UpdateDish - PATCH /api/admin/dishes/{id}
// Обновляет блюдо (только draft и approved)
func (h *AdminHandlers) UpdateDish(w http.ResponseWriter, r *http.Request) {
	// Получаем admin ID из контекста
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	adminID := claims.Subject

	dishID := chi.URLParam(r, "id")
	if dishID == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Dish ID is required")
		return
	}

	// Парсим запрос
	var req service.UpdateDishRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Обновляем блюдо
	dish, err := h.service.UpdateDish(dishID, req, adminID)
	if err != nil {
		if err.Error() == "dish not found" {
			utils.RespondWithError(w, http.StatusNotFound, "Dish not found")
			return
		}
		if err.Error() == "cannot edit dish with status: published (only draft and approved can be edited)" {
			utils.RespondWithError(w, http.StatusForbidden, "Cannot edit published dish")
			return
		}
		utils.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to update dish: %v", err))
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Dish updated successfully",
		"dish":    dish,
	})
}

// ApproveDish - POST /api/admin/dishes/{id}/approve
// Утверждает блюдо (draft → approved)
func (h *AdminHandlers) ApproveDish(w http.ResponseWriter, r *http.Request) {
	// Получаем admin ID из контекста
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	adminID := claims.Subject

	dishID := chi.URLParam(r, "id")
	if dishID == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Dish ID is required")
		return
	}

	// Утверждаем блюдо
	err := h.service.ApproveDish(dishID, adminID)
	if err != nil {
		if err.Error() == "dish not found" {
			utils.RespondWithError(w, http.StatusNotFound, "Dish not found")
			return
		}
		if err.Error() == "only draft dishes can be approved" {
			utils.RespondWithError(w, http.StatusBadRequest, "Only draft dishes can be approved")
			return
		}
		utils.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to approve dish: %v", err))
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Dish approved successfully",
	})
}

// PublishDish - POST /api/admin/dishes/{id}/publish
// Публикует блюдо (approved → published)
func (h *AdminHandlers) PublishDish(w http.ResponseWriter, r *http.Request) {
	// Получаем admin ID из контекста
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	adminID := claims.Subject

	dishID := chi.URLParam(r, "id")
	if dishID == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Dish ID is required")
		return
	}

	// Публикуем блюдо
	err := h.service.PublishDish(dishID, adminID)
	if err != nil {
		if err.Error() == "dish not found" {
			utils.RespondWithError(w, http.StatusNotFound, "Dish not found")
			return
		}
		if err.Error() == "only approved dishes can be published" {
			utils.RespondWithError(w, http.StatusBadRequest, "Only approved dishes can be published")
			return
		}
		if err.Error() == "cannot publish: ingredients are not available" {
			utils.RespondWithError(w, http.StatusConflict, "Cannot publish: ingredients are not available")
			return
		}
		utils.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to publish dish: %v", err))
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Dish published successfully",
	})
}

// UnpublishDish - POST /api/admin/dishes/{id}/unpublish
// Снимает блюдо с публикации (published → approved)
func (h *AdminHandlers) UnpublishDish(w http.ResponseWriter, r *http.Request) {
	// Получаем admin ID из контекста
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	adminID := claims.Subject

	dishID := chi.URLParam(r, "id")
	if dishID == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Dish ID is required")
		return
	}

	// Снимаем с публикации
	err := h.service.UnpublishDish(dishID, adminID)
	if err != nil {
		if err.Error() == "dish not found" {
			utils.RespondWithError(w, http.StatusNotFound, "Dish not found")
			return
		}
		if err.Error() == "only published dishes can be unpublished" {
			utils.RespondWithError(w, http.StatusBadRequest, "Only published dishes can be unpublished")
			return
		}
		utils.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to unpublish dish: %v", err))
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Dish unpublished successfully",
	})
}

// DeleteDish - DELETE /api/admin/dishes/{id}
// Удаляет блюдо (только draft)
func (h *AdminHandlers) DeleteDish(w http.ResponseWriter, r *http.Request) {
	// Получаем admin ID из контекста
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	adminID := claims.Subject

	dishID := chi.URLParam(r, "id")
	if dishID == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Dish ID is required")
		return
	}

	// Удаляем блюдо
	err := h.service.DeleteDish(dishID, adminID)
	if err != nil {
		if err.Error() == "dish not found" {
			utils.RespondWithError(w, http.StatusNotFound, "Dish not found")
			return
		}
		if err.Error() == "only draft dishes can be deleted" {
			utils.RespondWithError(w, http.StatusBadRequest, "Only draft dishes can be deleted")
			return
		}
		utils.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to delete dish: %v", err))
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Dish deleted successfully",
	})
}
