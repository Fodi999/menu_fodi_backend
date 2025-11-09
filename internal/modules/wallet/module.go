package wallet

import (
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/middleware"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/wallet/repo"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/wallet/service"
	wallethttp "github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/wallet/transport/http"
	"github.com/go-chi/chi/v5"
)

// Module represents the wallet module
type Module struct {
	handler *wallethttp.WalletHandler
}

// NewModule creates a new wallet module
func NewModule() *Module {
	// Initialize dependencies
	repository := repo.NewWalletRepository()
	svc := service.NewWalletService(repository)
	handler := wallethttp.NewWalletHandler(svc)

	return &Module{
		handler: handler,
	}
}

// RegisterRoutes registers all wallet routes
func (m *Module) RegisterRoutes(r chi.Router) {
	// Public routes (none for wallet)

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware)

		// Wallet operations
		r.Get("/wallet/balance", m.handler.GetBalance)
		r.Post("/wallet/purchase", m.handler.PurchaseTokens)
		r.Post("/wallet/spend", m.handler.SpendTokens)
		r.Get("/wallet/transactions", m.handler.GetTransactionHistory)

		// User-specific wallet routes
		r.Get("/user/{userId}/wallet", m.handler.GetWalletInfo)
		r.Post("/user/{userId}/wallet/grant-welcome", m.handler.GrantWelcomeTokens)
	})
}
