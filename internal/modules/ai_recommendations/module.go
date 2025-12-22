package ai_recommendations

import (
	"net/http"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/ai_recommendations/service"
	httpTransport "github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/ai_recommendations/transport/http"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

// Module - AI Recommendations модуль
type Module struct {
	db      *gorm.DB
	service *service.RecommendationService
	handler *httpTransport.AIRecommendationsHandler
}

// NewModule - конструктор модуля
func NewModule(db *gorm.DB) *Module {
	// Инициализируем сервис
	recommendationService := service.NewRecommendationService(db)

	// Инициализируем handler
	handler := httpTransport.NewAIRecommendationsHandler(recommendationService)

	return &Module{
		db:      db,
		service: recommendationService,
		handler: handler,
	}
}

// RegisterRoutes - регистрация роутов AI модуля
func (m *Module) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.Route("/api/ai", func(r chi.Router) {
		r.Use(authMiddleware) // Защищаем все AI endpoints авторизацией

		// GET /api/ai/recommendations - главный endpoint
		r.Get("/recommendations", m.handler.GetRecommendations)
	})
}
