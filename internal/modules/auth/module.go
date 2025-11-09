package auth

import (
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/middleware"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/auth/repo"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/auth/service"
	authhttp "github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/auth/transport/http"
	"github.com/go-chi/chi/v5"
)

// Module represents the auth module
type Module struct {
	handler *authhttp.AuthHandler
}

// NewModule creates a new auth module
func NewModule() *Module {
	// Initialize dependencies
	repository := repo.NewAuthRepository()
	svc := service.NewAuthService(repository)
	handler := authhttp.NewAuthHandler(svc)

	return &Module{
		handler: handler,
	}
}

// RegisterRoutes registers all auth routes
func (m *Module) RegisterRoutes(r chi.Router) {
	// Public routes
	r.Post("/auth/register", m.handler.Register)
	r.Post("/auth/login", m.handler.Login)
	r.Post("/auth/verify", m.handler.VerifyToken)

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware)
		r.Get("/auth/me", m.handler.GetCurrentUser)
	})
}
