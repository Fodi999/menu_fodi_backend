package dto

import "github.com/dmitrijfomin/menu-fodifood/backend/internal/models"

// HintRequest - запрос подсказки
type HintRequest struct {
	Question string `json:"question" binding:"required"`
}

// HintResponse - ответ с подсказкой
type HintResponse struct {
	Hint              string           `json:"hint"`
	SuggestedProducts []models.Product `json:"suggested_products"`
}
