package contact

import (
	"net/http"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/contact/service"
	httphandlers "github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/contact/transport/http"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// Module представляет модуль контактов
type Module struct {
	handlers *httphandlers.ContactHandlers
}

// NewModule создает новый модуль
func NewModule(logger *zap.Logger) *Module {
	// Инициализируем сервис
	contactService := service.NewContactService(logger)

	// Инициализируем обработчики
	handlers := httphandlers.NewContactHandlers(contactService, logger)

	return &Module{
		handlers: handlers,
	}
}

// RegisterRoutes регистрирует маршруты модуля
func (m *Module) RegisterRoutes(r chi.Router, _ func(next http.Handler) http.Handler) {
	// Публичный эндпоинт для контактной формы
	r.Post("/contact", m.handlers.SubmitContactForm)
}
