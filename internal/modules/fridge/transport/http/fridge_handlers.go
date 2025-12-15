package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/middleware"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/fridge/service"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/platform/logger"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// FridgeHandlers HTTP handlers для работы с холодильником
type FridgeHandlers struct {
	service *service.FridgeService
}

// NewFridgeHandlers создает новый экземпляр handlers
func NewFridgeHandlers(service *service.FridgeService) *FridgeHandlers {
	return &FridgeHandlers{service: service}
}

// AddItem добавляет продукт в холодильник
func (h *FridgeHandlers) AddItem(w http.ResponseWriter, r *http.Request) {
	// Получаем User ID из контекста (установлен middleware)
	userIDPtr := middleware.GetUserID(r)
	if userIDPtr == nil {
		logger.Error("user ID not found in context")
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	userID := userIDPtr.String()

	// Парсим запрос
	var req models.CreateFridgeItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("failed to decode request", zap.Error(err))
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Добавляем продукт
	response, err := h.service.AddItem(userID, req)
	if err != nil {
		logger.Error("failed to add fridge item",
			zap.Error(err),
			zap.String("user_id", userID),
			zap.String("ingredient_id", req.IngredientID))
		respondError(w, http.StatusInternalServerError, "failed to add item")
		return
	}

	respondSuccess(w, response)
}

// GetUserItems возвращает список продуктов пользователя
func (h *FridgeHandlers) GetUserItems(w http.ResponseWriter, r *http.Request) {
	// Получаем User ID из контекста
	userIDPtr := middleware.GetUserID(r)
	if userIDPtr == nil {
		logger.Error("user ID not found in context")
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	userID := userIDPtr.String()

	// Получаем список продуктов
	items, err := h.service.GetUserItems(userID)
	if err != nil {
		logger.Error("failed to get fridge items",
			zap.Error(err),
			zap.String("user_id", userID))
		respondError(w, http.StatusInternalServerError, "failed to get items")
		return
	}

	respondSuccess(w, map[string]interface{}{
		"items": items,
	})
}

// DeleteItem удаляет продукт из холодильника
func (h *FridgeHandlers) DeleteItem(w http.ResponseWriter, r *http.Request) {
	// Получаем User ID из контекста
	userIDPtr := middleware.GetUserID(r)
	if userIDPtr == nil {
		logger.Error("user ID not found in context")
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	userID := userIDPtr.String()

	// Получаем ID продукта из URL path parameter (chi router)
	itemID := chi.URLParam(r, "id")
	if itemID == "" {
		respondError(w, http.StatusBadRequest, "item ID is required")
		return
	}

	// Удаляем продукт
	if err := h.service.DeleteItem(itemID, userID); err != nil {
		logger.Error("failed to delete fridge item",
			zap.Error(err),
			zap.String("user_id", userID),
			zap.String("item_id", itemID))
		respondError(w, http.StatusInternalServerError, "failed to delete item")
		return
	}

	respondSuccess(w, map[string]interface{}{
		"success": true,
		"message": "item deleted successfully",
	})
}

// AddPrice добавляет событие изменения цены (event sourcing)
func (h *FridgeHandlers) AddPrice(w http.ResponseWriter, r *http.Request) {
	// Получаем User ID из контекста
	userIDPtr := middleware.GetUserID(r)
	if userIDPtr == nil {
		logger.Error("user ID not found in context")
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	userID := userIDPtr.String()

	// Получаем ID продукта из URL
	itemID := chi.URLParam(r, "id")
	if itemID == "" {
		respondError(w, http.StatusBadRequest, "item ID is required")
		return
	}

	// Парсим запрос
	var req models.AddPriceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Добавляем событие изменения цены
	if err := h.service.AddPrice(userID, itemID, req); err != nil {
		logger.Error("failed to add price",
			zap.Error(err),
			zap.String("user_id", userID),
			zap.String("item_id", itemID))
		
		// Маппируем доменные ошибки на HTTP коды
		switch {
		case errors.Is(err, service.ErrNotFound):
			respondError(w, http.StatusNotFound, "item not found")
			return
		case errors.Is(err, service.ErrForbidden):
			respondError(w, http.StatusForbidden, "access denied")
			return
		case errors.Is(err, service.ErrInvalidSource):
			respondError(w, http.StatusBadRequest, err.Error())
			return
		case errors.Is(err, service.ErrInvalidInput):
			respondError(w, http.StatusBadRequest, "invalid input")
			return
		default:
			// Логируем полную ошибку для отладки
			logger.Error("unexpected error in AddPrice", zap.Error(err))
			respondError(w, http.StatusInternalServerError, "failed to add price")
			return
		}
	}

	respondSuccess(w, map[string]interface{}{
		"success": true,
		"message": "price added successfully",
	})
}

// GetPriceHistory возвращает историю изменения цен
func (h *FridgeHandlers) GetPriceHistory(w http.ResponseWriter, r *http.Request) {
	// Получаем User ID из контекста
	userIDPtr := middleware.GetUserID(r)
	if userIDPtr == nil {
		logger.Error("user ID not found in context")
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	userID := userIDPtr.String()

	// Получаем ID продукта из URL
	itemID := chi.URLParam(r, "id")
	if itemID == "" {
		respondError(w, http.StatusBadRequest, "item ID is required")
		return
	}

	// Получаем историю
	history, err := h.service.GetPriceHistory(userID, itemID)
	if err != nil {
		logger.Error("failed to get price history",
			zap.Error(err),
			zap.String("user_id", userID),
			zap.String("item_id", itemID))
		
		// Маппируем доменные ошибки на HTTP коды
		switch {
		case errors.Is(err, service.ErrNotFound):
			respondError(w, http.StatusNotFound, "item not found")
			return
		case errors.Is(err, service.ErrForbidden):
			respondError(w, http.StatusForbidden, "access denied")
			return
		default:
			logger.Error("unexpected error in GetPriceHistory", zap.Error(err))
			respondError(w, http.StatusInternalServerError, "failed to get price history")
			return
		}
	}

	respondSuccess(w, map[string]interface{}{
		"history": history,
		"count":   len(history),
	})
}

// UpdateItemQuantity обновляет количество продукта
func (h *FridgeHandlers) UpdateItemQuantity(w http.ResponseWriter, r *http.Request) {
	// Получаем User ID из контекста
	userIDPtr := middleware.GetUserID(r)
	if userIDPtr == nil {
		logger.Error("user ID not found in context")
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	userID := userIDPtr.String()

	// Получаем ID продукта из URL
	itemID := chi.URLParam(r, "id")
	if itemID == "" {
		respondError(w, http.StatusBadRequest, "item ID is required")
		return
	}

	// Парсим запрос
	var req struct {
		Quantity float64 `json:"quantity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("failed to decode request", zap.Error(err))
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Валидация
	if req.Quantity <= 0 {
		respondError(w, http.StatusBadRequest, "quantity must be greater than 0")
		return
	}

	// Обновляем количество
	if err := h.service.UpdateItemQuantity(userID, itemID, req.Quantity); err != nil {
		logger.Error("failed to update quantity",
			zap.Error(err),
			zap.String("user_id", userID),
			zap.String("item_id", itemID),
			zap.Float64("new_quantity", req.Quantity))
		
		// Маппируем доменные ошибки на HTTP коды
		switch {
		case errors.Is(err, service.ErrNotFound):
			respondError(w, http.StatusNotFound, "item not found")
			return
		case errors.Is(err, service.ErrForbidden):
			respondError(w, http.StatusForbidden, "access denied")
			return
		default:
			logger.Error("unexpected error in UpdateItemQuantity", zap.Error(err))
			respondError(w, http.StatusInternalServerError, "failed to update quantity")
			return
		}
	}

	respondSuccess(w, map[string]interface{}{
		"message": "Quantity updated successfully",
	})
}

// Helper functions for consistent responses

func respondSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    data,
	})
}

func respondError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"message": message,
	})
}
