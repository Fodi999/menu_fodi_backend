package models

import "time"

// Product модель продукта (соответствует Prisma схеме)
type Product struct {
	ID          string    `gorm:"primaryKey;column:id" json:"id"`
	Name        string    `gorm:"column:name" json:"name"`
	Description *string   `gorm:"column:description" json:"description,omitempty"`
	Price       float64   `gorm:"column:price;type:decimal(10,2)" json:"price"`
	ImageURL    *string   `gorm:"column:imageUrl" json:"imageUrl,omitempty"`
	Weight      *string   `gorm:"column:weight" json:"weight,omitempty"`
	Category    string    `gorm:"column:category" json:"category"`
	IsVisible   bool      `gorm:"column:isVisible;default:false" json:"isVisible"`
	CreatedAt   time.Time `gorm:"column:createdAt" json:"createdAt"`

	// Связи
	Ingredients  []ProductIngredient   `gorm:"foreignKey:ProductID" json:"ingredients,omitempty"`
	SemiFinished []ProductSemiFinished `gorm:"foreignKey:ProductID" json:"semiFinished,omitempty"`
}

// ProductIngredient связь продукта с ингредиентом
type ProductIngredient struct {
	ID             string  `gorm:"primaryKey;column:id" json:"id"`
	ProductID      string  `gorm:"column:product_id;not null" json:"productId"`
	IngredientID   string  `gorm:"column:ingredient_id;not null" json:"ingredientId"`
	IngredientName string  `gorm:"column:ingredient_name" json:"ingredientName"`
	Quantity       float64 `gorm:"column:quantity;type:decimal(10,3)" json:"quantity"`
	Unit           string  `gorm:"column:unit" json:"unit"`
	PricePerUnit   float64 `gorm:"column:price_per_unit;type:decimal(10,2)" json:"pricePerUnit"`
	TotalPrice     float64 `gorm:"column:total_price;type:decimal(10,2)" json:"totalPrice"`
}

// TableName для ProductIngredient
func (ProductIngredient) TableName() string {
	return "product_ingredients"
}

// ProductSemiFinished связь продукта с полуфабрикатом
type ProductSemiFinished struct {
	ID               string  `gorm:"primaryKey;column:id" json:"id"`
	ProductID        string  `gorm:"column:product_id;not null" json:"productId"`
	SemiFinishedID   string  `gorm:"column:semi_finished_id;not null" json:"semiFinishedId"`
	SemiFinishedName string  `gorm:"column:semi_finished_name" json:"semiFinishedName"`
	Quantity         float64 `gorm:"column:quantity;type:decimal(10,3)" json:"quantity"`
	Unit             string  `gorm:"column:unit" json:"unit"`
	CostPerUnit      float64 `gorm:"column:cost_per_unit;type:decimal(10,2)" json:"costPerUnit"`
	TotalCost        float64 `gorm:"column:total_cost;type:decimal(10,2)" json:"totalCost"`
}

// TableName для ProductSemiFinished
func (ProductSemiFinished) TableName() string {
	return "product_semi_finished"
}

// TableName указывает имя таблицы для GORM
func (Product) TableName() string {
	return "Product"
}

// CreateProductRequest запрос на создание продукта
type CreateProductRequest struct {
	Name         string                     `json:"name" binding:"required"`
	Description  *string                    `json:"description"`
	Price        float64                    `json:"price" binding:"required"`
	ImageURL     *string                    `json:"imageUrl"`
	Weight       *string                    `json:"weight"`
	Category     string                     `json:"category" binding:"required"`
	IsVisible    bool                       `json:"isVisible"`
	Ingredients  []ProductIngredientInput   `json:"ingredients,omitempty"`
	SemiFinished []ProductSemiFinishedInput `json:"semiFinished,omitempty"`
}

// UpdateProductRequest запрос на обновление продукта
type UpdateProductRequest struct {
	Name         string                     `json:"name" binding:"required"`
	Description  *string                    `json:"description"`
	Price        float64                    `json:"price" binding:"required"`
	ImageURL     *string                    `json:"imageUrl"`
	Weight       *string                    `json:"weight"`
	Category     string                     `json:"category" binding:"required"`
	IsVisible    *bool                      `json:"isVisible"`
	Ingredients  []ProductIngredientInput   `json:"ingredients,omitempty"`
	SemiFinished []ProductSemiFinishedInput `json:"semiFinished,omitempty"`
}

// ProductIngredientInput входные данные для ингредиента продукта
type ProductIngredientInput struct {
	IngredientID   string  `json:"ingredientId" binding:"required"`
	IngredientName string  `json:"ingredientName" binding:"required"`
	Quantity       float64 `json:"quantity" binding:"required"`
	Unit           string  `json:"unit" binding:"required"`
	PricePerUnit   float64 `json:"pricePerUnit" binding:"required"`
	TotalPrice     float64 `json:"totalPrice" binding:"required"`
}

// ProductSemiFinishedInput входные данные для полуфабриката продукта
type ProductSemiFinishedInput struct {
	SemiFinishedID   string  `json:"semiFinishedId" binding:"required"`
	SemiFinishedName string  `json:"semiFinishedName" binding:"required"`
	Quantity         float64 `json:"quantity" binding:"required"`
	Unit             string  `json:"unit" binding:"required"`
	CostPerUnit      float64 `json:"costPerUnit" binding:"required"`
	TotalCost        float64 `json:"totalCost" binding:"required"`
}
