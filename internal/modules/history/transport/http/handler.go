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

// GetFridgeLosses returns analytics for expired/wasted items
// GET /api/history/losses?days=30
func (h *HistoryHandler) GetFridgeLosses(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(middleware.UserContextKey).(*models.User)
	if !ok || user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Unauthorized",
		})
		return
	}

	// Parse days parameter
	daysStr := r.URL.Query().Get("days")
	days := 30 // default last 30 days
	if daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
			days = d
		}
	}

	// Calculate date range
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)

	// Get expired/wasted events
	filters := database.HistoryFilters{
		EventType: "expired",
		StartDate: &startDate,
		EndDate:   &endDate,
		Limit:     1000, // Large limit for analytics
	}

	events, err := h.repo.GetByFilters(user.ID, filters)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Calculate analytics
	totalCost := 0.0
	totalItems := len(events)
	categoryBreakdown := make(map[string]int)
	costBreakdown := make(map[string]float64)

	for _, event := range events {
		var metadata models.ExpiredItemMetadata
		if err := json.Unmarshal(event.Metadata, &metadata); err == nil {
			totalCost += metadata.Cost
			
			// Category breakdown (you can enhance this with ingredient category)
			categoryBreakdown["items"]++
			costBreakdown["total"] += metadata.Cost
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"period": map[string]interface{}{
				"days":      days,
				"startDate": startDate.Format("2006-01-02"),
				"endDate":   endDate.Format("2006-01-02"),
			},
			"summary": map[string]interface{}{
				"totalItems":      totalItems,
				"totalCost":       totalCost,
				"currency":        "PLN",
				"avgCostPerItem":  func() float64 {
					if totalItems > 0 {
						return totalCost / float64(totalItems)
					}
					return 0
				}(),
			},
			"items": events,
		},
	})
}
