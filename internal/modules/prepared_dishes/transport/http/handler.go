package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/database"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/middleware"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
)

type PreparedDishesHandler struct {
	repo        *database.PreparedDishRepository
	historyRepo *database.HistoryRepository
	budgetRepo  *database.WeeklyBudgetRepository
}

func NewPreparedDishesHandler(repo *database.PreparedDishRepository, historyRepo *database.HistoryRepository, budgetRepo *database.WeeklyBudgetRepository) *PreparedDishesHandler {
	return &PreparedDishesHandler{
		repo:        repo,
		historyRepo: historyRepo,
		budgetRepo:  budgetRepo,
	}
}

// GetPreparedDishes returns user's prepared dishes with optional filters
// GET /api/prepared-dishes?category=pizza&available=true
func (h *PreparedDishesHandler) GetPreparedDishes(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(middleware.UserContextKey).(*models.User)
	if !ok || user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Unauthorized",
		})
		return
	}

	// Parse filters
	category := r.URL.Query().Get("category")
	availableOnly := r.URL.Query().Get("available") == "true"
	expiredOnly := r.URL.Query().Get("expired") == "true"

	var dishes []models.PreparedDish
	var err error

	if category != "" || availableOnly || expiredOnly {
		filters := database.PreparedDishFilters{
			Category:      category,
			AvailableOnly: availableOnly,
			ExpiredOnly:   expiredOnly,
		}
		dishes, err = h.repo.GetByUserIDWithFilters(user.ID, filters)
	} else {
		dishes, err = h.repo.GetByUserID(user.ID)
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
		"data":    dishes,
	})
}

// ConsumePortion consumes portions from a prepared dish
// POST /api/prepared-dishes/{id}/consume
func (h *PreparedDishesHandler) ConsumePortion(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(middleware.UserContextKey).(*models.User)
	if !ok || user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Unauthorized",
		})
		return
	}

	dishID := chi.URLParam(r, "id")
	if dishID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Dish ID required",
		})
		return
	}

	var req struct {
		Portions int `json:"portions"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Portions <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Invalid portions value",
		})
		return
	}

	// Verify ownership
	dish, err := h.repo.FindByID(dishID)
	if err != nil || dish == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Dish not found",
		})
		return
	}
	if dish.UserID != user.ID {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Access denied",
		})
		return
	}

	// Consume portions
	updated, err := h.repo.ConsumePortions(dishID, req.Portions)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Create history event for consumption
	recipeName := ""
	if updated.Recipe != nil {
		recipeName = updated.Recipe.LocalName
	}

	portions := req.Portions
	metadata := map[string]interface{}{
		"recipe_name":        recipeName,
		"portions_remaining": updated.PortionsAvailable,
		"dish_id":            dishID,
	}

	err = h.historyRepo.CreateWithMetadata(
		user.ID,
		models.EventTypeConsume,
		models.SourceTypePreparedDish,
		&dishID,
		&portions,
		metadata,
	)
	if err != nil {
		// Log error but don't fail the request
		// History is nice-to-have, not critical
	}

	// Update weekly budget (critical for budget tracking)
	if updated.CostPerPortion != nil {
		consumedCost := *updated.CostPerPortion * float64(req.Portions)
		err = h.budgetRepo.UpdateSpentBudget(user.ID, consumedCost)
		if err != nil {
			// Log error but don't fail - budget is tracked async
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":            true,
		"data":               updated,
		"portions_consumed":  req.Portions,
		"portions_remaining": updated.PortionsAvailable,
	})
}

// WasteDish marks a prepared dish as wasted (expired/discarded)
// POST /api/prepared-dishes/{id}/waste
func (h *PreparedDishesHandler) WasteDish(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(middleware.UserContextKey).(*models.User)
	if !ok || user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Unauthorized",
		})
		return
	}

	dishID := chi.URLParam(r, "id")
	if dishID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Dish ID required",
		})
		return
	}

	// Verify ownership
	dish, err := h.repo.FindByID(dishID)
	if err != nil || dish == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Dish not found",
		})
		return
	}
	if dish.UserID != user.ID {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Access denied",
		})
		return
	}

	// Calculate waste cost (remaining portions * cost per portion)
	wasteCost := 0.0
	if dish.CostPerPortion != nil {
		wasteCost = *dish.CostPerPortion * float64(dish.PortionsAvailable)
	}

	// Create history event for waste
	recipeName := ""
	if dish.Recipe != nil {
		recipeName = dish.Recipe.LocalName
	}

	portionsWasted := dish.PortionsAvailable
	metadata := map[string]interface{}{
		"recipe_name":     recipeName,
		"portions_wasted": portionsWasted,
		"waste_cost":      wasteCost,
		"dish_id":         dishID,
		"reason":          "manual_waste", // Could be expanded with user input
	}

	err = h.historyRepo.CreateWithMetadata(
		user.ID,
		models.EventTypeWaste,
		models.SourceTypePreparedDish,
		&dishID,
		&portionsWasted,
		metadata,
	)
	if err != nil {
		// Log but continue
	}

	// Update weekly budget waste_cost
	if wasteCost > 0 {
		err = h.budgetRepo.UpdateWasteCost(user.ID, wasteCost)
		if err != nil {
			// Log but continue
		}
	}

	// Delete the dish (it's wasted)
	err = h.repo.Delete(dishID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Failed to delete wasted dish",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":         true,
		"message":         "Dish marked as wasted",
		"portions_wasted": portionsWasted,
		"waste_cost":      wasteCost,
	})
}

// GetPreparedDishesStats returns statistics
// GET /api/prepared-dishes/stats
func (h *PreparedDishesHandler) GetPreparedDishesStats(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(middleware.UserContextKey).(*models.User)
	if !ok || user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Unauthorized",
		})
		return
	}

	dishes, err := h.repo.GetByUserID(user.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	totalDishes := len(dishes)
	totalPortions := 0
	availablePortions := 0
	consumedPortions := 0
	expiredDishes := 0

	for _, dish := range dishes {
		totalPortions += dish.PortionsInitial
		availablePortions += dish.PortionsAvailable
		consumedPortions += dish.ConsumedPortions()
		if dish.IsExpired() {
			expiredDishes++
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"total_dishes":       totalDishes,
			"total_portions":     totalPortions,
			"available_portions": availablePortions,
			"consumed_portions":  consumedPortions,
			"expired_dishes":     expiredDishes,
		},
	})
}
