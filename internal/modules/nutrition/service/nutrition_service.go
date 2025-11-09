package service

import (
	"strings"
)

// NutritionService calculates nutritional values for recipes
type NutritionService struct{}

// NewNutritionService creates a new nutrition service
func NewNutritionService() *NutritionService {
	return &NutritionService{}
}

// NutritionData represents nutritional information per 100g
type NutritionData struct {
	Calories float64 `json:"calories"` // kcal
	Protein  float64 `json:"protein"`  // g
	Fats     float64 `json:"fats"`     // g
	Carbs    float64 `json:"carbs"`    // g
	Fiber    float64 `json:"fiber"`    // g
}

// RecipeNutrition represents total nutrition for entire recipe
type RecipeNutrition struct {
	Total       NutritionData `json:"total"`
	PerServing  NutritionData `json:"perServing"`
	Servings    int           `json:"servings"`
	TotalWeight float64       `json:"totalWeight"` // grams
}

// CalculateRecipeNutrition calculates nutrition for a recipe
func (s *NutritionService) CalculateRecipeNutrition(ingredients []Ingredient, servings int) RecipeNutrition {
	total := NutritionData{}
	totalWeight := 0.0

	for _, ing := range ingredients {
		// Get nutrition per 100g
		nutrition := s.GetNutritionData(ing.Name)
		
		// Convert quantity to grams
		grams := s.convertToGrams(ing.Quantity, ing.Unit)
		totalWeight += grams
		
		// Calculate for actual amount (per 100g base)
		multiplier := grams / 100.0
		total.Calories += nutrition.Calories * multiplier
		total.Protein += nutrition.Protein * multiplier
		total.Fats += nutrition.Fats * multiplier
		total.Carbs += nutrition.Carbs * multiplier
		total.Fiber += nutrition.Fiber * multiplier
	}

	// Calculate per serving
	perServing := NutritionData{}
	if servings > 0 {
		perServing.Calories = total.Calories / float64(servings)
		perServing.Protein = total.Protein / float64(servings)
		perServing.Fats = total.Fats / float64(servings)
		perServing.Carbs = total.Carbs / float64(servings)
		perServing.Fiber = total.Fiber / float64(servings)
	}

	return RecipeNutrition{
		Total:       total,
		PerServing:  perServing,
		Servings:    servings,
		TotalWeight: totalWeight,
	}
}

