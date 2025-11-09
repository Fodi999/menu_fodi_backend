package stats

import (
	"net/http"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/stats/service"
	httphandlers "github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/stats/transport/http"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type Module struct {
	handlers *httphandlers.StatsHandlers
}

func NewModule(db *gorm.DB) *Module {
	statsService := service.NewStatsService(db)
	handlers := httphandlers.NewStatsHandlers(statsService)

	return &Module{
		handlers: handlers,
	}
}

func (m *Module) RegisterRoutes(r chi.Router, jwtMiddleware func(next http.Handler) http.Handler) {
	r.Route("/stats", func(r chi.Router) {
		r.Use(jwtMiddleware)

		r.Get("/", m.handlers.GetAdminStats)
		r.Get("/recent-orders", m.handlers.GetRecentOrders)
	})
}
