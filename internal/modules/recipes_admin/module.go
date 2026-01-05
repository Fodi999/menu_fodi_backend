package recipes_admin

import (
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/middleware"
	httphandler "github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/recipes_admin/transport/http"
	"github.com/go-chi/chi/v5"
)

// Module represents the recipes_admin module
type Module struct {
	handlers *httphandler.RecipeAdminHandlers
}

// NewModule creates a new recipes_admin module
func NewModule() *Module {
	return &Module{
		handlers: httphandler.NewRecipeAdminHandlers(),
	}
}

// RegisterRoutes registers all routes for the recipes_admin module
func (m *Module) RegisterRoutes(r chi.Router) {
	r.Route("/api/admin/recipes", func(r chi.Router) {
		// All routes require authentication
		r.Use(middleware.AuthMiddleware)
		
		// Create draft recipe (minimal validation)
		r.Post("/", m.handlers.CreateDraft)
		
		// Get all draft recipes
		r.Get("/drafts", m.handlers.GetDrafts)
		
		// Update draft recipe (PATCH - partial updates)
		r.Patch("/{id}", m.handlers.UpdateDraft)
		
		// Publish recipe (full validation)
		r.Post("/{id}/publish", m.handlers.Publish)
		
		// Archive recipe
		r.Post("/{id}/archive", m.handlers.Archive)
	})
}

// Name returns the module name
func (m *Module) Name() string {
	return "recipes_admin"
}
