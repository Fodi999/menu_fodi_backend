package http

import (
	"net/http"
	"strconv"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/stats/service"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/platform/httpx"
)

type StatsHandlers struct {
	service *service.StatsService
}

func NewStatsHandlers(srv *service.StatsService) *StatsHandlers {
	return &StatsHandlers{service: srv}
}

func (h *StatsHandlers) GetAdminStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.GetAdminStats()
	if err != nil {
		httpx.InternalError(w, "Failed to fetch stats")
		return
	}

	httpx.Success(w, stats)
}

func (h *StatsHandlers) GetRecentOrders(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 10

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 50 {
			limit = l
		}
	}

	orders, err := h.service.GetRecentOrders(limit)
	if err != nil {
		httpx.InternalError(w, "Failed to fetch recent orders")
		return
	}

	httpx.Success(w, orders)
}
