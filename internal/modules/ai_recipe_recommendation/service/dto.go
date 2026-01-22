package service

import "time"

// ============================================================================
// DTO для Recipe Recommendation Engine (Production-Grade)
// Принцип: Rules Engine решает, AI только объясняет
// ============================================================================

// ============================================================================
// REQUEST / RESPONSE (Top Level)
// ============================================================================

// RecipeMatchRequest - запрос на подбор рецептов
type RecipeMatchRequest struct {
	UserID   string `json:"user_id"`
	Language string `json:"language"` // pl, en, ru
	Limit    int    `json:"limit"`    // top N рецептов (default: 10)
	RecipeID string `json:"recipe_id,omitempty"` // Опционально: конкретный рецепт (UUID или canonical_name)
}

// RecipeRecommendationResponse - полный ответ системы (один контракт для frontend)
type RecipeRecommendationResponse struct {
	Decision     string        `json:"decision"`      // "ready" | "almost_ready" | "need_more"
	Summary      string        `json:"summary"`       // Локализованное резюме
	TotalMatches int           `json:"total_matches"` // Количество найденных рецептов
	Recipes      []RecipeDTO   `json:"recipes"`       // Отсортированы по match_percent (DESC)
}

// ============================================================================
// RECIPE DTO (Full Contract)
// ============================================================================

// RecipeDTO - ПОЛНЫЙ контракт рецепта для frontend
// Содержит ВСЁ что нужно для отображения: метаданные, ингредиенты, шаги, matching
type RecipeDTO struct {
	// Identification
	ID            string  `json:"id"`
	Title         string  `json:"title"`          // Локализованное название
	CanonicalName string  `json:"canonical_name"` // Stable key (language-independent)
	ImageURL      *string `json:"image_url"`      // Cloudinary URL (может быть null)
	
	// Recipe Metadata
	CookTime int `json:"cook_time"` // Минуты
	Servings int `json:"servings"`  // Порций (базовое значение для масштабирования)
	
	// Matching Metrics (Rules Engine)
	MatchPercent float64 `json:"match_percent"` // 0-100 (available / total * 100)
	MatchStatus  string  `json:"match_status"`  // "ready" | "almost_ready" | "not_ready"
	
	// Ingredients (Detailed Objects - НЕ строки!)
	AvailableIngredients []IngredientInfo `json:"available_ingredients"` // ✅ Что есть в холодильнике
	MissingIngredients   []IngredientInfo `json:"missing_ingredients"`   // ❌ Что нужно купить
	
	// Cooking Steps (локализованные)
	Steps []string `json:"steps,omitempty"` // ["Разогрейте сковороду", "Добавьте масло", ...]
	
	// AI Explanation (опционально, Phase 2)
	AI *AIBlock `json:"ai,omitempty"`
}

// IngredientInfo - детальная информация об ингредиенте
// Это объект (НЕ строка), чтобы frontend мог:
// - показать units (g, ml, pcs)
// - рассчитать цену
// - масштабировать порции
// - добавить в shopping cart
type IngredientInfo struct {
	ID            string  `json:"id"`
	CanonicalName string  `json:"canonical_name"` // Для внутренней логики
	DisplayName   string  `json:"display_name"`   // Локализованное название
	Quantity      float64 `json:"quantity"`       // 30, 2, 3
	Unit          string  `json:"unit"`           // "ml", "g", "pcs"
	Category      string  `json:"category"`       // "condiment", "vegetable", "protein"
}

// ============================================================================
// AI EXPLANATION (Phase 2 - Optional)
// ============================================================================

// AIBlock - AI объяснение (добавляется ПОСЛЕ Rules Engine)
type AIBlock struct {
	Title        string              `json:"title"`        // "Почти готово!" | "Можете готовить сейчас!"
	Reason       string              `json:"reason"`       // "Не хватает растительного масла"
	Substitutes  []SubstituteOption  `json:"substitutes"`  // Чем заменить недостающие
	GeneratedAt  time.Time           `json:"generated_at"` // Когда сгенерировано
}

// SubstituteOption - вариант замены ингредиента (AI)
type SubstituteOption struct {
	OriginalIngredient string   `json:"original_ingredient"` // "Растительное масло"
	Alternatives       []string `json:"alternatives"`        // ["Сливочное масло", "Топленое масло"]
	Reason             string   `json:"reason"`              // "Можно использовать любое масло для жарки"
}

// ============================================================================
// CONSTANTS (Decision Engine)
// ============================================================================

// MatchingDecision - решение системы (enum)
const (
	DecisionReady       = "ready"        // 🟢 Все ингредиенты есть (missing = 0)
	DecisionAlmostReady = "almost_ready" // 🟡 Не хватает 1-2 ингредиентов
	DecisionNeedMore    = "need_more"    // 🔴 Не хватает 3+ ингредиентов
)

// MatchStatus - статус matching одного рецепта (enum)
const (
	StatusReady       = "ready"        // 🟢 Можно готовить (missing = 0)
	StatusAlmostReady = "almost_ready" // 🟡 Почти готов (missing ≤ 2)
	StatusNotReady    = "not_ready"    // 🔴 Не готов (missing > 2)
)
