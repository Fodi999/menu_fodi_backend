package recipes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/recipes/service"
	httphandlers "github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/recipes/transport/http"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/platform/logger"
)

type Module struct {
	oldHandlers    *httphandlers.RecipeHandlers // Old user recipe handlers
	catalogHandler *httphandlers.RecipeHandler  // New catalog recipe handler
}

func NewModule(db *gorm.DB) *Module {
	// Old handlers (user recipes)
	oldHandlers := httphandlers.NewRecipeHandlers()
	
	// New catalog handlers (recipe matching, adaptation & cooking)
	matchService := service.NewRecipeMatchService(db)
	adapterService := service.NewRecipeAdapterService(db, nil) // TODO: Pass Groq client
	cookService := service.NewRecipeCookService(db)
	catalogHandler := httphandlers.NewRecipeHandler(matchService, adapterService, cookService, logger.Log)
	
	return &Module{
		oldHandlers:    oldHandlers,
		catalogHandler: catalogHandler,
	}
}

func (m *Module) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	// === OLD USER RECIPE ROUTES ===
	// Public routes
	r.Get("/posts", m.oldHandlers.GetAllPosts)
	r.Get("/users/{id}/posts", m.oldHandlers.GetUserPosts)
	r.Get("/user/{id}/posts", m.oldHandlers.GetUserPosts)
	r.Post("/recipes/{id}/view", m.oldHandlers.IncrementRecipeView)

	// Protected routes (user recipes)
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware)
		r.Post("/recipes", m.oldHandlers.CreateRecipe)
		r.Put("/recipes/{id}", m.oldHandlers.UpdateRecipe)
		r.Delete("/recipes/{id}", m.oldHandlers.DeleteRecipe)
	})

	// === NEW CATALOG RECIPE ROUTES ===
	// TODO: Remove public access after testing - these should be protected
	// Recipe matching (finds recipes based on fridge) - TEMPORARILY PUBLIC FOR TESTING
	r.Get("/recipes/match", m.catalogHandler.MatchRecipes)
	// Recipe recommendation (returns 1 best recipe for UI) - TEMPORARILY PUBLIC FOR TESTING
	r.Post("/recipes/recommendations", m.catalogHandler.GetRecommendation)
	// Recipe cooking - TEMPORARILY PUBLIC FOR TESTING (uses testUserID)
	r.Post("/recipes/{id}/cook", m.catalogHandler.CookRecipe)
	
	// Protected routes (require auth)
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware)
		// Recipe adaptation (AI adapts recipe to available ingredients)
		r.Post("/recipes/{id}/adapt", m.catalogHandler.AdaptRecipe)
		
		// Recipe detail (TODO: Implement)
		// r.Get("/recipes/{id}", m.catalogHandler.GetRecipeByID)
	})
}
