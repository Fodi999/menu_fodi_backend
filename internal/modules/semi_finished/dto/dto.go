package dto

// CreateSemiFinishedIngredientInput represents input for creating a semi-finished product ingredient
type CreateSemiFinishedIngredientInput struct {
	IngredientID  string  `json:"ingredientId"`
	IngredientName string `json:"ingredientName"`
	Quantity      float64 `json:"quantity"`
	Unit          string  `json:"unit"`
	PricePerUnit  float64 `json:"pricePerUnit"`
	TotalPrice    float64 `json:"totalPrice"`
}

// UpdateSemiFinishedIngredientInput represents input for updating ingredients
type UpdateSemiFinishedIngredientInput struct {
	IngredientID  string  `json:"ingredientId"`
	IngredientName string `json:"ingredientName"`
	Quantity      float64 `json:"quantity"`
	Unit          string  `json:"unit"`
	PricePerUnit  float64 `json:"pricePerUnit"`
	TotalPrice    float64 `json:"totalPrice"`
}

// CreateSemiFinishedRequest represents the request body for creating a semi-finished product
type CreateSemiFinishedRequest struct {
	Name           string                                  `json:"name"`
	Description    string                                  `json:"description"`
	Category       string                                  `json:"category"`
	OutputQuantity float64                                 `json:"outputQuantity"`
	OutputUnit     string                                  `json:"outputUnit"`
	Ingredients    []CreateSemiFinishedIngredientInput     `json:"ingredients"`
}

// UpdateSemiFinishedRequest represents the request body for updating a semi-finished product
type UpdateSemiFinishedRequest struct {
	Name           *string                                 `json:"name"`
	Description    *string                                 `json:"description"`
	Category       *string                                 `json:"category"`
	OutputQuantity *float64                                `json:"outputQuantity"`
	OutputUnit     *string                                 `json:"outputUnit"`
	Ingredients    []UpdateSemiFinishedIngredientInput     `json:"ingredients"`
}
