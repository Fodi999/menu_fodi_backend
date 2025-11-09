package dto

// ChefMentorRequest represents a user message in the conversation
type ChefMentorRequest struct {
	Message             string                `json:"message"`
	Language            string                `json:"language"` // ua, en, ru, pl
	ConversationHistory []ConversationMessage `json:"history,omitempty"`
	CurrentRecipe       *RecipeDraft          `json:"currentRecipe,omitempty"`
}

// ConversationMessage represents one message in the chat
type ConversationMessage struct {
	Role    string `json:"role"` // "user" or "assistant"
	Content string `json:"content"`
}

// RecipeDraft represents the recipe being built step-by-step
type RecipeDraft struct {
	Title        string             `json:"title,omitempty"`
	Description  string             `json:"description,omitempty"`
	Category     string             `json:"category,omitempty"`
	Difficulty   string             `json:"difficulty,omitempty"`
	Time         int                `json:"time,omitempty"`
	Portions     int                `json:"portions,omitempty"`
	Ingredients  []RecipeIngredient `json:"ingredients,omitempty"`
	Steps        []string           `json:"steps,omitempty"`
	GrossWeight  int                `json:"grossWeight,omitempty"`
	NetWeight    int                `json:"netWeight,omitempty"`
	Calories     int                `json:"calories,omitempty"`
	Protein      float64            `json:"protein,omitempty"`
	Fats         float64            `json:"fats,omitempty"`
	Carbs        float64            `json:"carbs,omitempty"`
	Yield        int                `json:"yield,omitempty"`
	Cost         float64            `json:"cost,omitempty"`
	TokensReward int                `json:"tokensReward,omitempty"`
	IsComplete   bool               `json:"isComplete,omitempty"`
}

// RecipeIngredient represents an ingredient
type RecipeIngredient struct {
	Name   string  `json:"name"`
	Amount float64 `json:"amount"`
	Unit   string  `json:"unit"`
	Gross  float64 `json:"gross,omitempty"`
	Net    float64 `json:"net,omitempty"`
}

// ChefMentorResponse represents the assistant's response
type ChefMentorResponse struct {
	Message          string       `json:"message"`
	Recipe           *RecipeDraft `json:"recipe"`
	NextQuestion     string       `json:"nextQuestion"`
	IsComplete       bool         `json:"isComplete"`
	SuggestedActions []string     `json:"suggestedActions,omitempty"`
}

// MealPlanRequest represents the request for meal plan generation
type MealPlanRequest struct {
	Language       string `json:"language"`
	TargetCalories int    `json:"targetCalories"`
	Days           int    `json:"days"`
	UseFridge      bool   `json:"useFridge"`
}

// DayMeal represents meals for a single day
type DayMeal struct {
	Day           string  `json:"day"`
	Breakfast     string  `json:"breakfast"`
	Lunch         string  `json:"lunch"`
	Dinner        string  `json:"dinner"`
	Snack         string  `json:"snack,omitempty"`
	TotalCalories float64 `json:"totalCalories"`
}

// MealPlanResponse represents the complete meal plan
type MealPlanResponse struct {
	Plan          []DayMeal `json:"plan"`
	TotalCalories float64   `json:"totalCalories"`
	AvgPerDay     float64   `json:"avgPerDay"`
	Success       bool      `json:"success"`
}

// RecipeGenerationRequest represents the input for recipe generation
type RecipeGenerationRequest struct {
	Title    string `json:"title"`
	Language string `json:"language"`
}

// GeneratedRecipe represents the AI-generated recipe
type GeneratedRecipe struct {
	Title        string             `json:"title"`
	Description  string             `json:"description"`
	Category     string             `json:"category"`
	Difficulty   string             `json:"difficulty"`
	Time         int                `json:"time"`
	Portions     int                `json:"portions"`
	Ingredients  []RecipeIngredient `json:"ingredients"`
	Steps        []string           `json:"steps"`
	GrossWeight  int                `json:"grossWeight"`
	NetWeight    int                `json:"netWeight"`
	Calories     int                `json:"calories"`
	Protein      float64            `json:"protein"`
	Fats         float64            `json:"fats"`
	Carbs        float64            `json:"carbs"`
	RecipeYield  int                `json:"yield"`
	Cost         float64            `json:"cost"`
	TokensReward int                `json:"tokensReward"`
	ImageUrl     string             `json:"imageUrl,omitempty"`
}

// FridgeRecommendationsRequest represents request for fridge-based recommendations
type FridgeRecommendationsRequest struct {
	DietaryPreferences []string `json:"dietaryPreferences,omitempty"`
	Cuisine            string   `json:"cuisine,omitempty"`
	MaxTime            int      `json:"maxTime,omitempty"`
}

// FridgeRecommendation represents a recipe recommendation based on fridge
type FridgeRecommendation struct {
	RecipeName      string   `json:"recipeName"`
	Description     string   `json:"description"`
	MatchPercentage int      `json:"matchPercentage"`
	MissingItems    []string `json:"missingItems"`
	PrepTime        int      `json:"prepTime"`
	Difficulty      string   `json:"difficulty"`
}
