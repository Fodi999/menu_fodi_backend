package ai

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/ai/service"
	aihttp "github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/ai/transport/http"
)

// Module represents AI module
type Module struct {
	handlers *aihttp.AIHandlers
}

// NewModule creates new AI module
func NewModule(db *gorm.DB) *Module {
	svc := service.NewAIService()
	handlers := aihttp.NewAIHandlers(svc, db)

	return &Module{
		handlers: handlers,
	}
}

// RegisterRoutes registers AI routes
func (m *Module) RegisterRoutes(r chi.Router, jwtMiddleware func(http.Handler) http.Handler) {
	r.Route("/ai", func(r chi.Router) {
		// Public AI endpoints
		r.Post("/chef-mentor", m.handlers.ChefMentor)
		r.Post("/recipe-generator", m.handlers.GenerateRecipe)

		// Protected AI endpoints (require auth)
		r.Group(func(r chi.Router) {
			r.Use(jwtMiddleware)

			r.Post("/meal-plan", m.handlers.GenerateMealPlan)
			r.Post("/fridge-recommendations", m.handlers.GetFridgeRecommendations)
			r.Post("/save-ingredients", m.handlers.SaveRecipeIngredientsToFridge)
			r.Post("/fridge/analyze", m.handlers.AnalyzeFridge)
			r.Post("/create-recipe-from-fridge", m.handlers.CreateRecipeFromFridge)
			r.Post("/add-missing-ingredients", m.handlers.AddMissingIngredients) // NEW: Add ingredientsMissing to fridge
		})
	})
}
