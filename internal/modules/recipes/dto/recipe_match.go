package dto

// RecipeMatchRequest - запрос на поиск рецептов
type RecipeMatchRequest struct {
	// Filters
	Country          string   `json:"country,omitempty"`          // "Poland", "Italy", etc.
	Category         string   `json:"category,omitempty"`         // "main", "dessert", "soup", etc.
	Difficulty       string   `json:"difficulty,omitempty"`       // "easy", "medium", "hard"
	MaxTime          int      `json:"maxTime,omitempty"`          // Maximum time in minutes
	ExcludeAllergens []string `json:"excludeAllergens,omitempty"` // ["gluten", "lactose"]
	IncludeDietTags  []string `json:"includeDietTags,omitempty"`  // ["vegetarian", "keto"]
	MinScore         float64  `json:"minScore,omitempty"`         // Minimum match score (0-100), default: 50
	Limit            int      `json:"limit,omitempty"`            // Max results, default: 10
}

// RecipeMatchResponse - стандартный response для /api/recipes/match
type RecipeMatchResponse struct {
	Success bool             `json:"success"`
	Data    *RecipeMatchData `json:"data,omitempty"`
	Message string           `json:"message,omitempty"`
	Error   string           `json:"error,omitempty"`
}

// RecipeMatchData - данные матчинга
type RecipeMatchData struct {
	Recipes []RecipeMatchItem `json:"recipes"`
	Count   int               `json:"count"`
}

// RecipeMatchItem - один рецепт с результатами матчинга
type RecipeMatchItem struct {
	// Recipe identity
	RecipeID      string `json:"recipeId"`
	CanonicalName string `json:"canonicalName"`
	LocalName     string `json:"localName"`
	Country       string `json:"country"`
	Category      string `json:"category"`
	Difficulty    string `json:"difficulty"`
	TimeMinutes   int    `json:"timeMinutes"`
	Servings      int    `json:"servings"`

	// Match results
	Score    float64 `json:"score"`    // 0-100, чем выше тем лучше
	Coverage float64 `json:"coverage"` // 0-1, процент покрытия ингредиентов (matched / required)

	// Ingredients breakdown
	UsedIngredients    []IngredientMatch `json:"usedIngredients"`    // Что используется из холодильника
	MissingIngredients []IngredientMatch `json:"missingIngredients"` // Что нужно докупить

	// Quick decisions
	CanCookNow bool `json:"canCookNow"` // true если все required ингредиенты есть

	// Economy calculations (clear semantics)
	CostToComplete  float64 `json:"costToComplete"`  // PLN: сколько стоит докупить недостающее
	UsedValue       float64 `json:"usedValue"`       // PLN: стоимость используемых ингредиентов из холодильника
	SavedMoney      float64 `json:"savedMoney"`      // PLN: сколько сэкономили используя продукты (= usedValue, UI: "Wartość z lodówki")
	TotalRecipeCost float64 `json:"totalRecipeCost"` // PLN: полная стоимость рецепта (usedValue + costToComplete)
	WasteRiskSaved  float64 `json:"wasteRiskSaved"`  // PLN: стоимость продуктов близких к истечению (предотвращение food waste)

	// Expiry priority
	HasExpiringItems   bool `json:"hasExpiringItems"`   // Есть ли продукты близкие к истечению
	ExpiringItemsCount int  `json:"expiringItemsCount"` // Сколько таких продуктов

	// Allergens and diet
	Allergens []string `json:"allergens,omitempty"` // ["gluten", "lactose"]
	DietTags  []string `json:"dietTags,omitempty"`  // ["vegetarian", "keto"]
}

// IngredientMatch - информация о конкретном ингредиенте
type IngredientMatch struct {
	IngredientID   string  `json:"ingredientId"`
	Name           string  `json:"name"`
	NameEN         string  `json:"name_en,omitempty"` // English name for translations
	NamePL         string  `json:"name_pl,omitempty"` // Polish name for translations
	NameRU         string  `json:"name_ru,omitempty"` // Russian name for translations
	Quantity       float64 `json:"quantity"`
	Unit           string  `json:"unit"`
	Optional       bool    `json:"optional,omitempty"`       // Для missingIngredients
	EstimatedCost  float64 `json:"estimatedCost,omitempty"`  // Для missingIngredients (PLN)
	Available      float64 `json:"available,omitempty"`      // Для usedIngredients (сколько есть)
	IsExpiringSoon bool    `json:"isExpiringSoon,omitempty"` // Для usedIngredients
}

// RecipeDetailRequest - запрос детальной информации о рецепте
type RecipeDetailRequest struct {
	RecipeID string `json:"recipeId" binding:"required"`
}

// RecipeDetailResponse - полная информация о рецепте
type RecipeDetailResponse struct {
	Success bool              `json:"success"`
	Data    *RecipeDetailData `json:"data,omitempty"`
	Message string            `json:"message,omitempty"`
	Error   string            `json:"error,omitempty"`
}

// RecipeDetailData - детальные данные рецепта
type RecipeDetailData struct {
	RecipeID      string `json:"recipeId"`
	CanonicalName string `json:"canonicalName"`
	LocalName     string `json:"localName"`
	Country       string `json:"country"`
	Region        string `json:"region,omitempty"`
	Category      string `json:"category"`
	Difficulty    string `json:"difficulty"`
	TimeMinutes   int    `json:"timeMinutes"`
	Servings      int    `json:"servings"`

	// Cooking instructions
	Steps []RecipeStep `json:"steps"`

	// Ingredients with details
	Ingredients []RecipeIngredientDetail `json:"ingredients"`

	// Nutrition
	NutritionProfile *NutritionProfile `json:"nutritionProfile,omitempty"`

	// Classifications
	Allergens []AllergenInfo `json:"allergens,omitempty"`
	DietTags  []DietTagInfo  `json:"dietTags,omitempty"`

	// Source
	Source *RecipeSource `json:"source,omitempty"`
}

// RecipeStep - один шаг приготовления
type RecipeStep struct {
	Step        int    `json:"step"`
	Instruction string `json:"instruction"`
}

// RecipeIngredientDetail - детальная информация об ингредиенте в рецепте
type RecipeIngredientDetail struct {
	IngredientID string  `json:"ingredientId"`
	Name         string  `json:"name"`
	Quantity     float64 `json:"quantity"`
	Unit         string  `json:"unit"`
	Optional     bool    `json:"optional"`
	SortOrder    int     `json:"sortOrder"`
}

// NutritionProfile - профиль питания
type NutritionProfile struct {
	Type     string `json:"type"`     // "balanced", "high-protein", "low-carb"
	Calories int    `json:"calories"` // kcal per serving
}

// AllergenInfo - информация об аллергене
type AllergenInfo struct {
	Name        string `json:"name"`        // "gluten"
	DisplayName string `json:"displayName"` // "Gluten"
	Icon        string `json:"icon"`        // "🌾"
}

// DietTagInfo - информация о диете
type DietTagInfo struct {
	Name        string `json:"name"`        // "vegetarian"
	DisplayName string `json:"displayName"` // "Vegetarian"
	Description string `json:"description"` // "No meat or fish"
}

// RecipeSource - источник рецепта
type RecipeSource struct {
	Type      string `json:"type"`      // "cookbook", "website", "traditional"
	Reference string `json:"reference"` // URL or title
}
