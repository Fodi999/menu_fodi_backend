package models

import (
	"time"
)

// FridgeItemStatus статус продукта в холодильнике
type FridgeItemStatus string

const (
	FridgeItemStatusFresh     FridgeItemStatus = "fresh"
	FridgeItemStatusExpired   FridgeItemStatus = "expired"
	FridgeItemStatusDiscarded FridgeItemStatus = "discarded"
)

// FridgeItem - продукт в холодильнике пользователя
type FridgeItem struct {
	ID           string           `gorm:"primaryKey;column:id" json:"id"`
	UserID       string           `gorm:"column:user_id;not null;index" json:"userId"`
	IngredientID string           `gorm:"column:ingredient_id;not null" json:"ingredientId"`
	Quantity     float64          `gorm:"column:quantity;not null" json:"quantity"`
	Unit         string           `gorm:"column:unit;not null" json:"unit"`
	ExpiresAt    *time.Time       `gorm:"column:expires_at" json:"expiresAt,omitempty"`
	Status       FridgeItemStatus `gorm:"column:status;default:'fresh'" json:"status"`
	DaysLeft     *int             `gorm:"column:days_left" json:"daysLeft,omitempty"`      // Вычисляется на backend
	PriceTotal   float64          `gorm:"column:price_total;default:0" json:"priceTotal"` // Для расчета потерь
	CreatedAt    time.Time        `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt    time.Time        `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`

	// Relations
	Ingredient *Ingredient `gorm:"foreignKey:IngredientID;references:ID" json:"ingredient,omitempty"`
	User       *User       `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
}

func (FridgeItem) TableName() string {
	return "fridge_items"
}
