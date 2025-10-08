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
	ID              string      `gorm:"primaryKey;column:id" json:"id"`
	IngredientID    string      `gorm:"column:ingredientId" json:"ingredientId"`
	Quantity        float64     `gorm:"column:quantity" json:"quantity"`
	UpdatedAt       time.Time   `gorm:"column:updatedAt;autoUpdateTime" json:"updatedAt"`
	BatchNumber     *string     `gorm:"column:batchNumber" json:"batchNumber,omitempty"`
	BruttoWeight    *float64    `gorm:"column:bruttoWeight" json:"bruttoWeight,omitempty"`
	NettoWeight     *float64    `gorm:"column:nettoWeight" json:"nettoWeight,omitempty"`
	WastePercentage *float64    `gorm:"column:wastePercentage" json:"wastePercentage,omitempty"`
	ExpiryDays      *int        `gorm:"column:expiryDays" json:"expiryDays,omitempty"`
	Supplier        *string     `gorm:"column:supplier" json:"supplier,omitempty"`
	Category        *string     `gorm:"column:category" json:"category,omitempty"`
	PriceBrutto     *float64    `gorm:"column:priceBrutto" json:"priceBrutto,omitempty"`
	PriceNetto      *float64    `gorm:"column:priceNetto" json:"priceNetto,omitempty"`
	PricePerUnit    *float64    `gorm:"column:pricePerUnit" json:"pricePerUnit,omitempty"` // Цена за единицу (кг/л/шт)
	Ingredient      *Ingredient `gorm:"foreignKey:IngredientID;references:ID" json:"ingredient,omitempty"`
}

// TableName указывает имя таблицы для GORM
func (StockItem) TableName() string {
	return "StockItem"
}

// StockMovement модель движения товаров на складе
type StockMovement struct {
	ID          string    `gorm:"primaryKey;column:id" json:"id"`
	StockItemID string    `gorm:"column:stockItemId" json:"stockItemId"`
	Type        string    `gorm:"column:type" json:"type"` // "in" (поступление) или "out" (расход)
	Quantity    float64   `gorm:"column:quantity" json:"quantity"`
	PriceBrutto *float64  `gorm:"column:priceBrutto" json:"priceBrutto,omitempty"`
	PriceNetto  *float64  `gorm:"column:priceNetto" json:"priceNetto,omitempty"`
	Note        *string   `gorm:"column:note" json:"note,omitempty"`
	CreatedAt   time.Time `gorm:"column:createdAt;autoCreateTime" json:"createdAt"`
}

// TableName указывает имя таблицы для GORM
func (StockMovement) TableName() string {
	return "StockMovement"
}

// CreateIngredientRequest запрос на создание ингредиента
type CreateIngredientRequest struct {
	Name            string  `json:"name"`
	Unit            string  `json:"unit"`
	Quantity        float64 `json:"quantity"`
	BruttoWeight    float64 `json:"bruttoWeight"`
	NettoWeight     float64 `json:"nettoWeight"`
	WastePercentage float64 `json:"wastePercentage"`
	ExpiryDays      int     `json:"expiryDays"`
	Supplier        string  `json:"supplier"`
	Category        string  `json:"category"`
	PriceBrutto     float64 `json:"priceBrutto"`
	PriceNetto      float64 `json:"priceNetto"`
	PricePerUnit    float64 `json:"pricePerUnit"` // Цена за единицу (кг/л/шт)
}

// UpdateIngredientRequest запрос на обновление ингредиента
type UpdateIngredientRequest struct {
	Name            string  `json:"name"`
	Unit            string  `json:"unit"`
	Quantity        float64 `json:"quantity"`
	BruttoWeight    float64 `json:"bruttoWeight"`
	NettoWeight     float64 `json:"nettoWeight"`
	WastePercentage float64 `json:"wastePercentage"`
	ExpiryDays      int     `json:"expiryDays"`
	Supplier        string  `json:"supplier"`
	Category        string  `json:"category"`
	PriceBrutto     float64 `json:"priceBrutto"`
	PriceNetto      float64 `json:"priceNetto"`
	PricePerUnit    float64 `json:"pricePerUnit"` // Цена за единицу (кг/л/шт)
}

// IngredientResponse DTO для ответа API (плоская структура для frontend)
type IngredientResponse struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Unit            string    `json:"unit"`
	BatchNumber     *string   `json:"batchNumber,omitempty"`
	Category        *string   `json:"category,omitempty"`
	Supplier        *string   `json:"supplier,omitempty"`
	BruttoWeight    *float64  `json:"bruttoWeight,omitempty"`
	NettoWeight     *float64  `json:"nettoWeight,omitempty"`
	WastePercentage *float64  `json:"wastePercentage,omitempty"`
	ExpiryDays      *int      `json:"expiryDays,omitempty"`
	PriceBrutto     *float64  `json:"priceBrutto,omitempty"`
	PriceNetto      *float64  `json:"priceNetto,omitempty"`
	PricePerUnit    *float64  `json:"pricePerUnit,omitempty"` // Цена за единицу (кг/л/шт)
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

// ToResponse преобразует StockItem в IngredientResponse
func (s *StockItem) ToResponse() *IngredientResponse {
	if s.Ingredient == nil {
		return nil
	}
	return &IngredientResponse{
		ID:              s.ID,
		Name:            s.Ingredient.Name,
		Unit:            s.Ingredient.Unit,
		BatchNumber:     s.BatchNumber,
		Category:        s.Category,
		Supplier:        s.Supplier,
		BruttoWeight:    s.BruttoWeight,
		NettoWeight:     s.NettoWeight,
		WastePercentage: s.WastePercentage,
		ExpiryDays:      s.ExpiryDays,
		PriceBrutto:     s.PriceBrutto,
		PriceNetto:      s.PriceNetto,
		PricePerUnit:    s.PricePerUnit, // Добавляем pricePerUnit в ответ
		CreatedAt:       s.Ingredient.CreatedAt,
		UpdatedAt:       s.UpdatedAt,
	}
}
