package models

import (
	"math"
	"strings"
	"time"
)

// SemiFinished представляет полуфабрикат
type SemiFinished struct {
	ID             string                   `gorm:"column:id;primaryKey" json:"id"`
	Name           string                   `gorm:"column:name" json:"name"`
	Description    *string                  `gorm:"column:description" json:"description,omitempty"`
	OutputQuantity float64                  `gorm:"column:output_quantity" json:"outputQuantity"`
	OutputUnit     string                   `gorm:"column:output_unit" json:"outputUnit"`
	CostPerUnit    float64                  `gorm:"column:cost_per_unit" json:"costPerUnit"`
	TotalCost      float64                  `gorm:"column:total_cost" json:"totalCost"`
	Category       string                   `gorm:"column:category" json:"category"`
	IsVisible      bool                     `gorm:"column:is_visible;default:true" json:"isVisible"`
	IsArchived     bool                     `gorm:"column:is_archived;default:false" json:"isArchived"`
	CreatedAt      time.Time                `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt      time.Time                `gorm:"column:updated_at" json:"updatedAt"`
	DeletedAt      *time.Time               `gorm:"column:deleted_at" json:"deletedAt,omitempty"`
	Ingredients    []SemiFinishedIngredient `gorm:"foreignKey:SemiFinishedID;constraint:OnDelete:CASCADE" json:"ingredients,omitempty"`
}

// TableName указывает имя таблицы для GORM
func (SemiFinished) TableName() string {
	return "semi_finished"
}

// SemiFinishedIngredient представляет ингредиент в составе полуфабриката
type SemiFinishedIngredient struct {
	ID             string  `gorm:"column:id;primaryKey" json:"id"`
	SemiFinishedID string  `gorm:"column:semi_finished_id" json:"semiFinishedId"`
	IngredientID   string  `gorm:"column:ingredient_id" json:"ingredientId"`
	IngredientName string  `gorm:"column:ingredient_name" json:"ingredientName"`
	Quantity       float64 `gorm:"column:quantity" json:"quantity"`
	Unit           string  `gorm:"column:unit" json:"unit"`
	PricePerUnit   float64 `gorm:"column:price_per_unit" json:"pricePerUnit"`
	TotalPrice     float64 `gorm:"column:total_price" json:"totalPrice"`
}

// TableName указывает имя таблицы для GORM
func (SemiFinishedIngredient) TableName() string {
	return "semi_finished_ingredients"
}

// CreateSemiFinishedRequest запрос на создание полуфабриката
type CreateSemiFinishedRequest struct {
	Name           string                        `json:"name"`
	Description    string                        `json:"description"`
	OutputQuantity float64                       `json:"outputQuantity"`
	OutputUnit     string                        `json:"outputUnit"`
	Category       string                        `json:"category"`
	Ingredients    []SemiFinishedIngredientInput `json:"ingredients"`
}

// SemiFinishedIngredientInput входные данные для ингредиента полуфабриката
type SemiFinishedIngredientInput struct {
	IngredientID   string  `json:"ingredientId"`
	IngredientName string  `json:"ingredientName"`
	Quantity       float64 `json:"quantity"`
	Unit           string  `json:"unit"`
	PricePerUnit   float64 `json:"pricePerUnit"`
	TotalPrice     float64 `json:"totalPrice"`
}

// UpdateSemiFinishedRequest запрос на обновление полуфабриката
type UpdateSemiFinishedRequest struct {
	Name           *string                       `json:"name,omitempty"`
	Description    *string                       `json:"description,omitempty"`
	OutputQuantity *float64                      `json:"outputQuantity,omitempty"`
	OutputUnit     *string                       `json:"outputUnit,omitempty"`
	Category       *string                       `json:"category,omitempty"`
	Ingredients    []SemiFinishedIngredientInput `json:"ingredients,omitempty"`
}

// normalizeFloat округляет число до указанного количества знаков
func normalizeFloat(value float64, decimals int) float64 {
	mult := math.Pow(10, float64(decimals))
	return math.Round(value*mult) / mult
}

// Normalize нормализует числовые значения полуфабриката
func (sf *SemiFinished) Normalize() {
	sf.OutputQuantity = normalizeFloat(sf.OutputQuantity, 3)
	sf.CostPerUnit = normalizeFloat(sf.CostPerUnit, 2)
	sf.TotalCost = normalizeFloat(sf.TotalCost, 2)
}

// NormalizeUnit приводит единицу измерения к стандартному виду
func (sfi *SemiFinishedIngredient) NormalizeUnit() {
	unit := strings.ToLower(sfi.Unit)
	switch {
	case strings.Contains(unit, "гр") || strings.Contains(unit, "gram"):
		sfi.Unit = "g"
	case strings.Contains(unit, "кг") || strings.Contains(unit, "kg") || strings.Contains(unit, "килограмм"):
		sfi.Unit = "kg"
	case strings.Contains(unit, "мл") || strings.Contains(unit, "ml") || strings.Contains(unit, "миллилитр"):
		sfi.Unit = "ml"
	case strings.Contains(unit, "литр") || unit == "л" || unit == "l":
		sfi.Unit = "l"
	case strings.Contains(unit, "шт") || strings.Contains(unit, "pcs") || strings.Contains(unit, "штук"):
		sfi.Unit = "pcs"
	}
}

// NormalizeIngredient нормализует все числовые значения ингредиента
func (sfi *SemiFinishedIngredient) NormalizeIngredient() {
	sfi.Quantity = normalizeFloat(sfi.Quantity, 3)
	sfi.PricePerUnit = normalizeFloat(sfi.PricePerUnit, 2)
	sfi.TotalPrice = normalizeFloat(sfi.TotalPrice, 2)
	sfi.NormalizeUnit()
}
