package http

import (
	"encoding/json"
	"net/http"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/middleware"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/meal_plan/dto"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/meal_plan/service"
	"github.com/dmitrijfomin/menu-fodifood/backend/pkg/utils"
)

// MealPlanHandlers handles HTTP requests for meal planning
type MealPlanHandlers struct {
	svc *service.MealPlanService
}

// NewMealPlanHandlers creates a new handlers instance
func NewMealPlanHandlers(svc *service.MealPlanService) *MealPlanHandlers {
	return &MealPlanHandlers{
		svc: svc,
	}
}

// GenerateMealPlan generates a meal plan
// POST /api/meal-plan
func (h *MealPlanHandlers) GenerateMealPlan(w http.ResponseWriter, r *http.Request) {
	var req dto.MealPlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get user ID from context
	userID := middleware.GetUserID(r).String()

	plan, err := h.svc.GenerateMealPlan(&req, userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, plan)
}
