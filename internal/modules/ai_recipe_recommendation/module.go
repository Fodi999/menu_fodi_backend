package ai_recipe_recommendation

import (
	"net/http"
	
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/ai_recipe_recommendation/service"
	httpTransport "github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/ai_recipe_recommendation/transport/http"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

// Module - модуль AI рекомендаций рецептов (архитектура 2025)
type Module struct {
	handler *httpTransport.AIRecipeHandler
}

// NewModule - конструктор модуля
func NewModule(db *gorm.DB) *Module {
	matchService := service.NewRecipeMatchService(db)
	handler := httpTransport.NewAIRecipeHandler(db, matchService)

	return &Module{
		handler: handler,
	}
}

// RegisterRoutes - регистрация маршрутов
func (m *Module) RegisterRoutes(r chi.Router, authMiddleware func(next http.Handler) http.Handler) {
	r.Route("/ai-recipe", func(r chi.Router) {
		r.Use(authMiddleware)
		
		// GET /api/ai-recipe/recommendation
		// Главный endpoint: backend решает, AI объясняет
		r.Get("/recommendation", m.handler.GetRecommendation)
	})
}
