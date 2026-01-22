package service

import "time"

// ============================================================================
// DTO для Recipe Recommendation Engine
// Принцип: Rules Engine решает, AI только объясняет
// ============================================================================

// RecipeMatchRequest - запрос на подбор рецептов
type RecipeMatchRequest struct {
	UserID   string `json:"user_id"`
	Language string `json:"language"` // pl, en, ru
	Limit    int    `json:"limit"`    // top N рецептов (default: 10)
}

// RecipeMatchResponse - ответ системы с рекомендациями
type RecipeMatchResponse struct {
	Decision     string              `json:"decision"`      // "ready" | "almost_ready" | "need_more"
	Summary      string              `json:"summary"`       // Краткое резюме для пользователя
	TotalMatches int                 `json:"total_matches"` // Сколько рецептов нашли
	Recipes      []RecipeMatchResult `json:"recipes"`       // Список рецептов (отсортирован по match_percent)
}

// RecipeMatchResult - результат matching одного рецепта
type RecipeMatchResult struct {
	// Идентификация
	ID            string `json:"id"`
	CanonicalName string `json:"canonical_name"` // stable key (не зависит от языка)
	Title         string `json:"title"`          // локализованное название
	
	// Метрики matching (Rules Engine)
	MatchPercent         float64 `json:"match_percent"`          // 0-100%
	MatchStatus          string  `json:"match_status"`           // "ready" | "almost_ready" | "not_ready"
	MissingCount         int     `json:"missing_count"`          // Сколько ингредиентов не хватает
	AvailableCount       int     `json:"available_count"`        // Сколько есть
	TotalRequired        int     `json:"total_required"`         // Всего нужно
	
	// Ингредиенты (детали)
	MissingIngredients   []IngredientInfo `json:"missing_ingredients"`   // Чего не хватает
	AvailableIngredients []IngredientInfo `json:"available_ingredients"` // Что есть
	
	// Метаданные рецепта
	CookTime  int    `json:"cook_time"`  // минуты
	Portions  int    `json:"portions"`   // порций
	ImageURL  string `json:"image_url"`  // картинка
	
	// AI объяснение (опционально, добавляется после)
	AIExplanation *AIExplanation `json:"ai_explanation,omitempty"`
}

// IngredientInfo - информация об ингредиенте
type IngredientInfo struct {
	ID            string  `json:"id"`
	CanonicalName string  `json:"canonical_name"` // stable key
	DisplayName   string  `json:"display_name"`   // локализованное название
	Quantity      float64 `json:"quantity"`
	Unit          string  `json:"unit"`
	Category      string  `json:"category"` // для группировки в UI
}

// AIExplanation - AI объяснение (добавляется отдельно)
type AIExplanation struct {
	Explanation  string              `json:"explanation"`  // Почему этот рецепт подходит
	Substitutes  []SubstituteOption  `json:"substitutes"`  // Чем заменить недостающие
	GeneratedAt  time.Time           `json:"generated_at"` // Когда сгенерировано
}

// SubstituteOption - вариант замены ингредиента
type SubstituteOption struct {
	OriginalIngredient string   `json:"original_ingredient"` // Что нужно заменить
	Alternatives       []string `json:"alternatives"`        // Чем можно заменить
	Reason             string   `json:"reason"`              // Почему подходит
}

// MatchingDecision - решение системы (enum)
const (
	DecisionReady       = "ready"        // 🟢 Все ингредиенты есть
	DecisionAlmostReady = "almost_ready" // 🟡 Не хватает 1-2 ингредиента
	DecisionNeedMore    = "need_more"    // 🔴 Не хватает 3+ ингредиентов
)

// MatchStatus - статус matching рецепта (enum)
const (
	StatusReady       = "ready"        // 100% match
	StatusAlmostReady = "almost_ready" // 67-99% match
	StatusNotReady    = "not_ready"    // < 67% match
)
