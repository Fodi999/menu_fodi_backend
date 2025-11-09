package recipes

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	httphandlers "github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/recipes/transport/http"
)

type Module struct {
	handlers *httphandlers.RecipeHandlers
}

func NewModule() *Module {
	handlers := httphandlers.NewRecipeHandlers()
	return &Module{handlers: handlers}
}

func (m *Module) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	// Public routes
	r.Get("/posts", m.handlers.GetAllPosts)
	r.Get("/users/{id}/posts", m.handlers.GetUserPosts)
	r.Get("/user/{id}/posts", m.handlers.GetUserPosts)
	r.Post("/recipes/{id}/view", m.handlers.IncrementRecipeView)

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware)
		r.Post("/recipes", m.handlers.CreateRecipe)
		r.Put("/recipes/{id}", m.handlers.UpdateRecipe)
		r.Delete("/recipes/{id}", m.handlers.DeleteRecipe)
	})
}
