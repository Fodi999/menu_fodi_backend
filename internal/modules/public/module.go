package public

import (
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/admin/service"
	"github.com/go-chi/chi/v5"
)

// Module - публичный модуль для SEO endpoints
type Module struct {
	handlers *PublicRecipeHandlers
}

// NewModule - создает новый публичный модуль
func NewModule() *Module {
	// Используем существующий AdminService (для доступа к рецептам)
	adminService := service.NewAdminService()

	return &Module{
		handlers: NewPublicRecipeHandlers(adminService),
	}
}

// RegisterRoutes - регистрирует публичные маршруты (без auth)
func (m *Module) RegisterRoutes(r chi.Router) {
	r.Route("/api/public", func(r chi.Router) {
		// Публичный каталог рецептов (SEO-friendly)
		r.Get("/recipes", m.handlers.GetPublicRecipes)

		// Отдельный рецепт по canonical name (SEO URL)
		r.Get("/recipes/{slug}", m.handlers.GetRecipeBySlug)
	})
}
