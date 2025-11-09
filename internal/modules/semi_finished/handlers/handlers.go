package handlers
package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/semi_finished/dto"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/semi_finished/service"
	"github.com/dmitrijfomin/menu-fodifood/backend/pkg/utils"
	"github.com/go-chi/chi/v5"
)

// SemiFinishedHandlers handles HTTP requests for semi-finished products
type SemiFinishedHandlers struct {
	service *service.SemiFinishedService
}

// NewSemiFinishedHandlers creates a new handlers instance
func NewSemiFinishedHandlers(svc *service.SemiFinishedService) *SemiFinishedHandlers {
	return &SemiFinishedHandlers{service: svc}
}

// GetAll retrieves all semi-finished products
// GET /api/semi-finished
func (h *SemiFinishedHandlers) GetAll(w http.ResponseWriter, r *http.Request) {
	semiFinished, err := h.service.GetAll()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch semi-finished products")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, semiFinished)
}

// GetByID retrieves a semi-finished product by ID
// GET /api/semi-finished/{id}
func (h *SemiFinishedHandlers) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	sf, err := h.service.GetByID(id)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Semi-finished product not found")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, sf)
}

// Create creates a new semi-finished product
// POST /api/semi-finished
func (h *SemiFinishedHandlers) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateSemiFinishedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	sf, err := h.service.Create(&req)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusCreated, sf)
}

// Update updates a semi-finished product
// PUT /api/semi-finished/{id}
func (h *SemiFinishedHandlers) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req dto.UpdateSemiFinishedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	sf, err := h.service.Update(id, &req)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "✅ Semi-finished updated successfully",
		"data":    sf,
	})
}

// Delete deletes a semi-finished product
// DELETE /api/semi-finished/{id}
func (h *SemiFinishedHandlers) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.service.Delete(id); err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Semi-finished product not found")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Semi-finished deleted successfully"})
}
