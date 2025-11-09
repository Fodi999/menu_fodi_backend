package nutrition

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/nutrition/service"
	httphandlers "github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/nutrition/transport/http"
)

// Module represents nutrition module
type Module struct {
	handlers *httphandlers.NutritionHandlers
}

// NewModule creates a new nutrition module
func NewModule() *Module {
	nutritionService := service.NewNutritionService()
	handlers := httphandlers.NewNutritionHandlers(nutritionService)

	return &Module{
		handlers: handlers,
	}
}

// RegisterRoutes registers all nutrition routes
func (m *Module) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.Route("/nutrition", func(r chi.Router) {
		// Public routes
		r.Get("/recipe/{id}", m.handlers.GetRecipeNutrition)

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware)
			r.Post("/calculate", m.handlers.CalculateCustomNutrition)
		})
	})
}
