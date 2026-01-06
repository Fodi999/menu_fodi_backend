package dto

// AvailableRecipesResponse - структурированный список рецептов по возможности приготовления
// GET /api/recipes/available
type AvailableRecipesResponse struct {
	Success bool                  `json:"success"`
	Data    *AvailableRecipesData `json:"data,omitempty"`
	Message string                `json:"message,omitempty"`
	Error   string                `json:"error,omitempty"`
}

// AvailableRecipesData - категоризированные рецепты
type AvailableRecipesData struct {
	CanCook      []AvailableRecipeItem `json:"canCook"`      // 100% match, можно готовить сейчас
	AlmostCook   []AvailableRecipeItem `json:"almostCook"`   // 67-99% match, не хватает 1-2 ингредиентов
	NeedToBuy    []AvailableRecipeItem `json:"needToBuy"`    // <67% match, нужно докупить
	TotalCount   int                   `json:"totalCount"`   // Общее количество рецептов
	CanCookCount int                   `json:"canCookCount"` // Сколько можно приготовить сейчас
}

// AvailableRecipeItem - упрощенная карточка рецепта для списка
type AvailableRecipeItem struct {
	RecipeID      string `json:"recipeId"`
	CanonicalName string `json:"canonicalName"`
	LocalName     string `json:"localName"`
	Category      string `json:"category"`
	Difficulty    string `json:"difficulty"`
	TimeMinutes   int    `json:"timeMinutes"`
	Servings      int    `json:"servings"`

	// Match info
	Match           int      `json:"match"`           // 0-100 (целое число для UI)
	CanCook         bool     `json:"canCook"`         // true если все required есть
	MissingCount    int      `json:"missingCount"`    // Сколько ингредиентов не хватает
	Missing         []string `json:"missing"`         // Названия недостающих ингредиентов (ru)
	UsedIngredients []string `json:"usedIngredients"` // Названия используемых ингредиентов (ru)

	// Economy (опционально, для сортировки)
	CostToComplete   float64 `json:"costToComplete,omitempty"`   // PLN
	WasteRiskSaved   float64 `json:"wasteRiskSaved,omitempty"`   // PLN
	HasExpiringItems bool    `json:"hasExpiringItems,omitempty"` // Приоритет
}
