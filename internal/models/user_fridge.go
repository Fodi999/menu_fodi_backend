package models

import (
	"time"

	"github.com/google/uuid"
)

// UserFridge represents ingredients/products in user's virtual fridge
type UserFridge struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"userId"`
	Product   string    `gorm:"type:varchar(255);not null" json:"product"`
	Quantity  float64   `gorm:"type:decimal(10,2);not null" json:"quantity"`
	Unit      string    `gorm:"type:varchar(20);not null" json:"unit"` // г, кг, мл, л, шт
	Available bool      `gorm:"default:false" json:"available"`        // true = в наличии, false = использовано

	// Optional metadata
	Category   string     `gorm:"type:varchar(50)" json:"category,omitempty"` // vegetables, meat, dairy, etc.
	ExpiryDate *time.Time `json:"expiryDate,omitempty"`                       // когда истекает срок годности
	AddedAt    time.Time  `gorm:"autoCreateTime" json:"addedAt"`
	UpdatedAt  time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`
}

// TableName sets table name
func (UserFridge) TableName() string {
	return "user_fridge"
}

// FridgeTransaction represents history of fridge changes
type FridgeTransaction struct {
	ID        uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"userId"`
	Product   string     `gorm:"type:varchar(255);not null" json:"product"`
	Quantity  float64    `gorm:"type:decimal(10,2);not null" json:"quantity"` // positive = add, negative = consume
	Unit      string     `gorm:"type:varchar(20);not null" json:"unit"`
	Action    string     `gorm:"type:varchar(50);not null" json:"action"` // "recipe_cooked", "manual_add", "manual_remove", "expired"
	RecipeID  *uuid.UUID `gorm:"type:uuid" json:"recipeId,omitempty"`     // if from recipe
	CreatedAt time.Time  `gorm:"autoCreateTime" json:"createdAt"`
}

// TableName sets table name
func (FridgeTransaction) TableName() string {
	return "fridge_transactions"
}
