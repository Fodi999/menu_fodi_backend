package dto

// CookRecipeRequest - запрос на приготовление рецепта
type CookRecipeRequest struct {
	RecipeID           string  `json:"recipeId" binding:"required"`
	ServingsMultiplier float64 `json:"servingsMultiplier"` // Optional: coefficient (2.0 = double recipe)
	TargetServings     int     `json:"targetServings"`     // Optional: absolute portions (2 = cook for 2 people)
	IdempotencyKey     string  `json:"idempotencyKey"`     // Optional, for preventing double-cooking
	Force              bool    `json:"force"`              // Optional, allow re-cooking already cooked recipe
}

// GetMultiplier calculates servingsMultiplier from either field
// Priority: ServingsMultiplier > TargetServings > default (1.0)
func (r *CookRecipeRequest) GetMultiplier(recipeServings int) float64 {
	// If explicit multiplier provided, use it
	if r.ServingsMultiplier > 0 {
		return r.ServingsMultiplier
	}

	// If targetServings provided, calculate multiplier
	// Example: recipe has 4 servings, user wants 2 → multiplier = 2/4 = 0.5
	if r.TargetServings > 0 && recipeServings > 0 {
		return float64(r.TargetServings) / float64(recipeServings)
	}

	// Default: cook 1x recipe
	return 1.0
}

// CookRecipeResponse - результат приготовления
type CookRecipeResponse struct {
	Success bool                `json:"success"`
	Data    *CookRecipeData     `json:"data,omitempty"`
	Message string              `json:"message,omitempty"`
	Error   string              `json:"error,omitempty"`
	Code    string              `json:"code,omitempty"`    // Error code like "INSUFFICIENT_INGREDIENTS"
	Missing []MissingIngredient `json:"missing,omitempty"` // Detailed list of missing ingredients
}

// MissingIngredient - детали недостающего ингредиента
type MissingIngredient struct {
	Name      string  `json:"name"`
	Needed    float64 `json:"needed"`
	Available float64 `json:"available"`
	Unit      string  `json:"unit"`
}

// CookRecipeData - детали приготовления
type CookRecipeData struct {
	CookLogID     string `json:"cookLogId"`
	RecipeID      string `json:"recipeId"`
	CanonicalName string `json:"canonicalName"`
	LocalName     string `json:"localName"`

	// Cooking details
	ServingsMultiplier float64 `json:"servingsMultiplier"`
	CookedAt           string  `json:"cookedAt"` // ISO 8601

	// Economy summary
	UsedValue       float64 `json:"usedValue"`      // PLN saved by using fridge
	WasteRiskSaved  float64 `json:"wasteRiskSaved"` // PLN waste prevented
	TotalRecipeCost float64 `json:"totalRecipeCost"`

	// What was deducted
	IngredientsUsed []CookedIngredient `json:"ingredientsUsed"`

	// Fridge state after cooking
	FridgeUpdated  bool `json:"fridgeUpdated"`
	RemainingItems int  `json:"remainingItems"` // How many fridge items left
}

// CookedIngredient - ингредиент, использованный при приготовлении
type CookedIngredient struct {
	IngredientID      string  `json:"ingredientId"`
	Name              string  `json:"name"`
	QuantityUsed      float64 `json:"quantityUsed"`
	Unit              string  `json:"unit"`
	PricePerUnit      float64 `json:"pricePerUnit,omitempty"`
	TotalCost         float64 `json:"totalCost,omitempty"`
	WasExpiringSoon   bool    `json:"wasExpiringSoon"`
	RemainingInFridge float64 `json:"remainingInFridge"` // How much is left after cooking
}
