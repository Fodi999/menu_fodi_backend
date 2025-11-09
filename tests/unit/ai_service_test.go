package unit

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/ai/dto"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/ai/service"
)

// TestChefMentorValidation tests Chef Mentor validation
func TestChefMentorValidation(t *testing.T) {
	tests := []struct {
		name    string
		req     dto.ChefMentorRequest
		wantErr bool
	}{
		{
			name: "valid message",
			req: dto.ChefMentorRequest{
				Message: "How do I make pasta?",
			},
			wantErr: false,
		},
		{
			name: "empty message",
			req: dto.ChefMentorRequest{
				Message: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := service.NewAIService()
			_, err := svc.ChefMentor(tt.req)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestMealPlanValidation tests Meal Plan validation
func TestMealPlanValidation(t *testing.T) {
	tests := []struct {
		name    string
		req     dto.MealPlanRequest
		wantErr bool
	}{
		{
			name: "valid meal plan",
			req: dto.MealPlanRequest{
				Days:           7,
				TargetCalories: 2000,
			},
			wantErr: false,
		},
		{
			name: "invalid days",
			req: dto.MealPlanRequest{
				Days:           0,
				TargetCalories: 2000,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := service.NewAIService()
			userID := uuid.New()
			_, err := svc.GenerateMealPlan(tt.req, &userID, nil)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestRecipeGeneration tests Recipe Generation
func TestRecipeGeneration(t *testing.T) {
	t.Run("valid recipe", func(t *testing.T) {
		svc := service.NewAIService()
		req := dto.RecipeGenerationRequest{
			Title:    "Pasta",
			Language: "en",
		}
		_, err := svc.GenerateRecipe(req)
		assert.NoError(t, err)
	})
}

// TestFridgeRecommendations tests Fridge Recommendations
func TestFridgeRecommendations(t *testing.T) {
	t.Run("with preferences", func(t *testing.T) {
		svc := service.NewAIService()
		req := dto.FridgeRecommendationsRequest{
			DietaryPreferences: []string{"vegetarian"},
			Cuisine:            "Italian",
			MaxTime:            30,
		}
		_, err := svc.GetFridgeRecommendations(req, nil)
		assert.NoError(t, err)
	})

	t.Run("empty request", func(t *testing.T) {
		svc := service.NewAIService()
		req := dto.FridgeRecommendationsRequest{}
		_, err := svc.GetFridgeRecommendations(req, nil)
		assert.NoError(t, err)
	})
}
