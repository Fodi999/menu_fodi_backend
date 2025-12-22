package dto

// AddMissingRequest - request to add missing ingredients from recipe to fridge
type AddMissingRequest struct {
	RecipeID string `json:"recipeId" binding:"required"` // Recipe UUID
}

// AddMissingResult - result of adding missing ingredients (NO TEXT, only data)
type AddMissingResult struct {
	Added   int         `json:"added"`   // Number of items added/updated
	Skipped int         `json:"skipped"` // Number of items already sufficient
	Items   []AddedItem `json:"items"`   // Details of added items
}

// AddedItem - details of one added ingredient (NO TEXT, only facts)
type AddedItem struct {
	IngredientID  string  `json:"ingredientId"`  // Ingredient catalog ID
	Name          string  `json:"name"`          // Ingredient name (for display)
	AddedQuantity float64 `json:"addedQuantity"` // How much was added
	Unit          string  `json:"unit"`          // Unit of measurement
}
