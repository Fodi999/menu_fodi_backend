package semi_finished

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/semi_finished/repo"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/semi_finished/service"
	httphandlers "github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/semi_finished/transport/http"
)

// Module represents semi-finished products module
type Module struct {
	handlers *httphandlers.SemiFinishedHandlers
}

// NewModule creates a new semi-finished products module
func NewModule(db *gorm.DB) *Module {
	repository := repo.NewSemiFinishedRepository(db)
	svc := service.NewSemiFinishedService(repository)
	handlers := httphandlers.NewSemiFinishedHandlers(svc)

	return &Module{
		handlers: handlers,
	}
}

// RegisterRoutes registers all semi-finished product routes
func (m *Module) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.Route("/semi-finished", func(r chi.Router) {
		// Public routes
		r.Get("/", m.handlers.GetAllSemiFinished)
		r.Get("/{id}", m.handlers.GetSemiFinishedByID)

		// Protected routes (admin only)
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware)
			// TODO: Add admin middleware check here if needed
			r.Post("/", m.handlers.CreateSemiFinished)
			r.Put("/{id}", m.handlers.UpdateSemiFinished)
			r.Delete("/{id}", m.handlers.DeleteSemiFinished)
		})
	})
}
