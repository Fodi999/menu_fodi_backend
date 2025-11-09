package marketplace

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/marketplace/repo"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/marketplace/service"
	marketplacehttp "github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/marketplace/transport/http"
)

// Module represents marketplace module
type Module struct {
	handlers *marketplacehttp.MarketplaceHandlers
}

// NewModule creates new marketplace module
func NewModule(db *gorm.DB) *Module {
	repository := repo.NewMarketplaceRepository(db)
	svc := service.NewMarketplaceService(repository)
	handlers := marketplacehttp.NewMarketplaceHandlers(svc)

	return &Module{
		handlers: handlers,
	}
}

// RegisterRoutes registers marketplace routes
func (m *Module) RegisterRoutes(r chi.Router, jwtMiddleware func(http.Handler) http.Handler) {
	r.Route("/marketplace", func(r chi.Router) {
		// Public endpoints
		r.Get("/recipes", m.handlers.GetMarketRecipes)
		r.Get("/leaderboard", m.handlers.GetLeaderboard)
		r.Get("/stats/{userId}", m.handlers.GetSellerStats)

		// Protected endpoints (require auth)
		r.Group(func(r chi.Router) {
			r.Use(jwtMiddleware)

			r.Post("/purchase", m.handlers.PurchaseRecipe)
			r.Get("/purchases", m.handlers.GetUserPurchases)
		})
	})
}
