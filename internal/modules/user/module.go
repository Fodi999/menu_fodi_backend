package user

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/database"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/user/repo"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/user/service"
	userhttp "github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/user/transport/http"
)

// Module represents user module
type Module struct {
	handlers *userhttp.UserHandlers
}

// NewModule creates new user module
func NewModule(db *gorm.DB) *Module {
	repository := repo.NewUserRepository(db)
	fridgeRepo := database.NewUserFridgeRepository(db)
	tokenBankRepo := &database.TokenBankRepository{} // No constructor, use direct instantiation

	svc := service.NewUserService(repository, fridgeRepo, tokenBankRepo)
	handlers := userhttp.NewUserHandlers(svc)

	return &Module{
		handlers: handlers,
	}
}

// RegisterRoutes registers user routes
func (m *Module) RegisterRoutes(r chi.Router, jwtMiddleware func(http.Handler) http.Handler) {
	// Register handler function для переиспользования
	registerUserRoutes := func(r chi.Router) {
		// All user routes require authentication
		r.Use(jwtMiddleware)

		// Profile endpoints
		r.Get("/profile", m.handlers.GetProfile)
		r.Put("/profile", m.handlers.UpdateProfile)

		// Progress & Dashboard
		r.Get("/progress", m.handlers.GetProgress)
		r.Get("/dashboard", m.handlers.GetDashboard)

		// Achievements
		r.Get("/achievements", m.handlers.GetAchievements)

		// Wallet
		r.Get("/wallet", m.handlers.GetWallet)
	}

	// Register routes under /user
	r.Route("/user", registerUserRoutes)

	// ALIAS: Register same routes under /users (for frontend compatibility)
	r.Route("/users", registerUserRoutes)
}
