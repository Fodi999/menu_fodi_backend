package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/database"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/middleware"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
)

type HistoryHandler struct {
	repo *database.HistoryRepository
}

func NewHistoryHandler(repo *database.HistoryRepository) *HistoryHandler {
	return &HistoryHandler{repo: repo}
}

// GetHistory returns user's history events with optional filters
// GET /api/history?type=consume&limit=50&start_date=2025-01-01
func (h *HistoryHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(middleware.UserContextKey).(*models.User)
	if !ok || user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Unauthorized",
		})
		return
	}

	// Parse query parameters
	eventType := r.URL.Query().Get("type")
	sourceType := r.URL.Query().Get("source_type")
	limitStr := r.URL.Query().Get("limit")
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	// Parse limit
	limit := 50 // default
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	// Parse dates
	var startDate, endDate *time.Time
	if startDateStr != "" {
		if t, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = &t
		}
	}
	if endDateStr != "" {
		if t, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = &t
		}
	}

	var events []models.HistoryEvent
	var err error

	// Use filtered query if filters present
	if eventType != "" || sourceType != "" || startDate != nil || endDate != nil {
		filters := database.HistoryFilters{
			EventType:  eventType,
			SourceType: sourceType,
			StartDate:  startDate,
			EndDate:    endDate,
			Limit:      limit,
		}
		events, err = h.repo.GetByFilters(user.ID, filters)
	} else {
		events, err = h.repo.GetByUserID(user.ID, limit)
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    events,
		"count":   len(events),
	})
}

// GetHistoryStats returns analytics statistics for user's history
// GET /api/history/stats?start_date=2025-01-01&end_date=2025-12-31
func (h *HistoryHandler) GetHistoryStats(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(middleware.UserContextKey).(*models.User)
	if !ok || user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Unauthorized",
		})
		return
	}

	// Parse date range
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	var startDate, endDate *time.Time
	if startDateStr != "" {
		if t, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = &t
		}
	}
	if endDateStr != "" {
		if t, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = &t
		}
	}

	stats, err := h.repo.GetStatsByUser(user.ID, startDate, endDate)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    stats,
	})
}

// GetRecentActivity returns recent history events for dashboard
// GET /api/history/recent?limit=10
func (h *HistoryHandler) GetRecentActivity(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(middleware.UserContextKey).(*models.User)
	if !ok || user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Unauthorized",
		})
		return
	}

	// Parse limit
	limitStr := r.URL.Query().Get("limit")
	limit := 10 // default
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	events, err := h.repo.GetRecentActivity(user.ID, limit)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    events,
	})
}
