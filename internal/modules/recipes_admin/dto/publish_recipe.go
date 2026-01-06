package dto

// PublishRecipeRequest - Publish рецепта с ПОЛНОЙ валидацией
type PublishRecipeRequest struct {
	// Ingredients - ОБЯЗАТЕЛЬНО при publish
	Ingredients []PublishIngredient `json:"ingredients" binding:"required,min=1"`

	// Steps - ОБЯЗАТЕЛЬНО при publish
	Steps []PublishStep `json:"steps" binding:"required,min=1"`

	// Optional:force publish even if validation warnings
	Force bool `json:"force,omitempty"`
}

// PublishIngredient - Ингредиент для публикации
type PublishIngredient struct {
	IngredientID string  `json:"ingredientId" binding:"required"`
	Quantity     float64 `json:"quantity" binding:"required,min=0"`
	Unit         string  `json:"unit" binding:"required"`
	Optional     bool    `json:"optional,omitempty"`
	Notes        string  `json:"notes,omitempty"`
}

// PublishStep - Шаг приготовления для публикации
type PublishStep struct {
	Order       int     `json:"order" binding:"required,min=1"`
	Description string  `json:"description" binding:"required,min=10"`
	Duration    *int    `json:"duration,omitempty"`    // Minutes
	Temperature *int    `json:"temperature,omitempty"` // Celsius
	ImageUrl    *string `json:"imageUrl,omitempty"`
}

// PublishRecipeResponse - Response после публикации
type PublishRecipeResponse struct {
	ID               string   `json:"id"`
	Title            string   `json:"title"`
	Status           string   `json:"status"` // Now "published"
	PublishedAt      string   `json:"publishedAt"`
	IngredientsCount int      `json:"ingredientsCount"`
	StepsCount       int      `json:"stepsCount"`
	Warnings         []string `json:"warnings,omitempty"` // Optional validation warnings
}

// ValidationError - Ошибки валидации при publish
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}
