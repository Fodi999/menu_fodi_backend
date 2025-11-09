package ingredients

import (
	"net/http"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/ingredients/service"
	httphandlers "github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/ingredients/transport/http"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type Module struct {
	handlers *httphandlers.IngredientsHandlers
}

func NewModule(db *gorm.DB) *Module {
	ingredientsService := service.NewIngredientsService(db)
	handlers := httphandlers.NewIngredientsHandlers(ingredientsService)

	return &Module{
		handlers: handlers,
	}
}

func (m *Module) RegisterRoutes(r chi.Router, jwtMiddleware func(next http.Handler) http.Handler) {
	r.Route("/ingredients", func(r chi.Router) {
		r.Use(jwtMiddleware)
		
		r.Get("/", m.handlers.GetAll)
		r.Post("/", m.handlers.Create)
		r.Get("/{id}", m.handlers.GetOne)
		r.Put("/{id}", m.handlers.Update)
		r.Delete("/{id}", m.handlers.Delete)
		r.Get("/{id}/movements", m.handlers.GetStockMovements)
	})
}
