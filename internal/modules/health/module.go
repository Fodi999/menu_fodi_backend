package health

import (
	"net/http"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/health/service"
	transporthttp "github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/health/transport/http"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

// Module представляет Health модуль
type Module struct {
	handlers *transporthttp.HealthHandlers
}

// NewModule создает новый Health модуль
func NewModule(db *gorm.DB) *Module {
	healthService := service.NewHealthService(db)
	handlers := transporthttp.NewHealthHandlers(healthService)

	return &Module{
		handlers: handlers,
	}
}

// RegisterRoutes регистрирует роуты модуля
func (m *Module) RegisterRoutes(r chi.Router, _ func(next http.Handler) http.Handler) {
	// Public route - health check не требует авторизации
	r.Get("/health", m.handlers.HealthCheck)
}
