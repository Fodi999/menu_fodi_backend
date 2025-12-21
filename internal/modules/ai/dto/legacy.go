package dto

import "time"

// Существующие DTO для старых методов (заглушки)

type ChefMentorRequest struct {
	Message             string        `json:"message"`
	Language            string        `json:"language,omitempty"`
	CurrentRecipe       *RecipeDraft  `json:"currentRecipe,omitempty"`
	ConversationHistory []ChatMessage `json:"conversationHistory,omitempty"`
}

type ChefMentorResponse struct {
	Message          string       `json:"message"`
	Recipe           *RecipeDraft `json:"recipe,omitempty"`
	NextQuestion     string       `json:"nextQuestion,omitempty"`
	IsComplete       bool         `json:"isComplete"`
	SuggestedActions []string     `json:"suggestedActions,omitempty"`
}

type RecipeDraft struct {
	Title       string   `json:"title,omitempty"`
	Ingredients []string `json:"ingredients,omitempty"`
	Steps       []string `json:"steps,omitempty"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type MealPlanRequest struct {
	Days           int    `json:"days"`
	TargetCalories int    `json:"targetCalories"`
	Language       string `json:"language,omitempty"`
	UseFridge      bool   `json:"useFridge,omitempty"`
}

type MealPlanResponse struct {
	Plan          []DayMeal `json:"plan"`
	TotalCalories float64   `json:"totalCalories"`
	AvgPerDay     float64   `json:"avgPerDay"`
	Success       bool      `json:"success"`
}

type DayMeal struct {
	Day           string  `json:"day"`
	Breakfast     string  `json:"breakfast"`
	Lunch         string  `json:"lunch"`
	Dinner        string  `json:"dinner"`
	TotalCalories float64 `json:"totalCalories"`
}

type RecipeGenerationRequest struct {
	Title    string `json:"title"`
	Language string `json:"language,omitempty"`
}

type GeneratedRecipe struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Difficulty  string `json:"difficulty"`
	Time        int    `json:"time"`
	Portions    int    `json:"portions"`
}

type FridgeRecommendationsRequest struct {
	Cuisine string `json:"cuisine,omitempty"`
	MaxTime int    `json:"maxTime,omitempty"`
}

type FridgeRecommendation struct {
	RecipeName      string   `json:"recipeName"`
	Description     string   `json:"description"`
	MatchPercentage int      `json:"matchPercentage"`
	MissingItems    []string `json:"missingItems"`
	PrepTime        int      `json:"prepTime"`
	Difficulty      string   `json:"difficulty"`
}

type AvailableIngredientDTO struct {
	Name      string     `json:"name"`
	Quantity  string     `json:"quantity"`
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
}

func NewAvailableIngredientDTO(name, quantity string, expiresAt *time.Time) AvailableIngredientDTO {
	return AvailableIngredientDTO{
		Name:      name,
		Quantity:  quantity,
		ExpiresAt: expiresAt,
	}
}

type SaveIngredientsRequest struct {
	Ingredients []IngredientToSave `json:"ingredients"`
}

type IngredientToSave struct {
	Name   string  `json:"name"`
	Amount float64 `json:"amount"`
	Unit   string  `json:"unit"`
}
