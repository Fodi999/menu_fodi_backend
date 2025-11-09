package meal_plan

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/meal_plan/service"
	httphandlers "github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/meal_plan/transport/http"
)

// Module represents meal plan module
type Module struct {
	handlers *httphandlers.MealPlanHandlers
}

// NewModule creates a new meal plan module
func NewModule() *Module {
	svc := service.NewMealPlanService()
	handlers := httphandlers.NewMealPlanHandlers(svc)

	return &Module{
		handlers: handlers,
	}
}

// RegisterRoutes registers all meal plan routes
func (m *Module) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware)
		r.Post("/meal-plan", m.handlers.GenerateMealPlan)
	})
}
