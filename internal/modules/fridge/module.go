package fridge

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/middleware"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/fridge/repo"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/fridge/service"
	fridgehttp "github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/fridge/transport/http"
)

// Module represents fridge module - для HOME_CHEF пользователей
type Module struct {
	handlers *fridgehttp.FridgeHandlers
}

// NewModule creates new fridge module
func NewModule(db *gorm.DB) *Module {
	repository := repo.NewFridgeRepository(db)
	svc := service.NewFridgeService(repository)
	handlers := fridgehttp.NewFridgeHandlers(svc)

	return &Module{
		handlers: handlers,
	}
}

// RegisterRoutes registers fridge routes - только для HOME_CHEF
func (m *Module) RegisterRoutes(r chi.Router, jwtMiddleware func(http.Handler) http.Handler) {
	r.Route("/fridge", func(r chi.Router) {
		// Требуется аутентификация + роль HOME_CHEF
		r.Use(jwtMiddleware)
		r.Use(middleware.RequireRole(models.RoleHomeChef))

		// Fridge item operations
		r.Get("/", m.handlers.GetUserFridge)
		r.Post("/", m.handlers.AddFridgeItem)
		r.Get("/available", m.handlers.GetAvailableItems)

		// Item-specific operations (with ID)
		r.Put("/{id}", m.handlers.UpdateFridgeItem)
		r.Delete("/{id}", m.handlers.DeleteFridgeItem)
	})
}
