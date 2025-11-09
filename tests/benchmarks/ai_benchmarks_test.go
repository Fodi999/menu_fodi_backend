package benchmarks

import (
	"testing"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/ai/dto"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/ai/service"
)

// BenchmarkAIChefMentorRequest benchmarks Chef Mentor API calls
func BenchmarkAIChefMentorRequest(b *testing.B) {
	svc := service.NewAIService()

	req := dto.ChefMentorRequest{
		Message: "How do I make a perfect risotto?",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = svc.ChefMentor(req)
	}
}

// BenchmarkAIRecipeGeneration benchmarks recipe generation performance
func BenchmarkAIRecipeGeneration(b *testing.B) {
	svc := service.NewAIService()

	req := dto.RecipeGenerationRequest{
		Title:    "Chocolate Cake",
		Language: "en",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = svc.GenerateRecipe(req)
	}
}

// BenchmarkAIMealPlanGeneration benchmarks meal plan generation
func BenchmarkAIMealPlanGeneration(b *testing.B) {
	svc := service.NewAIService()

	req := dto.MealPlanRequest{
		Days:           7,
		TargetCalories: 2000,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = svc.GenerateMealPlan(req, nil, nil)
	}
}

// BenchmarkAIFridgeRecommendations benchmarks fridge recommendations generation
func BenchmarkAIFridgeRecommendations(b *testing.B) {
	svc := service.NewAIService()

	req := dto.FridgeRecommendationsRequest{
		DietaryPreferences: []string{"vegetarian"},
		Cuisine:            "Italian",
		MaxTime:            45,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = svc.GetFridgeRecommendations(req, nil)
	}
}

// BenchmarkAIChefMentorParallel benchmarks concurrent Chef Mentor requests
func BenchmarkAIChefMentorParallel(b *testing.B) {
	svc := service.NewAIService()

	req := dto.ChefMentorRequest{
		Message: "Quick salad recipe",
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = svc.ChefMentor(req)
		}
	})
}

// BenchmarkAIRecipeGenerationParallel benchmarks concurrent recipe generation
func BenchmarkAIRecipeGenerationParallel(b *testing.B) {
	svc := service.NewAIService()

	req := dto.RecipeGenerationRequest{
		Title:    "Pasta Carbonara",
		Language: "en",
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = svc.GenerateRecipe(req)
		}
	})
}

// BenchmarkValidationOnly benchmarks just input validation (without API calls)
func BenchmarkValidationOnly(b *testing.B) {
	tests := []struct {
		name    string
		request interface{}
	}{
		{
			name: "empty message validation",
			request: dto.ChefMentorRequest{
				Message: "",
			},
		},
		{
			name: "empty title validation",
			request: dto.RecipeGenerationRequest{
				Title:    "",
				Language: "en",
			},
		},
		{
			name: "invalid days validation",
			request: dto.MealPlanRequest{
				Days:           0,
				TargetCalories: 2000,
			},
		},
	}

	svc := service.NewAIService()

	for _, test := range tests {
		b.Run(test.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				switch req := test.request.(type) {
				case dto.ChefMentorRequest:
					_, _ = svc.ChefMentor(req)
				case dto.RecipeGenerationRequest:
					_, _ = svc.GenerateRecipe(req)
				case dto.MealPlanRequest:
					_, _ = svc.GenerateMealPlan(req, nil, nil)
				}
			}
		})
	}
}
