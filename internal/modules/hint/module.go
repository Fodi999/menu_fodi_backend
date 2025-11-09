package hint

import (
	"net/http"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/hint/service"
	httphandlers "github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/hint/transport/http"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

// Module представляет модуль подсказок
type Module struct {
	handlers *httphandlers.HintHandlers
}

// NewModule создает новый модуль
func NewModule(db *gorm.DB) *Module {
	// Инициализируем сервис
	hintService := service.NewHintService(db)

	// Инициализируем обработчики
	handlers := httphandlers.NewHintHandlers(hintService)

	return &Module{
		handlers: handlers,
	}
}

// RegisterRoutes регистрирует маршруты модуля
func (m *Module) RegisterRoutes(r chi.Router, jwtMiddleware func(next http.Handler) http.Handler) {
	// Защищенный эндпоинт для подсказок
	r.Group(func(r chi.Router) {
		r.Use(jwtMiddleware)
		r.Post("/hint", m.handlers.GetHint)
	})
}
