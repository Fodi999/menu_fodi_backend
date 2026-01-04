package models

import (
	"time"

	"github.com/google/uuid"
)

// RecipeStep represents a single cooking instruction step in a specific language
// Supports multi-language recipes with proper relational structure
type RecipeStep struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	RecipeID    uuid.UUID `gorm:"column:recipe_id;type:uuid;not null;index" json:"recipeId"`
	StepNumber  int       `gorm:"column:step_number;not null" json:"stepNumber"`
	Language    string    `gorm:"type:varchar(5);not null;index;check:language IN ('pl', 'en', 'ru')" json:"language"`
	Instruction string    `gorm:"type:text;not null" json:"instruction"`
	CreatedAt   time.Time `gorm:"column:created_at;not null;default:now()" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"column:updated_at;not null;default:now()" json:"updatedAt"`
}

func (RecipeStep) TableName() string {
	return "RecipeStep"
}

// RecipeStepsByLanguage groups steps by language for easy access
type RecipeStepsByLanguage struct {
	Polish  []string `json:"pl,omitempty"`
	English []string `json:"en,omitempty"`
	Russian []string `json:"ru,omitempty"`
}

// GetStepsForLanguage returns ordered steps for a specific language
func GetStepsForRecipe(steps []RecipeStep, lang string) []string {
	var result []string
	for _, step := range steps {
		if step.Language == lang {
			result = append(result, step.Instruction)
		}
	}
	return result
}
