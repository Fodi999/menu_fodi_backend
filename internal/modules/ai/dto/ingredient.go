package dto

import "time"

// AvailableIngredientDTO - упрощённый DTO для AI слоя
// Не зависит от DB моделей, используется только для передачи данных в AI
type AvailableIngredientDTO struct {
	Name       string     `json:"name"`
	Quantity   string     `json:"quantity"`
	ExpiryDate *time.Time `json:"expiryDate,omitempty"`
}

// FromUserFridgeItem создаёт DTO из модели UserFridgeItem
func NewAvailableIngredientDTO(name, quantity string, expiryDate *time.Time) AvailableIngredientDTO {
	return AvailableIngredientDTO{
		Name:       name,
		Quantity:   quantity,
		ExpiryDate: expiryDate,
	}
}
