package dto

// MealPlanRequest represents the request body for meal plan generation
type MealPlanRequest struct {
	Language       string `json:"language"`        // ua, en, ru, pl
	TargetCalories int    `json:"targetCalories"`  // Daily calorie target
	Days           int    `json:"days"`            // Number of days (1-14)
	UseFridge      bool   `json:"useFridge"`       // Filter by available ingredients (optional)
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
