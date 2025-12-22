package prepareddishes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/database"
	httphandlers "github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/prepared_dishes/transport/http"
)

type Module struct {
	handler *httphandlers.PreparedDishesHandler
}

func NewModule(db *gorm.DB) *Module {
	repo := database.NewPreparedDishRepository(db)
	handler := httphandlers.NewPreparedDishesHandler(repo)

	return &Module{handler: handler}
}

func (m *Module) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.Route("/api/prepared-dishes", func(r chi.Router) {
		r.Use(authMiddleware)

		r.Get("/", m.handler.GetPreparedDishes)         // GET /api/prepared-dishes?category=pizza&available=true
		r.Get("/stats", m.handler.GetPreparedDishesStats) // GET /api/prepared-dishes/stats
		r.Post("/{id}/consume", m.handler.ConsumePortion) // POST /api/prepared-dishes/{id}/consume
	})
}