// GetNutritionData returns nutrition facts per 100g for an ingredient
func (s *NutritionService) GetNutritionData(ingredientName string) NutritionData {
	name := strings.ToLower(ingredientName)

	// Comprehensive nutrition database (per 100g)
	nutritionDB := map[string]NutritionData{
		// Грains & Pasta
		"рис":      {Calories: 130, Protein: 2.7, Fats: 0.3, Carbs: 28.0, Fiber: 0.4},
		"макарони": {Calories: 157, Protein: 5.8, Fats: 0.9, Carbs: 30.9, Fiber: 1.8},
		"греч":    {Calories: 123, Protein: 4.5, Fats: 1.6, Carbs: 25.0, Fiber: 2.7},
		"борошно":  {Calories: 364, Protein: 10.3, Fats: 1.0, Carbs: 76.0, Fiber: 2.7},
		"хліб":     {Calories: 265, Protein: 9.0, Fats: 3.2, Carbs: 49.0, Fiber: 2.7},

		// Proteins - Fish
		"вугор":    {Calories: 184, Protein: 18.4, Fats: 11.8, Carbs: 0.0, Fiber: 0.0},
		"лосось":   {Calories: 208, Protein: 20.0, Fats: 13.0, Carbs: 0.0, Fiber: 0.0},
		"тунець":   {Calories: 144, Protein: 23.3, Fats: 4.9, Carbs: 0.0, Fiber: 0.0},
		"риба":     {Calories: 150, Protein: 20.0, Fats: 7.0, Carbs: 0.0, Fiber: 0.0},

		// Proteins - Meat
		"курка":     {Calories: 165, Protein: 31.0, Fats: 3.6, Carbs: 0.0, Fiber: 0.0},
		"свинина":   {Calories: 242, Protein: 16.0, Fats: 21.0, Carbs: 0.0, Fiber: 0.0},
		"яловичина": {Calories: 250, Protein: 26.0, Fats: 15.0, Carbs: 0.0, Fiber: 0.0},
		"говядина":  {Calories: 250, Protein: 26.0, Fats: 15.0, Carbs: 0.0, Fiber: 0.0},
		"м'ясо":     {Calories: 250, Protein: 25.0, Fats: 17.0, Carbs: 0.0, Fiber: 0.0},
		"бекон":     {Calories: 541, Protein: 37.0, Fats: 42.0, Carbs: 1.4, Fiber: 0.0},

		// Proteins - Seafood
		"креветки": {Calories: 99, Protein: 24.0, Fats: 0.3, Carbs: 0.2, Fiber: 0.0},
		"кальмар":  {Calories: 92, Protein: 15.6, Fats: 1.4, Carbs: 3.1, Fiber: 0.0},

		// Vegetables
		"огірок":   {Calories: 15, Protein: 0.8, Fats: 0.1, Carbs: 3.6, Fiber: 0.5},
		"помідор":  {Calories: 18, Protein: 0.9, Fats: 0.2, Carbs: 3.9, Fiber: 1.2},
		"морква":   {Calories: 41, Protein: 0.9, Fats: 0.2, Carbs: 9.6, Fiber: 2.8},
		"цибуля":   {Calories: 40, Protein: 1.1, Fats: 0.1, Carbs: 9.3, Fiber: 1.7},
		"картопля": {Calories: 77, Protein: 2.0, Fats: 0.1, Carbs: 17.0, Fiber: 2.2},
		"капуста":  {Calories: 25, Protein: 1.3, Fats: 0.1, Carbs: 5.8, Fiber: 2.5},
		"буряк":    {Calories: 43, Protein: 1.6, Fats: 0.2, Carbs: 9.6, Fiber: 2.8},
		"перець":   {Calories: 27, Protein: 1.0, Fats: 0.3, Carbs: 6.0, Fiber: 2.1},

		// Fruits
		"авокадо": {Calories: 160, Protein: 2.0, Fats: 14.7, Carbs: 8.5, Fiber: 6.7},
		"яблуко":  {Calories: 52, Protein: 0.3, Fats: 0.2, Carbs: 14.0, Fiber: 2.4},
		"банан":   {Calories: 89, Protein: 1.1, Fats: 0.3, Carbs: 23.0, Fiber: 2.6},

		// Dairy
		"молоко":   {Calories: 60, Protein: 3.2, Fats: 3.3, Carbs: 4.7, Fiber: 0.0},
		"сир":      {Calories: 356, Protein: 24.0, Fats: 29.0, Carbs: 0.5, Fiber: 0.0},
		"пармезан": {Calories: 431, Protein: 38.0, Fats: 29.0, Carbs: 4.1, Fiber: 0.0},
		"йогурт":   {Calories: 59, Protein: 10.0, Fats: 0.4, Carbs: 3.6, Fiber: 0.0},
		"вершки":   {Calories: 340, Protein: 2.2, Fats: 37.0, Carbs: 2.8, Fiber: 0.0},
		"масло":    {Calories: 717, Protein: 0.9, Fats: 81.0, Carbs: 0.7, Fiber: 0.0},

		// Eggs
		"яйця": {Calories: 155, Protein: 13.0, Fats: 11.0, Carbs: 1.1, Fiber: 0.0},

		// Oils & Fats
		"олія":          {Calories: 884, Protein: 0.0, Fats: 100.0, Carbs: 0.0, Fiber: 0.0},
		"оливкова олія": {Calories: 884, Protein: 0.0, Fats: 100.0, Carbs: 0.0, Fiber: 0.0},

		// Seaweed
		"норі": {Calories: 35, Protein: 5.8, Fats: 0.3, Carbs: 5.1, Fiber: 0.3},

		// Condiments
		"соус":   {Calories: 50, Protein: 1.0, Fats: 2.0, Carbs: 7.0, Fiber: 0.5},
		"кетчуп": {Calories: 97, Protein: 1.0, Fats: 0.1, Carbs: 25.0, Fiber: 0.3},
		"майонез": {Calories: 680, Protein: 1.0, Fats: 75.0, Carbs: 2.6, Fiber: 0.0},

		// Liquids
		"вода": {Calories: 0, Protein: 0, Fats: 0, Carbs: 0, Fiber: 0},
	}

	// Try exact match first
	if data, exists := nutritionDB[name]; exists {
		return data
	}

	// Try partial match
	for key, data := range nutritionDB {
		if strings.Contains(name, key) || strings.Contains(key, name) {
			return data
		}
	}

	// Default: medium-calorie ingredient
	return NutritionData{
		Calories: 100,
		Protein:  5.0,
		Fats:     3.0,
		Carbs:    15.0,
		Fiber:    1.0,
	}
}

// convertToGrams converts quantity to grams
func (s *NutritionService) convertToGrams(quantity float64, unit string) float64 {
	unit = strings.ToLower(unit)

	switch unit {
	case "кг", "kg":
		return quantity * 1000
	case "г", "g", "грам", "gram":
		return quantity
	case "мл", "ml":
		return quantity // assume 1ml = 1g for liquids
	case "л", "l":
		return quantity * 1000
	case "шт", "pcs", "piece":
		return quantity * 100 // assume 1 piece = 100g average
	case "ст.л.", "tbsp", "столова ложка":
		return quantity * 15 // 1 tbsp = 15g
	case "ч.л.", "tsp", "чайна ложка":
		return quantity * 5 // 1 tsp = 5g
	default:
		return quantity // assume grams if unknown
	}
}

// Ingredient represents a recipe ingredient for nutrition calculation
type Ingredient struct {
	Name     string
	Quantity float64
	Unit     string
}
