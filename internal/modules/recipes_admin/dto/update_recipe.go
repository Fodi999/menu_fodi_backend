package dto

// UpdateRecipeRequest - PATCH draft рецепта (можно обновлять что угодно пока draft)
type UpdateRecipeRequest struct {
	Title         *string `json:"title,omitempty"`
	CanonicalName *string `json:"canonicalName,omitempty"`
	Description   *string `json:"description,omitempty"`
	ImageUrl      *string `json:"imageUrl,omitempty"`
	Country       *string `json:"country,omitempty"`
	Category      *string `json:"category,omitempty"`
	Difficulty    *string `json:"difficulty,omitempty"`
	TimeMinutes   *int    `json:"timeMinutes,omitempty"`
	Servings      *int    `json:"servings,omitempty"`
	Region        *string `json:"region,omitempty"`
	PortionWeight *int    `json:"portionWeight,omitempty"`

	// Nutrition (optional)
	GrossWeight *int     `json:"grossWeight,omitempty"`
	NetWeight   *int     `json:"netWeight,omitempty"`
	Calories    *int     `json:"calories,omitempty"`
	Protein     *float64 `json:"protein,omitempty"`
	Fats        *float64 `json:"fats,omitempty"`
	Carbs       *float64 `json:"carbs,omitempty"`

	// Ingredients - можно добавлять/редактировать в draft
	Ingredients []UpdateIngredient `json:"ingredients,omitempty"`

	// Steps - можно добавлять/редактировать в draft
	Steps []UpdateStep `json:"steps,omitempty"`

	// Multilingual (для catalog recipes)
	NamePl        *string `json:"namePl,omitempty"`
	NameEn        *string `json:"nameEn,omitempty"`
	NameRu        *string `json:"nameRu,omitempty"`
	DescriptionPl *string `json:"descriptionPl,omitempty"`
	DescriptionEn *string `json:"descriptionEn,omitempty"`
	DescriptionRu *string `json:"descriptionRu,omitempty"`
}

// UpdateIngredient - Ингредиент для draft
type UpdateIngredient struct {
	IngredientID string  `json:"ingredientId"`
	Quantity     float64 `json:"quantity"`
	Unit         string  `json:"unit"`
	Optional     bool    `json:"optional,omitempty"`
	Notes        string  `json:"notes,omitempty"`
}

// UpdateStep - Шаг для draft
type UpdateStep struct {
	Order       int     `json:"order"`
	Description string  `json:"description"`
	Duration    *int    `json:"duration,omitempty"`    // Minutes
	Temperature *int    `json:"temperature,omitempty"` // Celsius
	ImageUrl    *string `json:"imageUrl,omitempty"`
}

// UpdateRecipeResponse - Response после обновления
type UpdateRecipeResponse struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Status    string `json:"status"`
	UpdatedAt string `json:"updatedAt"`
}
