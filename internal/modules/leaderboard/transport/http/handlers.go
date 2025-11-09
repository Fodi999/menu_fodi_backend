package http

import (
	"net/http"
	"strconv"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/leaderboard/service"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/platform/httpx"
)

type LeaderboardHandlers struct {
	service *service.LeaderboardService
}

func NewLeaderboardHandlers(srv *service.LeaderboardService) *LeaderboardHandlers {
	return &LeaderboardHandlers{service: srv}
}

func (h *LeaderboardHandlers) GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	sortBy := r.URL.Query().Get("sortBy")
	limitStr := r.URL.Query().Get("limit")
	limit := 50

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	response, err := h.service.GetLeaderboard(sortBy, limit)
	if err != nil {
		httpx.InternalError(w, "Failed to fetch leaderboard")
		return
	}

	httpx.Success(w, response)
}
