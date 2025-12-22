package history

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/database"
	httphandlers "github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/history/transport/http"
)

type Module struct {
	handler *httphandlers.HistoryHandler
}

func NewModule(db *gorm.DB) *Module {
	repo := database.NewHistoryRepository(db)
	handler := httphandlers.NewHistoryHandler(repo)

	return &Module{handler: handler}
}

func (m *Module) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.Route("/api/history", func(r chi.Router) {
		r.Use(authMiddleware)

		r.Get("/", m.handler.GetHistory)             // GET /api/history?type=consume&limit=50
		r.Get("/stats", m.handler.GetHistoryStats)   // GET /api/history/stats?start_date=2025-01-01
		r.Get("/recent", m.handler.GetRecentActivity) // GET /api/history/recent?limit=10
	})
}
