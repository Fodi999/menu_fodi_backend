package models

import "time"

// Ingredient модель ингредиента (соответствует Prisma схеме)
type Ingredient struct {
	ID        string    `gorm:"primaryKey;column:id" json:"id"`
	Name      string    `gorm:"column:name" json:"name"`
	Unit      string    `gorm:"column:unit" json:"unit"` // "g", "ml", "pcs"
	CreatedAt time.Time `gorm:"column:createdAt;autoCreateTime" json:"createdAt"`
}

// TableName указывает имя таблицы для GORM
func (Ingredient) TableName() string {
	return "Ingredient"
}

// StockItem модель складских остатков (соответствует Prisma схеме)
type StockItem struct {
	ID           string    `gorm:"primaryKey;column:id" json:"id"`
	IngredientID string    `gorm:"column:ingredientId" json:"ingredientId"`
	Quantity     float64   `gorm:"column:quantity" json:"quantity"`
	UpdatedAt    time.Time `gorm:"column:updatedAt;autoUpdateTime" json:"updatedAt"`
	Ingredient   *Ingredient `gorm:"foreignKey:IngredientID" json:"ingredient,omitempty"`
}

// TableName указывает имя таблицы для GORM
func (StockItem) TableName() string {
	return "StockItem"
}

// CreateIngredientRequest запрос на создание ингредиента
type CreateIngredientRequest struct {
	Name     string  `json:"name"`
	Unit     string  `json:"unit"`
	Quantity float64 `json:"quantity"`
}

// UpdateIngredientRequest запрос на обновление ингредиента
type UpdateIngredientRequest struct {
	Name     string  `json:"name"`
	Unit     string  `json:"unit"`
	Quantity float64 `json:"quantity"`
}
