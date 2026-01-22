package http

import (
	"fmt"
	"net/http"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/middleware"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/ai_recipe_recommendation/service"
	"github.com/dmitrijfomin/menu-fodifood/backend/pkg/utils"
)

// ============================================================================
// HTTP HANDLER для Recipe Recommendation Engine (2025)
// Принцип: тонкий слой между HTTP и бизнес-логикой
// ============================================================================

// RecommendationHandler - HTTP handler для рекомендаций
type RecommendationHandler struct {
	engine *service.RecommendationEngine
}

// NewRecommendationHandler - конструктор
func NewRecommendationHandler(engine *service.RecommendationEngine) *RecommendationHandler {
	return &RecommendationHandler{engine: engine}
}

// GetRecommendations - GET /api/recipes/recommendations
// Query params:
//   - lang: language (pl, en, ru) - default: pl
//   - limit: max recipes to return - default: 10
func (h *RecommendationHandler) GetRecommendations(w http.ResponseWriter, r *http.Request) {
	// Извлекаем userID из контекста (установлен middleware.Auth)
	userIDPtr := middleware.GetUserID(r)
	if userIDPtr == nil {
		utils.RespondError(w, http.StatusUnauthorized, "unauthorized", "user ID not found in context")
		return
	}
	userID := userIDPtr.String()

	// Читаем query параметры
	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = "pl" // default
	}

	limit := 10 // default
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		// Parse limit (простая валидация)
		if limitInt := parseLimit(limitStr); limitInt > 0 {
			limit = limitInt
		}
	}

	// Создаем запрос
	req := service.RecipeMatchRequest{
		UserID:   userID,
		Language: lang,
		Limit:    limit,
	}

	// Вызываем Rules Engine
	response, err := h.engine.GetRecommendations(req)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "failed to get recommendations", err.Error())
		return
	}

	// Отправляем ответ
	utils.RespondJSON(w, http.StatusOK, response)
}

// parseLimit - парсит limit из строки
func parseLimit(s string) int {
	var limit int
	if _, err := fmt.Sscanf(s, "%d", &limit); err == nil {
		if limit > 0 && limit <= 50 { // Max 50 recipes
			return limit
		}
	}
	return 10 // default
}

// ============================================================================
// API CONTRACT:
// ============================================================================
//
// GET /api/recipes/recommendations?lang=ru&limit=10
//
// Headers:
//   Authorization: Bearer <jwt_token>
//
// Response 200 OK:
// {
//   "decision": "almost_ready",
//   "summary": "Почти готово! Не хватает всего нескольких ингредиентов.",
//   "total_matches": 5,
//   "recipes": [
//     {
//       "id": "uuid",
//       "canonical_name": "scrambled_eggs",
//       "title": "Яичница",
//       "match_percent": 67.0,
//       "match_status": "almost_ready",
//       "missing_count": 2,
//       "available_count": 4,
//       "total_required": 6,
//       "missing_ingredients": [
//         {
//           "id": "uuid",
//           "canonical_name": "egg",
//           "display_name": "Яйцо",
//           "quantity": 3,
//           "unit": "pcs",
//           "category": "egg"
//         }
//       ],
//       "available_ingredients": [...],
//       "cook_time": 30,
//       "portions": 2,
//       "image_url": "https://..."
//     }
//   ]
// }
//
// ============================================================================
