package service

import (
	"testing"

	"github.com/google/uuid"
)

func TestValidateAIResponse(t *testing.T) {
	originalTitle := "Łosoś z ryżem"
	originalIngredients := []RecipeIngredientInput{
		{IngredientID: uuid.New().String(), Quantity: 150, Unit: "g"},
		{IngredientID: uuid.New().String(), Quantity: 100, Unit: "g"},
	}

	t.Run("Valid response - all data preserved", func(t *testing.T) {
		response := &AIRecipeResponse{
			Title:       originalTitle,
			Language:    "pl",
			Description: "Test description",
			Servings:    1,
			TimeMinutes: 25,
			Difficulty:  "easy",
			Calories:    520,
			Steps: []RecipeStepAI{
				{Order: 1, Text: "Step 1", Time: 5},
			},
			Ingredients: []AIRecipeIngredient{
				{IngredientID: originalIngredients[0].IngredientID, Name: "Salmon", Amount: 150, Unit: "g"},
				{IngredientID: originalIngredients[1].IngredientID, Name: "Rice", Amount: 100, Unit: "g"},
			},
		}

		err := validateAIResponse(response, originalTitle, originalIngredients)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	})

	t.Run("Invalid - title changed", func(t *testing.T) {
		response := &AIRecipeResponse{
			Title:       "Different Title",
			Language:    "pl",
			Description: "Test description",
			Servings:    1,
			TimeMinutes: 25,
			Difficulty:  "easy",
			Calories:    520,
			Steps: []RecipeStepAI{
				{Order: 1, Text: "Step 1", Time: 5},
			},
			Ingredients: []AIRecipeIngredient{
				{IngredientID: originalIngredients[0].IngredientID, Name: "Salmon", Amount: 150, Unit: "g"},
				{IngredientID: originalIngredients[1].IngredientID, Name: "Rice", Amount: 100, Unit: "g"},
			},
		}

		err := validateAIResponse(response, originalTitle, originalIngredients)
		if err == nil {
			t.Error("Expected error for changed title, got nil")
		}
	})

	t.Run("Invalid - ingredient count mismatch", func(t *testing.T) {
		response := &AIRecipeResponse{
			Title:       originalTitle,
			Language:    "pl",
			Description: "Test description",
			Servings:    1,
			TimeMinutes: 25,
			Difficulty:  "easy",
			Calories:    520,
			Steps: []RecipeStepAI{
				{Order: 1, Text: "Step 1", Time: 5},
			},
			Ingredients: []AIRecipeIngredient{
				{IngredientID: originalIngredients[0].IngredientID, Name: "Salmon", Amount: 150, Unit: "g"},
				// Missing second ingredient
			},
		}

		err := validateAIResponse(response, originalTitle, originalIngredients)
		if err == nil {
			t.Error("Expected error for ingredient count mismatch, got nil")
		}
	})

	t.Run("Invalid - unknown ingredient ID", func(t *testing.T) {
		response := &AIRecipeResponse{
			Title:       originalTitle,
			Language:    "pl",
			Description: "Test description",
			Servings:    1,
			TimeMinutes: 25,
			Difficulty:  "easy",
			Calories:    520,
			Steps: []RecipeStepAI{
				{Order: 1, Text: "Step 1", Time: 5},
			},
			Ingredients: []AIRecipeIngredient{
				{IngredientID: originalIngredients[0].IngredientID, Name: "Salmon", Amount: 150, Unit: "g"},
				{IngredientID: uuid.New().String(), Name: "Rice", Amount: 100, Unit: "g"}, // Wrong ID
			},
		}

		err := validateAIResponse(response, originalTitle, originalIngredients)
		if err == nil {
			t.Error("Expected error for unknown ingredient ID, got nil")
		}
	})

	t.Run("Invalid - empty ingredient ID", func(t *testing.T) {
		response := &AIRecipeResponse{
			Title:       originalTitle,
			Language:    "pl",
			Description: "Test description",
			Servings:    1,
			TimeMinutes: 25,
			Difficulty:  "easy",
			Calories:    520,
			Steps: []RecipeStepAI{
				{Order: 1, Text: "Step 1", Time: 5},
			},
			Ingredients: []AIRecipeIngredient{
				{IngredientID: originalIngredients[0].IngredientID, Name: "Salmon", Amount: 150, Unit: "g"},
				{IngredientID: "", Name: "Rice", Amount: 100, Unit: "g"}, // Empty ID
			},
		}

		err := validateAIResponse(response, originalTitle, originalIngredients)
		if err == nil {
			t.Error("Expected error for empty ingredient ID, got nil")
		}
	})
}
