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
	legacyHandler       *httpTransport.AIRecipeHandler
	recommendationHandler *httpTransport.RecommendationHandler
}

// NewModule - конструктор модуля
func NewModule(db *gorm.DB) *Module {
	// Legacy service (для обратной совместимости)
	matchService := service.NewRecipeMatchService(db)
	legacyHandler := httpTransport.NewAIRecipeHandler(db, matchService)

	// NEW: Recommendation Service (2025 Architecture - clean & testable)
	recommendationService := service.NewRecommendationService(db)
	recommendationHandler := httpTransport.NewRecommendationHandler(recommendationService)

	return &Module{
		legacyHandler:       legacyHandler,
		recommendationHandler: recommendationHandler,
	}
}

// RegisterRoutes - регистрация маршрутов
func (m *Module) RegisterRoutes(r chi.Router, authMiddleware func(next http.Handler) http.Handler) {
	r.Route("/ai-recipe", func(r chi.Router) {
		r.Use(authMiddleware)

		// LEGACY: GET /api/ai-recipe/recommendation (для обратной совместимости)
		r.Get("/recommendation", m.legacyHandler.GetRecommendation)
	})

	// NEW: Правильная архитектура 2025
	r.Route("/recipe-recommendations", func(r chi.Router) {
		r.Use(authMiddleware)

		// GET /api/recipe-recommendations?lang=ru&limit=10
		// Rules Engine решает, AI объясняет (опционально)
		r.Get("/", m.recommendationHandler.GetRecommendations)
		
		// GET /api/recipe-recommendations/{id}?lang=ru
		// Один рецепт с проверкой холодильника (inFridge для каждого ингредиента)
		r.Get("/{id}", m.recommendationHandler.GetSingleRecipeWithFridge)
	})
}
