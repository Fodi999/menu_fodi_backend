package dto

// BatchFridgeUpdateRequest - batch добавление ингредиентов с автоматическим refresh рецептов
type BatchFridgeUpdateRequest struct {
	Items []FridgeItemAdd `json:"items" binding:"required"`
}

// FridgeItemAdd - один ингредиент для добавления
type FridgeItemAdd struct {
	IngredientID string  `json:"ingredientId,omitempty"`
	Name         string  `json:"name" binding:"required"`
	Quantity     float64 `json:"quantity" binding:"required"`
	Unit         string  `json:"unit" binding:"required"`
	ExpiresAt    *string `json:"expiresAt,omitempty"` // ISO 8601 format
}

// BatchFridgeUpdateResponse - результат batch операции с обновленными рецептами
type BatchFridgeUpdateResponse struct {
	Success bool             `json:"success"`
	Data    *BatchUpdateData `json:"data,omitempty"`
	Message string           `json:"message,omitempty"`
	Error   string           `json:"error,omitempty"`
}

// BatchUpdateData - данные после batch update
type BatchUpdateData struct {
	AddedCount     int               `json:"addedCount"`     // Сколько добавлено
	UpdatedRecipes []RecipeMatchItem `json:"updatedRecipes"` // Обновленные рецепты
	RecipesCount   int               `json:"recipesCount"`   // Сколько рецептов в результате
}
