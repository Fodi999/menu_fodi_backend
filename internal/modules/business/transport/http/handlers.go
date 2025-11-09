package http

import (
	"encoding/json"
	"net/http"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/business/dto"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/business/service"
	"github.com/dmitrijfomin/menu-fodifood/backend/pkg/utils"
	"github.com/go-chi/chi/v5"
)

type BusinessHandlers struct {
	svc *service.BusinessService
}

func NewBusinessHandlers(svc *service.BusinessService) *BusinessHandlers {
	return &BusinessHandlers{svc: svc}
}

func (h *BusinessHandlers) GetBusinesses(w http.ResponseWriter, r *http.Request) {
	businesses, err := h.svc.GetBusinesses()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch businesses")
		return
	}
	utils.RespondWithJSON(w, http.StatusOK, businesses)
}

func (h *BusinessHandlers) GetBusinessByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	business, err := h.svc.GetBusinessByID(id)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Business not found")
		return
	}
	utils.RespondWithJSON(w, http.StatusOK, business)
}

func (h *BusinessHandlers) CreateBusiness(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateBusinessRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	business, err := h.svc.CreateBusiness(&req)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	utils.RespondWithJSON(w, http.StatusCreated, business)
}

func (h *BusinessHandlers) UpdateBusiness(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req dto.UpdateBusinessRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	business, err := h.svc.UpdateBusiness(id, &req)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	utils.RespondWithJSON(w, http.StatusOK, business)
}

func (h *BusinessHandlers) DeleteBusiness(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.DeleteBusiness(id); err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Business not found")
		return
	}
	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Business deleted"})
}

func (h *BusinessHandlers) GetBusinessTokens(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	token, err := h.svc.GetBusinessTokens(id)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Business tokens not found")
		return
	}
	utils.RespondWithJSON(w, http.StatusOK, token)
}
