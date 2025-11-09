package dto

import "github.com/dmitrijfomin/menu-fodifood/backend/internal/models"

type CreateIngredientRequest struct {
	Name            string  `json:"name" binding:"required"`
	Unit            string  `json:"unit" binding:"required"`
	Quantity        float64 `json:"quantity"`
	BruttoWeight    float64 `json:"bruttoWeight"`
	NettoWeight     float64 `json:"nettoWeight"`
	WastePercentage float64 `json:"wastePercentage"`
	ExpiryDays      int     `json:"expiryDays"`
	Supplier        string  `json:"supplier"`
	Category        string  `json:"category"`
	PriceBrutto     float64 `json:"priceBrutto"`
	PriceNetto      float64 `json:"priceNetto"`
	PricePerUnit    float64 `json:"pricePerUnit"`
}

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
	PricePerUnit    float64 `json:"pricePerUnit"`
}

type IngredientResponse struct {
	*models.StockItem
}

type StockMovementsResponse struct {
	Movements []models.StockMovement `json:"movements"`
}
