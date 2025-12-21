package dto

// FridgeAnalysisResponse структурированный ответ от AI (JSON)
type FridgeAnalysisResponse struct {
	Type string      `json:"type"` // "recipe", "meal_plan", "waste_analysis", "budget_analysis"
	Data interface{} `json:"data"` // Конкретная структура в зависимости от type
}

// RecipeResponse ответ для today_meals
type RecipeResponse struct {
	Type              string           `json:"type"`               // "recipe"
	Title             string           `json:"title"`              // "Smażona wołowina z cebulą"
	Portions          int              `json:"portions"`           // 2
	IngredientsUsed   []UsedIngredient `json:"ingredients_used"`   // Список использованных продуктов
	Steps             []string         ` json:"steps"`             // Шаги приготовления
	CookingTime       int              `json:"cooking_time"`       // Минуты
	ExpiresPriority   string           `json:"expires_priority"`   // "critical", "warning", "ok"
	CulinaryTechnique string           `json:"culinary_technique"` // "smażenie", "gotowanie", "pieczenie"
}

// UsedIngredient продукт из холодильника использованный в рецепте
type UsedIngredient struct {
	Name     string  `json:"name"`     // "Wołowina (rostbef)"
	Quantity float64 `json:"quantity"` // 400
	Unit     string  `json:"unit"`     // "g"
}

// ThreeDaysPlanResponse ответ для 3_days_plan
type ThreeDaysPlanResponse struct {
	Type string    `json:"type"` // "meal_plan"
	Days []DayPlan `json:"days"` // 3 дня
}

// DayPlan план на один день
type DayPlan struct {
	DayNumber int         `json:"day_number"` // 1, 2, 3
	Meal      MealDetails `json:"meal"`       // Одно блюдо на день
}

// MealDetails детали блюда
type MealDetails struct {
	Title           string           `json:"title"`            // "Kurczak z warzywami"
	IngredientsUsed []UsedIngredient `json:"ingredients_used"` // Продукты
	CookingTime     int              `json:"cooking_time"`     // Минуты
	Instructions    string           `json:"instructions"`     // Краткая инструкция (2-3 шага)
}

// WasteAnalysisResponse ответ для reduce_waste
type WasteAnalysisResponse struct {
	Type            string         `json:"type"`            // "waste_analysis"
	UrgentItems     []ExpiringItem `json:"urgent_items"`    // ≤2 дня (CRITICAL)
	UseSoonItems    []ExpiringItem `json:"use_soon_items"`  // 3-5 дней (WARNING)
	Recommendations []string       `json:"recommendations"` // Общие советы
	PotentialLoss   float64        `json:"potential_loss"`  // Сумма если выбросить (PLN)
	Currency        string         `json:"currency"`        // "PLN"
}

// ExpiringItem продукт с коротким сроком
type ExpiringItem struct {
	Name         string  `json:"name"`          // "Wołowina"
	DaysLeft     int     `json:"days_left"`     // 2
	Quantity     float64 `json:"quantity"`      // 3000
	Unit         string  `json:"unit"`          // "g"
	TotalValue   float64 `json:"total_value"`   // 61.68
	SuggestedUse string  `json:"suggested_use"` // "Przygotuj gulasz lub smaż z cebulą"
}

// BudgetAnalysisResponse ответ для budget_review
type BudgetAnalysisResponse struct {
	Type              string          `json:"type"`               // "budget_analysis"
	TotalValue        float64         `json:"total_value"`        // 88.86
	Currency          string          `json:"currency"`           // "PLN"
	ProductsCount     int             `json:"products_count"`     // 4
	AverageValue      float64         `json:"average_value"`      // 22.21
	MostExpensive     []ExpensiveItem `json:"most_expensive"`     // Топ-3
	CriticalExpensive []ExpensiveItem `json:"critical_expensive"` // Дорогие с коротким сроком
	PotentialLoss     float64         `json:"potential_loss"`     // Сумма риска
	Recommendations   []BudgetAdvice  `json:"recommendations"`    // Советы по экономии
}

// ExpensiveItem дорогой продукт
type ExpensiveItem struct {
	Name       string  `json:"name"`        // "Wołowina (rostbef)"
	TotalValue float64 `json:"total_value"` // 61.68
	DaysLeft   int     `json:"days_left"`   // 2
	Risk       string  `json:"risk"`        // "critical", "warning", "ok"
}

// BudgetAdvice совет по бюджету
type BudgetAdvice struct {
	Type        string `json:"type"`        // "use_expensive", "cheaper_alternative", "waste_prevention"
	Title       string `json:"title"`       // "Wykorzystaj wołowinę dziś"
	Description string `json:"description"` // "Wołowina jest najdroższym produktem..."
	Savings     string `json:"savings"`     // "Oszczędzisz 60 PLN" (optional)
}
