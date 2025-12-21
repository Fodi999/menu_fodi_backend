package models

import (
	"time"

	"github.com/google/uuid"
)

// FridgeTransaction tracks all fridge operations (add, use, remove)
type FridgeTransaction struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    string     `gorm:"type:text;not null;index:idx_fridge_transactions_user_id" json:"userId"`
	Product   string     `gorm:"type:varchar(255);not null" json:"product"`
	Quantity  float64    `gorm:"type:numeric(10,2);not null" json:"quantity"`
	Unit      string     `gorm:"type:varchar(20);not null" json:"unit"`
	Action    string     `gorm:"type:varchar(50);not null" json:"action"` // "add", "use", "remove", "cook"
	RecipeID  *uuid.UUID `gorm:"type:uuid" json:"recipeId,omitempty"`
	CreatedAt time.Time  `gorm:"type:timestamptz" json:"createdAt"`
}

func (FridgeTransaction) TableName() string {
	return "fridge_transactions"
}
