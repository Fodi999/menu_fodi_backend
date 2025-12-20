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
	Name         string   `json:"name"`                   // Название продукта
	Category     string   `json:"category"`               // protein, vegetable, dairy, etc
	Quantity     float64  `json:"quantity"`               // Количество
	Unit         string   `json:"unit"`                   // g, ml, szt
	DaysLeft     *int     `json:"daysLeft,omitempty"`     // Дней до истечения (nil если нет срока)
	Status       string   `json:"status"`                 // "ok" | "warning" | "critical" | "expired"
	PricePerUnit *float64 `json:"pricePerUnit,omitempty"` // Цена за единицу (для расчёта economy)
	Currency     string   `json:"currency,omitempty"`     // PLN, EUR, USD
}

// CreateRecipeFromFridgeRequest запрос на создание рецепта из холодильника
type CreateRecipeFromFridgeRequest struct {
	Language string `json:"language,omitempty"` // "pl" | "en" | "ru" (default: "pl")
}

// CreateRecipeFromFridgeResponse ответ с созданным рецептом
type CreateRecipeFromFridgeResponse struct {
	Success      bool                `json:"success"`
	Recipe       *RestaurantRecipe   `json:"recipe,omitempty"`
	UsedProducts []UsedProductInfo   `json:"usedProducts,omitempty"` // Какие продукты использованы
	Message      string              `json:"message,omitempty"`      // Сообщение об ошибке или предупреждение
}

// RestaurantRecipe профессиональный ресторанный рецепт
type RestaurantRecipe struct {
	Name               string              `json:"name"`               // Название блюда
	Description        string              `json:"description"`        // Короткий описание
	IngredientsUsed    []RecipeIngredient  `json:"ingredientsUsed"`    // Продукты ИЗ холодильника
	IngredientsMissing []RecipeIngredient  `json:"ingredientsMissing"` // Продукты которые нужно ДОКУПИТЬ
	Steps              []string            `json:"steps"`              // Пошаговая инструкция
	CookingTime        int                 `json:"cookingTime"`        // Время приготовления (минуты)
	ChefTips           []string            `json:"chefTips"`           // Советы шефа
	ExpiryPriority     string              `json:"expiryPriority"`     // "critical" | "warning" | "ok"
	Economy            *RecipeEconomy      `json:"economy,omitempty"`  // Экономическая выгода
}

// RecipeIngredient ингредиент в рецепте с точным количеством
type RecipeIngredient struct {
	Name     string  `json:"name"`     // Название ингредиента
	Quantity float64 `json:"quantity"` // Количество
	Unit     string  `json:"unit"`     // Единица измерения (g, ml, szt)
}

// RecipeEconomy экономическая информация о рецепте
type RecipeEconomy struct {
	UsedFromFridge     bool    `json:"usedFromFridge"`     // Использованы продукты из холодильника
	UsedValue          float64 `json:"usedValue"`          // Стоимость использованных продуктов из холодильника
	EstimatedExtraCost float64 `json:"estimatedExtraCost"` // Примерная стоимость недостающих продуктов (pantry)
	SavedMoney         float64 `json:"savedMoney"`         // Сэкономлено денег (usedValue - extraCost)
	Currency           string  `json:"currency"`           // Валюта (PLN, EUR, USD)
}

// UsedProductInfo информация об использованном продукте с расчётом стоимости
type UsedProductInfo struct {
	Name         string  `json:"name"`
	QuantityUsed float64 `json:"quantityUsed"`
	Unit         string  `json:"unit"`
	PricePerUnit float64 `json:"pricePerUnit"`         // Цена за единицу (PLN/g, PLN/ml)
	UsedCost     float64 `json:"usedCost"`             // Стоимость использованного количества
	Currency     string  `json:"currency"`             // Валюта
	DaysLeft     *int    `json:"daysLeft,omitempty"`
}
