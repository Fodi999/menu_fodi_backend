package models

import (
	"time"
)

// UserSavedRecipe represents a recipe saved by a user
type UserSavedRecipe struct {
	ID       string     `json:"id" db:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID   string     `json:"userId" db:"user_id" gorm:"type:text;not null;index"`
	RecipeID string     `json:"recipeId" db:"recipe_id" gorm:"not null"` // Database column is UUID, but Go uses string
	Servings int        `json:"servings" db:"servings" gorm:"not null;default:2;check:servings > 0"`
	Source   string     `json:"source" db:"source" gorm:"type:text;not null;default:'fridge'"`
	SavedAt  time.Time  `json:"savedAt" db:"saved_at" gorm:"not null;default:now()"`
	CookedAt *time.Time `json:"cookedAt,omitempty" db:"cooked_at" gorm:"type:timestamptz"`

	// Relations (manually loaded, not via GORM preload)
	Recipe *RecipeCatalog `json:"recipe,omitempty" gorm:"-"`
}

// TableName specifies the table name for UserSavedRecipe model
func (UserSavedRecipe) TableName() string {
	return "user_saved_recipes"
}

// SavedRecipeWithMatch extends UserSavedRecipe with matching information
type SavedRecipeWithMatch struct {
	UserSavedRecipe
	CanCookNow bool `json:"canCookNow"`
}
