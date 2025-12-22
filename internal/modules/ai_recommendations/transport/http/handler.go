package http

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/middleware"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/ai_recommendations/dto"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/ai_recommendations/service"
)

// AIRecommendationsHandler - обработчик AI рекомендаций
type AIRecommendationsHandler struct {
	service *service.RecommendationService
}

// NewAIRecommendationsHandler - конструктор
func NewAIRecommendationsHandler(service *service.RecommendationService) *AIRecommendationsHandler {
	return &AIRecommendationsHandler{
		service: service,
	}
}

// GetRecommendations - GET /api/ai/recommendations
// @Summary Получить AI рекомендации
// @Description Возвращает персонализированные рекомендации на основе аналитики данных пользователя
// @Tags AI Recommendations
// @Security BearerAuth
// @Produce json
// @Success 200 {object} dto.GetRecommendationsResponse
// @Failure 401 {object} dto.GetRecommendationsResponse
// @Failure 500 {object} dto.GetRecommendationsResponse
// @Router /api/ai/recommendations [get]
func (h *AIRecommendationsHandler) GetRecommendations(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Получаем user из middleware
	user, ok := r.Context().Value(middleware.UserContextKey).(*models.User)
	if !ok || user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(dto.GetRecommendationsResponse{
			Success: false,
			Error:   "User not authenticated",
		})
		return
	}

	// Вызываем AI engine
	log.Printf("[AI] Generating recommendations for user: %s", user.ID)
	
	recommendations, err := h.service.GetRecommendations(user.ID)
	if err != nil {
		log.Printf("[AI] Error generating recommendations: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(dto.GetRecommendationsResponse{
			Success: false,
			Error:   "Failed to generate recommendations",
		})
		return
	}

	log.Printf("[AI] Generated recommendations: urgent=%d, budget=%d, cook=%d, insights=%d",
		len(recommendations.Urgent),
		len(recommendations.Budget),
		len(recommendations.Cook),
		len(recommendations.Insights),
	)

	// Возвращаем результат
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(dto.GetRecommendationsResponse{
		Success: true,
		Data:    recommendations,
	})
}
