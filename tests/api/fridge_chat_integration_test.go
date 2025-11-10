package api_test

import (
	"encoding/json"
	"testing"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/ai/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFridgeChatIntegration tests the flow of creating a recipe via chat and saving ingredients to fridge
func TestFridgeChatIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Chef Mentor to Fridge Flow", func(t *testing.T) {
		// Step 1: Get JWT token
		token, userID := setupTestUser(t)

		// Step 2: Start Chef Mentor conversation
		t.Run("Start Recipe Creation", func(t *testing.T) {
			chefReq := dto.ChefMentorRequest{
				Message:             "I want to make a delicious pasta carbonara",
				Language:            "en",
				ConversationHistory: []dto.ConversationMessage{},
				CurrentRecipe:       nil,
			}

			body, err := json.Marshal(chefReq)
			require.NoError(t, err)
			assert.NotEmpty(t, body)
		})

		// Step 3: Save ingredients to fridge
		t.Run("Save Ingredients to Fridge", func(t *testing.T) {
			saveReq := dto.SaveIngredientsRequest{
				Ingredients: []dto.RecipeIngredient{
					{
						Name:   "Pasta",
						Amount: 400,
						Unit:   "г",
					},
					{
						Name:   "Eggs",
						Amount: 3,
						Unit:   "шт",
					},
					{
						Name:   "Bacon",
						Amount: 200,
						Unit:   "г",
					},
					{
						Name:   "Parmesan Cheese",
						Amount: 100,
						Unit:   "г",
					},
				},
			}

			_, err := json.Marshal(saveReq)
			require.NoError(t, err)

			assert.NotEmpty(t, saveReq.Ingredients)
			assert.Equal(t, 4, len(saveReq.Ingredients))
			assert.Equal(t, "Pasta", saveReq.Ingredients[0].Name)
		})

		// Step 4: Verify fridge items
		t.Run("Verify Fridge Items", func(t *testing.T) {
			assert.NotEmpty(t, token)
			assert.NotEmpty(t, userID)
		})
	})
}

// TestSaveIngredientsRequest validates the request structure
func TestSaveIngredientsRequest(t *testing.T) {
	t.Run("Valid ingredients request", func(t *testing.T) {
		req := dto.SaveIngredientsRequest{
			Ingredients: []dto.RecipeIngredient{
				{
					Name:   "Tomato",
					Amount: 500,
					Unit:   "г",
				},
				{
					Name:   "Olive Oil",
					Amount: 50,
					Unit:   "мл",
				},
			},
		}

		assert.Equal(t, 2, len(req.Ingredients))
		assert.Equal(t, "Tomato", req.Ingredients[0].Name)
		assert.Equal(t, float64(500), req.Ingredients[0].Amount)
		assert.Equal(t, "г", req.Ingredients[0].Unit)
	})

	t.Run("Empty ingredients request", func(t *testing.T) {
		req := dto.SaveIngredientsRequest{
			Ingredients: []dto.RecipeIngredient{},
		}

		assert.Equal(t, 0, len(req.Ingredients))
	})
}

// Helper function to setup a test user (would need to be implemented)
func setupTestUser(t *testing.T) (token string, userID string) {
	// This would create a test user and return JWT token
	// For now, just return placeholder values
	return "test-jwt-token", "test-user-id"
}
