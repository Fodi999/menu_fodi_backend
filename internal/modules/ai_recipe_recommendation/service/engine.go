package service

import (
	"gorm.io/gorm"
)

// ============================================================================
// RECIPE RECOMMENDATION ENGINE (2025 Architecture)
// Принцип: Rules Engine решает, AI только объясняет
// ============================================================================

// RecommendationEngine - главный сервис для рекомендаций рецептов
type RecommendationEngine struct {
	matcher        *RecipeMatcher
	decisionEngine *DecisionEngine
	// explainer будет добавлен позже для AI объяснений
}

// NewRecommendationEngine - конструктор
func NewRecommendationEngine(db *gorm.DB) *RecommendationEngine {
	matcher := NewRecipeMatcher(db)
	decisionEngine := NewDecisionEngine(matcher)

	return &RecommendationEngine{
		matcher:        matcher,
		decisionEngine: decisionEngine,
	}
}

// GetRecommendations - получает рекомендации рецептов для пользователя
// Это главный метод для вызова из HTTP handler
func (r *RecommendationEngine) GetRecommendations(req RecipeMatchRequest) (*RecipeMatchResponse, error) {
	// Шаг 1-5: Rules Engine делает всю работу
	response, err := r.decisionEngine.MakeRecommendation(req)
	if err != nil {
		return nil, err
	}

	// TODO: Шаг 6 (опционально): AI добавляет объяснения для top-N рецептов
	// if req.WithAI {
	//     r.explainer.AddExplanations(response.Recipes[:3])
	// }

	return response, nil
}

// ============================================================================
// ПРЕИМУЩЕСТВА ЭТОЙ АРХИТЕКТУРЫ:
// ============================================================================
// 
// ✅ Масштабируемо:
//    - 1M пользователей → rules дешёвые
//    - AI вызывается только для top-N рецептов (опционально)
// 
// ✅ Предсказуемо:
//    - нет «галлюцинаций»
//    - пользователь понимает, почему 67%
//    - прозрачная логика scoring
// 
// ✅ Бизнес-логика под контролем:
//    - ты можешь менять правила matching
//    - ты можешь менять формулу scoring
//    - AI не ломает систему
// 
// ✅ Тестируемо:
//    - каждый компонент можно тестировать отдельно
//    - Rules Engine = детерминированная логика
//    - AI = изолированный слой
// 
// ============================================================================
// КАК ИСПОЛЬЗОВАТЬ:
// ============================================================================
// 
// engine := service.NewRecommendationEngine(db)
// 
// req := service.RecipeMatchRequest{
//     UserID:   "user-123",
//     Language: "ru",
//     Limit:    10,
// }
// 
// response, err := engine.GetRecommendations(req)
// 
// // Response structure:
// // {
// //   "decision": "almost_ready",
// //   "summary": "Почти готово! Не хватает всего нескольких ингредиентов.",
// //   "total_matches": 5,
// //   "recipes": [
// //     {
// //       "id": "scrambled-eggs",
// //       "title": "Яичница",
// //       "match_percent": 67,
// //       "match_status": "almost_ready",
// //       "missing_count": 2,
// //       "missing_ingredients": [...],
// //       "available_ingredients": [...]
// //     }
// //   ]
// // }
// 
// ============================================================================
