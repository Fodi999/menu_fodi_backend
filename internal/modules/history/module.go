package history

import (
	"log"
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
	log.Println("🔧 Registering history module routes...")
	r.Route("/history", func(r chi.Router) {
		r.Use(authMiddleware)

		r.Get("/", m.handler.GetHistory)              // GET /api/history?type=consume&limit=50
		r.Get("/stats", m.handler.GetHistoryStats)    // GET /api/history/stats?start_date=2025-01-01
		r.Get("/recent", m.handler.GetRecentActivity) // GET /api/history/recent?limit=10
		r.Get("/losses", m.handler.GetFridgeLosses)   // GET /api/history/losses?days=30 - Expired items analytics
	})
	log.Println("✅ History module routes registered:")
	log.Println("   GET /api/history - Event history")
	log.Println("   GET /api/history/stats - Statistics")
	log.Println("   GET /api/history/recent - Recent activity")
	log.Println("   GET /api/history/losses - Expired items analytics")
}
