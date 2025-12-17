package dto

// FridgeAnalyzeRequest запрос на анализ холодильника через AI
type FridgeAnalyzeRequest struct {
	Goal        string      `json:"goal" binding:"required"`     // "today_meals" | "3_days_plan" | "reduce_waste" | "budget_review"
	Language    string      `json:"language,omitempty"`          // "pl" | "en" | "ru" (default: "pl")
	Preferences Preferences `json:"preferences,omitempty"`
}

// Preferences дополнительные предпочтения пользователя
type Preferences struct {
	Time   string `json:"time,omitempty"`   // "fast" | "normal"
	Budget string `json:"budget,omitempty"` // "low" | "normal"
}

// FridgeItemDTO минимальный DTO для отправки в AI
// ⚠️ ВАЖНО: AI НЕ ВИДИТ ID, user_id, ingredient_id (приватность + безопасность)
type FridgeItemDTO struct {
	Name       string   `json:"name"`                 // Название продукта
	Category   string   `json:"category"`             // protein, vegetable, dairy, etc
	Quantity   float64  `json:"quantity"`             // Количество
	Unit       string   `json:"unit"`                 // g, ml, szt
	DaysLeft   *int     `json:"daysLeft,omitempty"`   // Дней до истечения (nil если нет срока)
	Status     string   `json:"status"`               // "ok" | "warning" | "critical" | "expired"
	TotalPrice *float64 `json:"totalPrice,omitempty"` // Общая стоимость (если известна)
	Currency   string   `json:"currency,omitempty"`   // PLN, EUR, USD
}
