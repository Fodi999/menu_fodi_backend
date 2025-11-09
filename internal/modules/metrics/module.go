package metrics

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/metrics/service"
	httphandlers "github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/metrics/transport/http"
)

type Module struct {
	handlers *httphandlers.MetricsHandlers
}

func NewModule() *Module {
	svc := service.NewMetricsService()
	handlers := httphandlers.NewMetricsHandlers(svc)
	return &Module{handlers: handlers}
}

func (m *Module) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.Route("/metrics", func(r chi.Router) {
		r.Get("/{businessId}", m.handlers.GetBusinessMetrics)
	})
}
