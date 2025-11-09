package business

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/business/repo"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/business/service"
	httphandlers "github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/business/transport/http"
)

type Module struct {
	handlers *httphandlers.BusinessHandlers
}

func NewModule(db *gorm.DB) *Module {
	repository := repo.NewBusinessRepository(db)
	svc := service.NewBusinessService(repository)
	handlers := httphandlers.NewBusinessHandlers(svc)
	return &Module{handlers: handlers}
}

func (m *Module) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.Route("/businesses", func(r chi.Router) {
		r.Get("/", m.handlers.GetBusinesses)
		r.Get("/{id}", m.handlers.GetBusinessByID)
		r.Get("/{id}/tokens", m.handlers.GetBusinessTokens)

		r.Group(func(r chi.Router) {
			r.Use(authMiddleware)
			r.Post("/", m.handlers.CreateBusiness)
			r.Put("/{id}", m.handlers.UpdateBusiness)
			r.Delete("/{id}", m.handlers.DeleteBusiness)
		})
	})
}
