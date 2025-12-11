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
	// PUBLIC SSE ENDPOINT — БЕЗ АВТОРИЗАЦИИ (EventSource не может отправлять headers)
	r.Route("/treasury", func(r chi.Router) {
		r.Get("/stream", m.handlers.StreamTreasury) // SSE stream - публичный доступ
	})

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

		// Dashboard
		r.Get("/dashboard", m.handlers.GetAdminDashboard)

		// Admin Profile
		r.Get("/profile", m.handlers.GetAdminProfile)

		// Token Bank
		r.Get("/token-bank", m.handlers.GetAllTokenBanks)
		r.Get("/token-bank/stats", m.handlers.GetTokenBankStats)
		r.Get("/token-bank/{userID}", m.handlers.GetUserTokenBank)
		r.Post("/token-bank/allocate", m.handlers.AllocateTokens)
		r.Post("/token-bank/revoke", m.handlers.RevokeTokens)
		r.Put("/token-bank/balance", m.handlers.SetTokenBalance)

		// Treasury
		r.Get("/treasury", m.handlers.GetTreasuryInfo)
		r.Get("/treasury/stats", m.handlers.GetTreasuryStats)          // Detailed Treasury statistics
		r.Get("/token-bank/treasury", m.handlers.GetTreasuryBalance) // Simplified endpoint
		r.Post("/treasury/allocate", m.handlers.AllocateFromTreasury)
	})
}