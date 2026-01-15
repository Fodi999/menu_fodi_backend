package http

import (
	"encoding/json"
	"net/http"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/fridge/service"
	"github.com/dmitrijfomin/menu-fodifood/backend/pkg/utils"
	"github.com/go-chi/chi/v5"
)

type FridgeHandlersV2 struct {
	service service.FridgeServiceV2
}

func NewFridgeHandlersV2(svc service.FridgeServiceV2) *FridgeHandlersV2 {
	return &FridgeHandlersV2{service: svc}
}

// GetItems GET /api/fridge/items - получить все продукты пользователя
func (h *FridgeHandlersV2) GetItems(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)

	items, err := h.service.GetItems(userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch fridge items")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": items,
	})
}

// AddItem POST /api/fridge/items - добавить продукт
func (h *FridgeHandlersV2) AddItem(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)

	var req service.AddFridgeItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	item, err := h.service.AddItem(userID, req)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"data": item,
	})
}

// UpdateItem PATCH /api/fridge/items/:id - обновить продукт
func (h *FridgeHandlersV2) UpdateItem(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	itemID := chi.URLParam(r, "id")

	var req service.UpdateFridgeItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	item, err := h.service.UpdateItem(itemID, userID, req)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": item,
	})
}

// DeleteItem DELETE /api/fridge/items/:id - удалить продукт
func (h *FridgeHandlersV2) DeleteItem(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	itemID := chi.URLParam(r, "id")

	if err := h.service.DeleteItem(itemID, userID); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Item deleted successfully",
	})
}

// DiscardItem POST /api/fridge/items/:id/discard - выбросить продукт
func (h *FridgeHandlersV2) DiscardItem(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	itemID := chi.URLParam(r, "id")

	if err := h.service.DiscardItem(itemID, userID); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Item discarded successfully",
	})
}
