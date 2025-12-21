package dto

import "time"

// AdaptRecipeRequest - запрос на адаптацию рецепта под холодильник
type AdaptRecipeRequest struct {
	RecipeID          string                 `json:"recipeId" binding:"required"`
	FridgeSnapshot    []FridgeIngredient     `json:"fridgeSnapshot" binding:"required"`    // Текущее состояние холодильника
	MissingIngredients []string              `json:"missingIngredients,omitempty"`         // Что не хватает
	UserPreferences   *AdaptationPreferences `json:"userPreferences,omitempty"`            // Опциональные предпочтения
	Language          string                 `json:"language,omitempty"`                   // "pl", "en", default: "pl"
}

// FridgeIngredient - ингредиент из холодильника для адаптации
type FridgeIngredient struct {
	IngredientID   string     `json:"ingredientId"`
	Name           string     `json:"name"`
	Quantity       float64    `json:"quantity"`
	Unit           string     `json:"unit"`
	IsExpiringSoon bool       `json:"isExpiringSoon,omitempty"`
	ExpiresAt      *time.Time `json:"expiresAt,omitempty"`
}

// AdaptationPreferences - пользовательские предпочтения для адаптации
type AdaptationPreferences struct {
	ReduceServings    *int     `json:"reduceServings,omitempty"`    // Уменьшить порции (если мало ингредиентов)
	AllowSubstitutions bool    `json:"allowSubstitutions"`          // Разрешить замену ингредиентов
	PreferExpiring    bool     `json:"preferExpiring"`              // Приоритет продуктам с истекающим сроком
	AvoidAllergens    []string `json:"avoidAllergens,omitempty"`    // Избегать аллергенов
	SimplifySteps     bool     `json:"simplifySteps"`               // Упростить шаги (для новичков)
}

// AdaptRecipeResponse - результат адаптации рецепта
type AdaptRecipeResponse struct {
	Success bool               `json:"success"`
	Data    *AdaptedRecipeData `json:"data,omitempty"`
	Message string             `json:"message,omitempty"`
	Error   string             `json:"error,omitempty"`
}

// AdaptedRecipeData - адаптированный рецепт
type AdaptedRecipeData struct {
	// Original recipe info (unchanged)
	OriginalRecipeID   string `json:"originalRecipeId"`
	OriginalName       string `json:"originalName"`
	OriginalServings   int    `json:"originalServings"`
	
	// Adapted recipe (modified by AI)
	AdaptedName        string       `json:"adaptedName"`        // Может быть изменено AI (например, "Carbonara z kurczakiem")
	AdaptedServings    int          `json:"adaptedServings"`    // Может быть уменьшено
	AdaptedSteps       []RecipeStep `json:"adaptedSteps"`       // Переписанные шаги
	AdaptedIngredients []AdaptedIngredient `json:"adaptedIngredients"` // Ингредиенты с заменами
	
	// Adaptation summary
	Adaptations        []Adaptation `json:"adaptations"`        // Что было изменено
	CanCookNow         bool         `json:"canCookNow"`         // Можно ли готовить после адаптации
	DifficultyChange   string       `json:"difficultyChange"`   // "easier", "same", "harder"
	TimeChange         int          `json:"timeChange"`         // Изменение времени (минуты, может быть отрицательным)
	
	// Metadata
	AdaptedAt          time.Time    `json:"adaptedAt"`
}

// AdaptedIngredient - ингредиент с возможной заменой
type AdaptedIngredient struct {
	OriginalName  string  `json:"originalName"`            // Оригинальный ингредиент
	SubstitutedWith *string `json:"substitutedWith,omitempty"` // Замена (если была)
	Quantity      float64 `json:"quantity"`                // Количество (может быть изменено)
	Unit          string  `json:"unit"`
	IsAvailable   bool    `json:"isAvailable"`             // Есть ли в холодильнике
	Reason        string  `json:"reason,omitempty"`        // Почему заменили
}

// Adaptation - описание одной адаптации
type Adaptation struct {
	Type        string `json:"type"`        // "substitution", "portion_reduced", "step_simplified", "ingredient_removed"
	Description string `json:"description"` // Человеко-читаемое описание
	Impact      string `json:"impact"`      // "minor", "moderate", "major"
}

// SaveAdaptedRecipeRequest - сохранить адаптированный рецепт для повторного использования
type SaveAdaptedRecipeRequest struct {
	AdaptedRecipeData AdaptedRecipeData `json:"adaptedRecipe" binding:"required"`
	SaveAs            string            `json:"saveAs,omitempty"` // Custom name
}

// SaveAdaptedRecipeResponse - результат сохранения
type SaveAdaptedRecipeResponse struct {
	Success        bool   `json:"success"`
	SavedRecipeID  string `json:"savedRecipeId,omitempty"`
	Message        string `json:"message,omitempty"`
	Error          string `json:"error,omitempty"`
}
