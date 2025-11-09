package http

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
	svc *service.SemiFinishedService
}

// NewSemiFinishedHandlers creates a new handlers instance
func NewSemiFinishedHandlers(svc *service.SemiFinishedService) *SemiFinishedHandlers {
	return &SemiFinishedHandlers{
		svc: svc,
	}
}

// GetAllSemiFinished retrieves all semi-finished products
// GET /api/semi-finished
func (h *SemiFinishedHandlers) GetAllSemiFinished(w http.ResponseWriter, r *http.Request) {
	semiFinished, err := h.svc.GetAll()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch semi-finished products")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, semiFinished)
}

// GetSemiFinishedByID retrieves a semi-finished product by ID
// GET /api/semi-finished/{id}
func (h *SemiFinishedHandlers) GetSemiFinishedByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	sf, err := h.svc.GetByID(id)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Semi-finished product not found")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, sf)
}

// CreateSemiFinished creates a new semi-finished product
// POST /api/semi-finished
func (h *SemiFinishedHandlers) CreateSemiFinished(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateSemiFinishedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	sf, err := h.svc.Create(&req)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusCreated, sf)
}

// UpdateSemiFinished updates a semi-finished product
// PUT /api/semi-finished/{id}
func (h *SemiFinishedHandlers) UpdateSemiFinished(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req dto.UpdateSemiFinishedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	sf, err := h.svc.Update(id, &req)
	if err != nil {
		if err.Error() == "semi-finished not found" {
			utils.RespondWithError(w, http.StatusNotFound, "Semi-finished product not found")
		} else {
			utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, sf)
}

// DeleteSemiFinished deletes a semi-finished product
// DELETE /api/semi-finished/{id}
func (h *SemiFinishedHandlers) DeleteSemiFinished(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.svc.Delete(id); err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Semi-finished product not found")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Semi-finished product deleted successfully"})
}
