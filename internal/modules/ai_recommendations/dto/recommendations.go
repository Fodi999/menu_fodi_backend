package dto

// GetRecommendationsResponse - структура ответа с рекомендациями
type GetRecommendationsResponse struct {
	Success bool                 `json:"success"`
	Data    *RecommendationsData `json:"data,omitempty"`
	Error   string               `json:"error,omitempty"`
}

// RecommendationsData - данные рекомендаций
type RecommendationsData struct {
	Urgent   []UrgentRecommendation `json:"urgent"`   // TYPE 1 - срочные действия
	Budget   []BudgetRecommendation `json:"budget"`   // TYPE 2 - бюджетные предупреждения
	Cook     []CookSuggestion       `json:"cook"`     // TYPE 3 - что приготовить
	Insights []WasteInsight         `json:"insights"` // TYPE 4 - паттерны потерь
}

// UrgentRecommendation - TYPE 1: срочные действия (expires soon)
type UrgentRecommendation struct {
	Type      string  `json:"type"`                // prepared_expiring, fridge_expiring
	Message   string  `json:"message"`             // "Zjedz dziś Rosół — wygasa jutro"
	DishID    *string `json:"dish_id,omitempty"`   // UUID prepared_dish (если prepared_expiring)
	DishName  *string `json:"dish_name,omitempty"` // Название блюда
	ItemID    *string `json:"item_id,omitempty"`   // UUID fridge_item (если fridge_expiring)
	ItemName  *string `json:"item_name,omitempty"` // Название ингредиента
	ExpiresAt string  `json:"expires_at"`          // ISO 8601
	HoursLeft int     `json:"hours_left"`          // Часов до истечения срока
	Score     float64 `json:"score"`               // Urgency score (0-100)
}

// BudgetRecommendation - TYPE 2: бюджетные предупреждения
type BudgetRecommendation struct {
	Type           string  `json:"type"`            // budget_warning, budget_ok, budget_exceeded
	Message        string  `json:"message"`         // "Pozostało tylko 180 PLN na ten tydzień"
	Remaining      float64 `json:"remaining"`       // Сколько осталось PLN
	DaysLeft       int     `json:"days_left"`       // Дней до конца недели
	AvgDailySpend  float64 `json:"avg_daily_spend"` // Средний расход в день
	ProjectedSpend float64 `json:"projected_spend"` // Прогноз расхода до конца недели
	IsOnTrack      bool    `json:"is_on_track"`     // true = укладываемся, false = превысим
	Score          float64 `json:"score"`           // Budget pressure (0-100)
}

// CookSuggestion - TYPE 3: что приготовить
type CookSuggestion struct {
	Type               string   `json:"type"`                          // cook_suggestion
	RecipeID           string   `json:"recipe_id"`                     // UUID рецепта
	RecipeName         string   `json:"recipe_name"`                   // Название рецепта
	Category           string   `json:"category"`                      // Категория (main, soup, salad)
	Reason             string   `json:"reason"`                        // "100% składników w lodówce"
	IngredientMatch    float64  `json:"ingredient_match"`              // 0-100% доступность ингредиентов
	PreferenceScore    float64  `json:"preference_score"`              // 0-100 на основе истории
	FreshnessUrgency   float64  `json:"freshness_urgency"`             // 0-100 использует скоропортящиеся
	MissingCount       int      `json:"missing_count"`                 // Сколько ингредиентов не хватает
	MissingIngredients []string `json:"missing_ingredients,omitempty"` // Список недостающих
	Score              float64  `json:"score"`                         // Итоговый score
}

// WasteInsight - TYPE 4: паттерны потерь (аналитика)
type WasteInsight struct {
	Type       string  `json:"type"`               // waste_insight, portion_insight, category_insight
	Message    string  `json:"message"`            // "Najczęściej marnujesz dania z kategorii main"
	Category   *string `json:"category,omitempty"` // Категория с высоким waste
	WasteRate  float64 `json:"waste_rate"`         // 0-100% процент потерь
	TotalWaste float64 `json:"total_waste"`        // PLN потерь за период
	Period     string  `json:"period"`             // "last_week", "last_month"
	Suggestion string  `json:"suggestion"`         // Рекомендация что сделать
	Score      float64 `json:"score"`              // Важность инсайта (0-100)
}
