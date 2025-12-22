package dto

// RecommendationRequest - запрос на получение рекомендации рецепта
type RecommendationRequest struct {
	Mode             string   `json:"mode" binding:"required"`    // "fridge" - подбор по холодильнику
	Limit            int      `json:"limit,omitempty"`            // default: 5
	ExcludeRecipeIds []string `json:"excludeRecipeIds,omitempty"` // UUIDs рецептов для исключения
}

// RecommendationResponse - response в формате совместимом с текущим UI
type RecommendationResponse struct {
	Success            bool                `json:"success"`
	Data               *RecommendationData `json:"data,omitempty"`
	Message            string              `json:"message,omitempty"`
	Error              string              `json:"error,omitempty"`
	RequiresUserAction bool                `json:"requiresUserAction,omitempty"` // true = показать модальное окно с кнопкой
}

// RecommendationData - данные рекомендации (1 лучший рецепт)
type RecommendationData struct {
	Recipe  RecipeInfo  `json:"recipe"`
	Match   MatchInfo   `json:"match"`
	Economy EconomyInfo `json:"economy"`
}

// RecipeInfo - информация о рецепте (совместимый формат)
type RecipeInfo struct {
	ID            string   `json:"id"`
	CanonicalName string   `json:"canonicalName"`
	LocalName     string   `json:"localName"`
	Country       string   `json:"country"`
	Category      string   `json:"category"`
	Difficulty    string   `json:"difficulty"`
	TimeMinutes   int      `json:"timeMinutes"`
	Servings      int      `json:"servings"`
	Steps         []string `json:"steps"`
	Allergens     []string `json:"allergens,omitempty"`
	DietTags      []string `json:"dietTags,omitempty"`
}

// MatchInfo - информация о матчинге с холодильником
type MatchInfo struct {
	CanCookNow      bool                       `json:"canCookNow"`      // true если все required есть
	MissingRequired []MissingIngredientForBuy  `json:"missingRequired"` // Что нужно докупить
	UsedIngredients []UsedIngredient           `json:"usedIngredients"` // Что используется из холодильника
}

// MissingIngredientForBuy - недостающий ингредиент для покупки
type MissingIngredientForBuy struct {
	IngredientID  string  `json:"ingredientId"`
	Name          string  `json:"name"`
	Quantity      float64 `json:"quantity"`
	Unit          string  `json:"unit"`
	EstimatedCost float64 `json:"estimatedCost"` // PLN
}

// UsedIngredient - используемый ингредиент из холодильника
type UsedIngredient struct {
	IngredientID   string  `json:"ingredientId"`
	Name           string  `json:"name"`
	Quantity       float64 `json:"quantity"`
	Unit           string  `json:"unit"`
	Available      float64 `json:"available"`      // Сколько есть в холодильнике
	IsExpiringSoon bool    `json:"isExpiringSoon"` // Близок к истечению
}

// EconomyInfo - экономическая информация
type EconomyInfo struct {
	UsedFromFridge float64 `json:"usedFromFridge"` // PLN: стоимость использованных продуктов
	Saved          float64 `json:"saved"`          // PLN: сколько сэкономили (= usedFromFridge)
}

// SaveRecipeRequest - запрос на сохранение рецепта
type SaveRecipeRequest struct {
	RecipeID string `json:"recipeId" binding:"required"`
	Servings int    `json:"servings,omitempty"` // default: 2
	Source   string `json:"source,omitempty"`   // default: "fridge"
}
