package recipes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/database"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/middleware"
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

	// Initialize repositories
	sessionRepository := database.NewUserRecipeSessionRepository(db)
	savedRecipeRepo := database.NewUserSavedRecipeRepository(db)

	// New catalog handlers (recipe matching, adaptation & cooking)
	matchService := service.NewRecipeMatchService(db)
	adapterService := service.NewRecipeAdapterService(db, nil) // TODO: Pass Groq client
	cookService := service.NewRecipeCookService(db)
	catalogHandler := httphandlers.NewRecipeHandler(
		matchService,
		adapterService,
		cookService,
		sessionRepository,
		savedRecipeRepo,
		logger.Log,
	)

	return &Module{
		oldHandlers:    oldHandlers,
		catalogHandler: catalogHandler,
	}
}

func (m *Module) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	// === OLD USER RECIPE ROUTES (DISABLED - NOT USED BY FRONTEND) ===
	// Public routes
	// r.Get("/posts", m.oldHandlers.GetAllPosts)
	// r.Get("/users/{id}/posts", m.oldHandlers.GetUserPosts)
	// r.Get("/user/{id}/posts", m.oldHandlers.GetUserPosts)
	r.Post("/recipes/{id}/view", m.oldHandlers.IncrementRecipeView)

	// Protected routes (user recipes)
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware)
		r.Post("/recipes", m.oldHandlers.CreateRecipe)
		r.Put("/recipes/{id}", m.oldHandlers.UpdateRecipe)
		r.Delete("/recipes/{id}", m.oldHandlers.DeleteRecipe)
	})

	// === NEW CATALOG RECIPE ROUTES ===
	// Recipe catalog statistics (public - no sensitive data, only counts)
	r.Get("/recipes/stats", m.catalogHandler.GetRecipeStats)
	
	// Recipe listing with filters (public for browsing catalog)
	r.Get("/recipes", m.catalogHandler.ListRecipes)
	
	// TODO: Remove public access after testing - these should be protected
	// Recipe matching (finds recipes based on fridge) - TEMPORARILY PUBLIC FOR TESTING
	r.Get("/recipes/match", m.catalogHandler.MatchRecipes)
	// Recipe recommendation (returns 1 best recipe for UI) - TEMPORARILY PUBLIC FOR TESTING
	r.Post("/recipes/recommendations", m.catalogHandler.GetRecommendation)

	// Recipe detail by ID (public with optional auth for fridge matching)
	// If user is authenticated, adds inFridge flags to ingredients
	r.With(middleware.OptionalAuthMiddleware).Get("/recipes/{id}", m.catalogHandler.GetRecipeByID)

	// Protected routes (require auth)
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware)
		
		// User saved recipes - NOW PROTECTED
		r.Post("/user/recipes/save", m.catalogHandler.SaveRecipe)
		r.Get("/user/recipes/saved", m.catalogHandler.GetSavedRecipes)
		
		// Recipe cooking (deducts from fridge)
		r.Post("/recipes/{id}/cook", m.catalogHandler.CookRecipe)
		
		// Recipe adaptation (AI adapts recipe to available ingredients)
		r.Post("/recipes/{id}/adapt", m.catalogHandler.AdaptRecipe)
	})
}
