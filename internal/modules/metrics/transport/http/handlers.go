package http

import (
	"net/http"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/metrics/service"
	"github.com/dmitrijfomin/menu-fodifood/backend/pkg/utils"
	"github.com/go-chi/chi/v5"
)

type MetricsHandlers struct {
	svc *service.MetricsService
}

func NewMetricsHandlers(svc *service.MetricsService) *MetricsHandlers {
	return &MetricsHandlers{svc: svc}
}

func (h *MetricsHandlers) GetBusinessMetrics(w http.ResponseWriter, r *http.Request) {
	businessID := chi.URLParam(r, "businessId")
	metrics, err := h.svc.GetBusinessMetrics(businessID)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Business not found")
		return
	}
	utils.RespondWithJSON(w, http.StatusOK, metrics)
}
