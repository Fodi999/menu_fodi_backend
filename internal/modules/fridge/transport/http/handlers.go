package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/middleware"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/fridge/dto"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/fridge/service"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/platform/httpx"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/platform/logger"
)

// FridgeHandlers contains fridge HTTP handlers
type FridgeHandlers struct {
	service service.FridgeService
}

// NewFridgeHandlers creates new fridge handlers
func NewFridgeHandlers(service service.FridgeService) *FridgeHandlers {
	return &FridgeHandlers{service: service}
}

// GetUserFridge godoc
// @Summary Get user fridge
// @Description Get all items in user's fridge
// @Tags fridge
// @Security BearerAuth
// @Produce json
// @Success 200 {object} dto.FridgeListResponse
// @Failure 401 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /api/fridge [get]
func (h *FridgeHandlers) GetUserFridge(w http.ResponseWriter, r *http.Request) {
	userIDPtr := middleware.GetUserID(r)
	if userIDPtr == nil {
		logger.Error("user ID not found in context")
		httpx.Unauthorized(w, "unauthorized")
		return
	}
	userID := *userIDPtr

	fridgeList, err := h.service.GetUserFridge(userID)
	if err != nil {
		logger.Error("failed to get fridge", zap.Error(err), zap.String("user_id", userID.String()))
		httpx.InternalError(w, "failed to get fridge items")
		return
	}

	httpx.Success(w, fridgeList)
}

// AddFridgeItem godoc
// @Summary Add fridge item
// @Description Add a new item to user's fridge
// @Tags fridge
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.AddFridgeItemRequest true "Add item request"
// @Success 200 {object} httpx.MessageResponse
// @Failure 400 {object} httpx.ErrorResponse
// @Failure 401 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /api/fridge [post]
func (h *FridgeHandlers) AddFridgeItem(w http.ResponseWriter, r *http.Request) {
	userIDPtr := middleware.GetUserID(r)
	if userIDPtr == nil {
		logger.Error("user ID not found in context")
		httpx.Unauthorized(w, "unauthorized")
		return
	}
	userID := *userIDPtr

	var req dto.AddFridgeItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("failed to decode request", zap.Error(err))
		httpx.BadRequest(w, "invalid request body")
		return
	}

	if err := h.service.AddItem(userID, req); err != nil {
		switch err {
		case service.ErrEmptyProduct, service.ErrInvalidQuantity, service.ErrEmptyUnit:
			httpx.BadRequest(w, err.Error())
		default:
			logger.Error("failed to add item", zap.Error(err), zap.String("user_id", userID.String()))
			httpx.InternalError(w, "failed to add item")
		}
		return
	}

	httpx.Success(w, map[string]interface{}{
		"success": true,
		"message": "item added to fridge",
	})
}

// UpdateFridgeItem godoc
// @Summary Update fridge item
// @Description Update quantity or availability of a fridge item
// @Tags fridge
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Item ID"
// @Param request body dto.UpdateFridgeItemRequest true "Update item request"
// @Success 200 {object} dto.FridgeItemResponse
// @Failure 400 {object} httpx.ErrorResponse
// @Failure 401 {object} httpx.ErrorResponse
// @Failure 404 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /api/fridge/{id} [put]
func (h *FridgeHandlers) UpdateFridgeItem(w http.ResponseWriter, r *http.Request) {
	userIDPtr := middleware.GetUserID(r)
	if userIDPtr == nil {
		logger.Error("user ID not found in context")
		httpx.Unauthorized(w, "unauthorized")
		return
	}
	userID := *userIDPtr

	itemIDStr := chi.URLParam(r, "id")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		httpx.BadRequest(w, "invalid item ID")
		return
	}

	var req dto.UpdateFridgeItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("failed to decode request", zap.Error(err))
		httpx.BadRequest(w, "invalid request body")
		return
	}

	item, err := h.service.UpdateItem(itemID, userID, req)
	if err != nil {
		switch err {
		case service.ErrNoUpdates:
			httpx.BadRequest(w, err.Error())
		default:
			logger.Error("failed to update item", zap.Error(err),
				zap.String("user_id", userID.String()),
				zap.String("item_id", itemID.String()))
			httpx.InternalError(w, "failed to update item")
		}
		return
	}

	httpx.Success(w, map[string]interface{}{
		"success": true,
		"message": "item updated successfully",
		"item":    item,
	})
}

// DeleteFridgeItem godoc
// @Summary Delete fridge item
// @Description Remove an item from user's fridge
// @Tags fridge
// @Security BearerAuth
// @Produce json
// @Param id path string true "Item ID"
// @Success 200 {object} httpx.MessageResponse
// @Failure 400 {object} httpx.ErrorResponse
// @Failure 401 {object} httpx.ErrorResponse
// @Failure 404 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /api/fridge/{id} [delete]
func (h *FridgeHandlers) DeleteFridgeItem(w http.ResponseWriter, r *http.Request) {
	userIDPtr := middleware.GetUserID(r)
	if userIDPtr == nil {
		logger.Error("user ID not found in context")
		httpx.Unauthorized(w, "unauthorized")
		return
	}
	userID := *userIDPtr

	itemIDStr := chi.URLParam(r, "id")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		httpx.BadRequest(w, "invalid item ID")
		return
	}

	if err := h.service.DeleteItem(itemID, userID); err != nil {
		logger.Error("failed to delete item", zap.Error(err),
			zap.String("user_id", userID.String()),
			zap.String("item_id", itemID.String()))
		httpx.InternalError(w, "failed to delete item")
		return
	}

	httpx.Success(w, map[string]interface{}{
		"success": true,
		"message": "item deleted successfully",
	})
}

// GetAvailableItems godoc
// @Summary Get available items
// @Description Get all available items in user's fridge
// @Tags fridge
// @Security BearerAuth
// @Produce json
// @Success 200 {array} dto.FridgeItemResponse
// @Failure 401 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /api/fridge/available [get]
func (h *FridgeHandlers) GetAvailableItems(w http.ResponseWriter, r *http.Request) {
	userIDPtr := middleware.GetUserID(r)
	if userIDPtr == nil {
		logger.Error("user ID not found in context")
		httpx.Unauthorized(w, "unauthorized")
		return
	}
	userID := *userIDPtr

	items, err := h.service.GetAvailableItems(userID)
	if err != nil {
		logger.Error("failed to get available items", zap.Error(err), zap.String("user_id", userID.String()))
		httpx.InternalError(w, "failed to get available items")
		return
	}

	httpx.Success(w, map[string]interface{}{
		"success": true,
		"items":   items,
		"count":   len(items),
	})
}
