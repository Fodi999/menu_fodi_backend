package app

import (
	"net/http"
	"time"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/middleware"
	// "github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/academy" // DISABLED: Not used in MVP
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/admin"
	aimodule "github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/ai"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/ai_recommendations"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/auth"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/budget"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/business"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/contact"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/fridge"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/health"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/hint"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/history"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/ingredients"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/leaderboard"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/marketplace"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/meal_plan"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/meta"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/metrics"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/nutrition"
	prepareddishes "github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/prepared_dishes"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/recipes"
	// "github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/semi_finished" // DISABLED: Not used in MVP
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/stats"
	// "github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/task" // DISABLED: Not used in MVP
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/user"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/wallet"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/websocket"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/platform/logger"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// setupModularRoutes configures routes using new modular architecture
// This is the NEW way - using DDD modules instead of flat handlers
func (a *App) setupModularRoutes() http.Handler {
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Timeout(60 * time.Second))

	// CORS configuration
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{
			"http://localhost:3000",
			"http://localhost:3001",
			"https://menu-fodi.vercel.app",
			"https://*.vercel.app",
		},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// Initialize modules
	authModule := auth.NewModule()
	walletModule := wallet.NewModule()
	userModule := user.NewModule(a.db)
	fridgeModule := fridge.NewModule(a.db)
	aiModule := aimodule.NewModule(a.db)
	aiRecommendationsModule := ai_recommendations.NewModule(a.db) // ЭТАП 3 - AI Recommendations
	marketplaceModule := marketplace.NewModule(a.db)
	// academyModule := academy.NewModule(a.db) // DISABLED: Not used in MVP
	healthModule := health.NewModule(a.db)
	contactModule := contact.NewModule(logger.Log)
	hintModule := hint.NewModule(a.db)
	historyModule := history.NewModule(a.db) // User activity history and analytics
	ingredientsModule := ingredients.NewModule(a.db)
	leaderboardModule := leaderboard.NewModule(a.db)

	// NEW MODULES
	adminModule := admin.NewModule()
	budgetModule := budget.NewModule(a.db)                 // Weekly budget tracking
	businessModule := business.NewModule(a.db)
	mealPlanModule := meal_plan.NewModule()
	metaModule := meta.NewModule()                          // Metadata (countries, cuisines, categories, difficulties)
	metricsModule := metrics.NewModule()
	nutritionModule := nutrition.NewModule()
	preparedDishesModule := prepareddishes.NewModule(a.db) // Prepared dishes after cooking
	recipesModule := recipes.NewModule(a.db)                // Updated: Pass DB for catalog services
	// semiFinishedModule := semi_finished.NewModule(a.db) // DISABLED: Not used in MVP
	statsModule := stats.NewModule(a.db)
	// taskModule := task.NewModule() // DISABLED: Not used in MVP
	websocketModule := websocket.NewModule() // WebSocket real-time events

	// Register health module early (before /api routes)
	healthModule.RegisterRoutes(r, middleware.AuthMiddleware)

	// Register contact module (public endpoint)
	contactModule.RegisterRoutes(r, middleware.AuthMiddleware)

	// API routes
	r.Route("/api", func(r chi.Router) {
		// === NEW MODULAR APPROACH ===
		// Register auth module routes
		authModule.RegisterRoutes(r)

		// Register wallet module routes
		walletModule.RegisterRoutes(r)

		// Register user module routes
		userModule.RegisterRoutes(r, middleware.AuthMiddleware)

		// Register fridge module routes
		fridgeModule.RegisterRoutes(r, middleware.AuthMiddleware)

		// Register AI module routes
		aiModule.RegisterRoutes(r, middleware.AuthMiddleware)

		// Register AI Recommendations module routes (ЭТАП 3 - Decision Engine)
		aiRecommendationsModule.RegisterRoutes(r, middleware.AuthMiddleware)

		// Register marketplace module routes
		marketplaceModule.RegisterRoutes(r, middleware.AuthMiddleware)

		// DISABLED: Academy module not used in MVP
		// academyModule.RegisterRoutes(r, middleware.AuthMiddleware)

		// Register hint module routes
		hintModule.RegisterRoutes(r, middleware.AuthMiddleware)

		// Register history module routes (activity tracking and analytics)
		historyModule.RegisterRoutes(r, middleware.AuthMiddleware)

		// Register ingredients module routes
		ingredientsModule.RegisterRoutes(r, middleware.AuthMiddleware)

		// Register leaderboard module routes (public)
		leaderboardModule.RegisterRoutes(r, middleware.AuthMiddleware)

		// === NEW MODULAR ROUTES ===
		// Register admin module routes (with admin middleware)
		adminModule.RegisterRoutes(r, middleware.AuthMiddleware, middleware.AdminMiddleware, middleware.SuperAdminMiddleware)

		// Register business module routes
		businessModule.RegisterRoutes(r, middleware.AuthMiddleware)

		// Register budget module routes (weekly food budget tracking)
		budgetModule.RegisterRoutes(r, middleware.AuthMiddleware)

		// Register meal plan module routes
		mealPlanModule.RegisterRoutes(r, middleware.AuthMiddleware)

		// Register meta module routes (public metadata endpoints)
		metaModule.RegisterRoutes(r)

		// Register metrics module routes
		metricsModule.RegisterRoutes(r, middleware.AuthMiddleware)

		// Register nutrition module routes
		nutritionModule.RegisterRoutes(r, middleware.AuthMiddleware)

		// Register prepared dishes module routes (cook result management)
		preparedDishesModule.RegisterRoutes(r, middleware.AuthMiddleware)

		// Register recipes module routes
		recipesModule.RegisterRoutes(r, middleware.AuthMiddleware)

		// DISABLED: Semi-finished module not used in MVP
		// semiFinishedModule.RegisterRoutes(r, middleware.AuthMiddleware)

		// Register stats module routes
		statsModule.RegisterRoutes(r, middleware.AuthMiddleware)

		// DISABLED: Task module (gamification) not used in MVP
		// taskModule.RegisterRoutes(r)
	})

	// WebSocket routes (outside /api, they don't need JSON structure)
	websocketModule.RegisterRoutes(r)

	return r
}
