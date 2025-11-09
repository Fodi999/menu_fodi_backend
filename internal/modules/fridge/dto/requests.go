package dto

import (
	"time"

	"github.com/google/uuid"
)

// AddFridgeItemRequest represents request to add fridge item
type AddFridgeItemRequest struct {
	Product  string  `json:"product" validate:"required"`
	Quantity float64 `json:"quantity" validate:"required,gt=0"`
	Unit     string  `json:"unit" validate:"required"`
}

// UpdateFridgeItemRequest represents request to update fridge item
type UpdateFridgeItemRequest struct {
	Quantity  float64 `json:"quantity,omitempty"`
	Available *bool   `json:"available,omitempty"`
}

// FridgeItemResponse represents a fridge item
type FridgeItemResponse struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"userId"`
	Product   string    `json:"product"`
	Quantity  float64   `json:"quantity"`
	Unit      string    `json:"unit"`
	Available bool      `json:"available"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// FridgeListResponse represents list of fridge items
type FridgeListResponse struct {
	Success bool                 `json:"success"`
	Items   []FridgeItemResponse `json:"items"`
	Count   int                  `json:"count"`
}

// FridgeRecommendationsRequest represents request for recommendations
type FridgeRecommendationsRequest struct {
	DietaryPreferences []string `json:"dietaryPreferences,omitempty"`
	Cuisine            string   `json:"cuisine,omitempty"`
	MaxTime            int      `json:"maxTime,omitempty"` // in minutes
}

// RecipeRecommendation represents a recipe recommendation
type RecipeRecommendation struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Description     string   `json:"description"`
	MatchPercentage int      `json:"matchPercentage"`
	MissingItems    []string `json:"missingItems"`
	PrepTime        int      `json:"prepTime"`
	Difficulty      string   `json:"difficulty"`
	ImageURL        string   `json:"imageUrl,omitempty"`
}
