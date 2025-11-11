package admin

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/admin/service"
	httphandlers "github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/admin/transport/http"
)

type Module struct {
	handlers *httphandlers.AdminHandlers
}

func NewModule() *Module {
	// Создаём экземпляры сервиса и политики
	adminService := service.NewAdminService()
	adminPolicy := service.NewAdminPolicy()

	// Инжектируем зависимости в handlers
	handlers := httphandlers.NewAdminHandlers(adminService, adminPolicy)

	return &Module{handlers: handlers}
}

func (m *Module) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler, adminMiddleware func(http.Handler) http.Handler) {
	r.Route("/admin", func(r chi.Router) {
		r.Use(authMiddleware)
		r.Use(adminMiddleware)

		// Users
		r.Get("/users", m.handlers.GetAllUsers)
		r.Put("/users/{id}", m.handlers.UpdateUser)
		r.Delete("/users/{id}", m.handlers.DeleteUser)
		r.Patch("/users/update-role", m.handlers.UpdateUserRole)

		// Orders
		r.Get("/orders", m.handlers.GetAllOrders)
		r.Get("/orders/recent", m.handlers.GetRecentOrders)
		r.Put("/orders/{id}/status", m.handlers.UpdateOrderStatus)

		// Stats
		r.Get("/stats", m.handlers.GetAdminStats)
	})
}
