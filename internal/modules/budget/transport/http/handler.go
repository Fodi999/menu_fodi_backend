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

type BudgetHandler struct {
	repo *database.WeeklyBudgetRepository
}

func NewBudgetHandler(repo *database.WeeklyBudgetRepository) *BudgetHandler {
	return &BudgetHandler{repo: repo}
}

// GetCurrentWeekBudget returns budget for current week
// GET /api/budget/current
func (h *BudgetHandler) GetCurrentWeekBudget(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(middleware.UserContextKey).(*models.User)
	if !ok || user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Unauthorized",
		})
		return
	}

	budget, err := h.repo.GetOrCreateCurrentWeek(user.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Add computed fields
	response := map[string]interface{}{
		"budget":            budget,
		"remaining":         budget.GetRemainingBudget(),
		"waste_percentage":  budget.GetWastePercentage(),
		"spent_percentage":  budget.GetSpentPercentage(),
		"is_over_budget":    budget.IsOverBudget(),
		"saved_money":       budget.GetSavedMoney(),
		"efficiency_score":  budget.GetEfficiencyScore(),
		"week_number":       budget.GetWeekNumber(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    response,
	})
}

// GetWeeklyBudgets returns budgets for recent weeks
// GET /api/budget/weekly?weeks=4
func (h *BudgetHandler) GetWeeklyBudgets(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(middleware.UserContextKey).(*models.User)
	if !ok || user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Unauthorized",
		})
		return
	}

	// Parse weeks parameter
	weeksStr := r.URL.Query().Get("weeks")
	weeks := 4 // default
	if weeksStr != "" {
		if w, err := strconv.Atoi(weeksStr); err == nil && w > 0 {
			weeks = w
		}
	}

	budgets, err := h.repo.GetRecentWeeks(user.ID, weeks)
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
		"data":    budgets,
	})
}

// SetPlannedBudget sets the planned budget for current week
// PUT /api/budget/plan
func (h *BudgetHandler) SetPlannedBudget(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(middleware.UserContextKey).(*models.User)
	if !ok || user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Unauthorized",
		})
		return
	}

	var req struct {
		Amount float64 `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Amount < 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Invalid amount",
		})
		return
	}

	err := h.repo.SetPlannedBudget(user.ID, req.Amount)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Get updated budget
	budget, _ := h.repo.GetOrCreateCurrentWeek(user.ID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Planned budget updated",
		"data":    budget,
	})
}

// GetBudgetStats returns aggregated statistics
// GET /api/budget/stats
func (h *BudgetHandler) GetBudgetStats(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(middleware.UserContextKey).(*models.User)
	if !ok || user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Unauthorized",
		})
		return
	}

	stats, err := h.repo.GetWeeklyStats(user.ID)
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

// GetBudgetForWeek returns budget for specific week
// GET /api/budget/week/{date} where date is YYYY-MM-DD
func (h *BudgetHandler) GetBudgetForWeek(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(middleware.UserContextKey).(*models.User)
	if !ok || user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Unauthorized",
		})
		return
	}

	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Date parameter required (YYYY-MM-DD)",
		})
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Invalid date format",
		})
		return
	}

	budget, err := h.repo.GetByUserAndWeek(user.ID, date)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	if budget == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Budget not found for this week",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    budget,
	})
}
