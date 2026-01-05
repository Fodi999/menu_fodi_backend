package meta

import (
	"github.com/go-chi/chi/v5"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/meta/service"
	transportHTTP "github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/meta/transport/http"
)

type Module struct {
	service  service.MetaService
	handlers *transportHTTP.MetaHandlers
}

func NewModule() *Module {
	svc := service.NewMetaService()
	handlers := transportHTTP.NewMetaHandlers(svc)

	return &Module{
		service:  svc,
		handlers: handlers,
	}
}

func (m *Module) RegisterRoutes(r chi.Router) {
	r.Route("/meta", func(r chi.Router) {
		r.Get("/countries", m.handlers.GetCountries)
		r.Get("/cuisines", m.handlers.GetCuisines)
		r.Get("/categories", m.handlers.GetCategories)
		r.Get("/difficulties", m.handlers.GetDifficulties)
	})
}
