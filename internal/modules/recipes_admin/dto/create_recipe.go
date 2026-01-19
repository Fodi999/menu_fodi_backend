package dto

// CreateRecipeRequest - МИНИМАЛЬНЫЙ payload для создания draft рецепта (CMS-черновик)
type CreateRecipeRequest struct {
	// ОБЯЗАТЕЛЬНЫЕ поля (минимум для draft)
	LocalName     string  `json:"localName" binding:"required"`  // Отображаемое имя
	CanonicalName *string `json:"canonicalName,omitempty"`       // Slug (optional)
	Category      string  `json:"category" binding:"required"`   // main, soup, dessert, appetizer, salad
	Difficulty    string  `json:"difficulty" binding:"required"` // easy, medium, hard

	// ОПЦИОНАЛЬНЫЕ поля (можно добавить позже через PATCH)
	Description   string  `json:"description,omitempty"`
	ImageUrl      string  `json:"imageUrl,omitempty"`
	Country       string  `json:"country,omitempty"` // Default: PL
	Region        *string `json:"region,omitempty"`
	TimeMinutes   int     `json:"timeMinutes,omitempty"` // Default: 30
	Servings      int     `json:"servings,omitempty"`    // Default: 1
	PortionWeight *int    `json:"portionWeight,omitempty"`

	// Multi-language support (ОПЦИОНАЛЬНО - если не заполнено, будет автоперевод через AI)
	NamePL        *string   `json:"namePl,omitempty"`
	NameEN        *string   `json:"nameEn,omitempty"`
	NameRU        *string   `json:"nameRu,omitempty"`
	DescriptionPL *string   `json:"descriptionPl,omitempty"`
	DescriptionEN *string   `json:"descriptionEn,omitempty"`
	DescriptionRU *string   `json:"descriptionRu,omitempty"`
	StepsPL       *[]string `json:"stepsPl,omitempty"`
	StepsEN       *[]string `json:"stepsEn,omitempty"`
	StepsRU       *[]string `json:"stepsRu,omitempty"`

	// Nutrition (опционально)
	GrossWeight *int     `json:"grossWeight,omitempty"`
	NetWeight   *int     `json:"netWeight,omitempty"`
	Calories    *int     `json:"calories,omitempty"`
	Protein     *float64 `json:"protein,omitempty"`
	Fats        *float64 `json:"fats,omitempty"`
	Carbs       *float64 `json:"carbs,omitempty"`
}

// CreateRecipeResponse - Response для созданного draft рецепта
type CreateRecipeResponse struct {
	ID            string  `json:"id"`
	LocalName     string  `json:"localName"`               // Display name
	CanonicalName *string `json:"canonicalName,omitempty"` // Slug
	Status        string  `json:"status"`                  // Always "draft"
	Category      string  `json:"category"`
	Difficulty    string  `json:"difficulty"`
	AuthorID      string  `json:"authorId"`
	CreatedAt     string  `json:"createdAt"`
}
