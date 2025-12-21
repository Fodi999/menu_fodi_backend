package models

import (
	"time"

	"github.com/lib/pq"
)

// UserRecipeSession tracks recipe recommendation session for a user
type UserRecipeSession struct {
	UserID            string         `json:"userId" db:"user_id" gorm:"type:text;primaryKey"`
	LastRecipeID      *string        `json:"lastRecipeId,omitempty" db:"last_recipe_id" gorm:"type:uuid"`
	ExcludedRecipeIDs pq.StringArray `json:"excludedRecipeIds" db:"excluded_recipe_ids" gorm:"type:uuid[]"`
	UpdatedAt         time.Time      `json:"updatedAt" db:"updated_at" gorm:"not null;default:now()"`
}

// TableName specifies the table name for UserRecipeSession model
func (UserRecipeSession) TableName() string {
	return "user_recipe_sessions"
}
