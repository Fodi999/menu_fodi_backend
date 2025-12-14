package service

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/ai/dto"
	ai_core "github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/ai_core"
)

var (
	ErrEmptyMessage    = errors.New("message cannot be empty")
	ErrEmptyTitle      = errors.New("title cannot be empty")
	ErrInvalidDays     = errors.New("days must be between 1 and 14")
	ErrInvalidCalories = errors.New("target calories must be positive")
)

// AIService handles AI business logic
type AIService interface {
	// Chef Mentor
	ChefMentor(req dto.ChefMentorRequest) (*dto.ChefMentorResponse, error)

	// Meal Planning - использует DTO вместо моделей
	GenerateMealPlan(req dto.MealPlanRequest, userID *uuid.UUID, fridgeItems []dto.AvailableIngredientDTO) (*dto.MealPlanResponse, error)

	// Recipe Generation
	GenerateRecipe(req dto.RecipeGenerationRequest) (*dto.GeneratedRecipe, error)

	// Fridge Recommendations - использует DTO вместо моделей
	GetFridgeRecommendations(req dto.FridgeRecommendationsRequest, fridgeItems []dto.AvailableIngredientDTO) ([]dto.FridgeRecommendation, error)
}

type aiService struct {
	groqClient *ai_core.GroqClient
}

// NewAIService creates new AI service
func NewAIService() AIService {
	return &aiService{
		groqClient: ai_core.NewGroqClient(),
	}
}

func (s *aiService) ChefMentor(req dto.ChefMentorRequest) (*dto.ChefMentorResponse, error) {
	// Validate
	if strings.TrimSpace(req.Message) == "" {
		return nil, ErrEmptyMessage
	}

	// Default language
	if req.Language == "" {
		req.Language = "ua"
	}

	// Initialize recipe draft if needed
	if req.CurrentRecipe == nil {
		req.CurrentRecipe = &dto.RecipeDraft{}
	}

	// Build system prompt
	systemPrompt := buildMentorSystemPrompt(req.Language)
	contextPrompt := buildRecipeContext(req.CurrentRecipe, req.Language)

	// Build conversation messages
	messages := []ai_core.GroqMessage{
		{Role: "system", Content: systemPrompt},
	}

	// Add conversation history
	for _, msg := range req.ConversationHistory {
		messages = append(messages, ai_core.GroqMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// Add context and current message
	messages = append(messages, ai_core.GroqMessage{
		Role:    "user",
		Content: contextPrompt + "\n\n" + req.Message,
	})

	// Call AI
	response, err := s.groqClient.Chat(messages, 0.7, 1000)
	if err != nil {
		return nil, fmt.Errorf("AI error: %w", err)
	}

	// Parse response and update recipe
	// This is simplified - real implementation would parse JSON from AI
	chefResponse := &dto.ChefMentorResponse{
		Message:      response.Choices[0].Message.Content,
		Recipe:       req.CurrentRecipe,
		NextQuestion: generateNextQuestion(req.CurrentRecipe, req.Language),
		IsComplete:   isRecipeComplete(req.CurrentRecipe),
	}

	// Add suggested actions if recipe is complete
	if chefResponse.IsComplete {
		chefResponse.SuggestedActions = []string{
			"save_recipe",
			"save_ingredients_to_fridge",
			"generate_meal_plan",
		}
	}

	return chefResponse, nil
}

func (s *aiService) GenerateMealPlan(req dto.MealPlanRequest, userID *uuid.UUID, fridgeItems []dto.AvailableIngredientDTO) (*dto.MealPlanResponse, error) {
	// Validate
	if req.Days < 1 || req.Days > 14 {
		return nil, ErrInvalidDays
	}
	if req.TargetCalories <= 0 {
		req.TargetCalories = 2000
	}
	if req.Language == "" {
		req.Language = "ua"
	}

	// Build prompt
	prompt := buildMealPlanPrompt(req, fridgeItems)

	// Call AI
	messages := []ai_core.GroqMessage{
		{Role: "system", Content: "You are a professional nutritionist and meal planning expert."},
		{Role: "user", Content: prompt},
	}

	response, err := s.groqClient.Chat(messages, 0.7, 2000)
	if err != nil {
		return nil, fmt.Errorf("AI error: %w", err)
	}

	// Parse response (simplified)
	plan := parseMealPlan(response.Choices[0].Message.Content, req.Days)

	return &dto.MealPlanResponse{
		Plan:          plan,
		TotalCalories: calculateTotalCalories(plan),
		AvgPerDay:     calculateAvgCalories(plan),
		Success:       true,
	}, nil
}

func (s *aiService) GenerateRecipe(req dto.RecipeGenerationRequest) (*dto.GeneratedRecipe, error) {
	// Validate
	if strings.TrimSpace(req.Title) == "" {
		return nil, ErrEmptyTitle
	}

	if req.Language == "" {
		req.Language = "pl"
	}

	// Build prompt
	prompt := buildRecipePrompt(req.Title, req.Language)

	// Call AI
	messages := []ai_core.GroqMessage{
		{Role: "system", Content: "You are a professional chef. Generate recipes in valid JSON format only."},
		{Role: "user", Content: prompt},
	}

	response, err := s.groqClient.Chat(messages, 0.7, 2000)
	if err != nil {
		return nil, fmt.Errorf("AI error: %w", err)
	}

	// Parse JSON response
	recipe, err := parseRecipeJSON(response.Choices[0].Message.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse recipe: %w", err)
	}

	return recipe, nil
}

func (s *aiService) GetFridgeRecommendations(req dto.FridgeRecommendationsRequest, fridgeItems []dto.AvailableIngredientDTO) ([]dto.FridgeRecommendation, error) {
	if len(fridgeItems) == 0 {
		return []dto.FridgeRecommendation{}, nil
	}

	// Build prompt with fridge items
	prompt := buildFridgeRecommendationsPrompt(req, fridgeItems)

	// Call AI
	messages := []ai_core.GroqMessage{
		{Role: "system", Content: "You are a creative chef that suggests recipes based on available ingredients."},
		{Role: "user", Content: prompt},
	}

	response, err := s.groqClient.Chat(messages, 0.8, 1500)
	if err != nil {
		return nil, fmt.Errorf("AI error: %w", err)
	}

	// Parse recommendations
	recommendations := parseFridgeRecommendations(response.Choices[0].Message.Content)

	return recommendations, nil
}

// Helper functions (simplified versions)

func buildMentorSystemPrompt(language string) string {
	prompts := map[string]string{
		"ua": "Ти професійний шеф-кухар і наставник. Допомагай створювати рецепти крок за кроком.",
		"en": "You are a professional chef and mentor. Help create recipes step by step.",
		"ru": "Ты профессиональный шеф-повар и наставник. Помогай создавать рецепты шаг за шагом.",
		"pl": "Jesteś profesjonalnym szefem kuchni i mentorem. Pomagaj tworzyć przepisy krok po kroku.",
	}
	if p, ok := prompts[language]; ok {
		return p
	}
	return prompts["ua"]
}

func buildRecipeContext(recipe *dto.RecipeDraft, language string) string {
	if recipe == nil || recipe.Title == "" {
		return ""
	}
	return fmt.Sprintf("Current recipe: %s\nIngredients: %d\nSteps: %d",
		recipe.Title, len(recipe.Ingredients), len(recipe.Steps))
}

func generateNextQuestion(recipe *dto.RecipeDraft, language string) string {
	if recipe.Title == "" {
		return "What dish would you like to create?"
	}
	if len(recipe.Ingredients) == 0 {
		return "What ingredients do you need?"
	}
	if len(recipe.Steps) == 0 {
		return "What are the cooking steps?"
	}
	return "Anything else to add?"
}

func isRecipeComplete(recipe *dto.RecipeDraft) bool {
	return recipe != nil &&
		recipe.Title != "" &&
		len(recipe.Ingredients) > 0 &&
		len(recipe.Steps) > 0
}

func buildMealPlanPrompt(req dto.MealPlanRequest, fridgeItems []dto.AvailableIngredientDTO) string {
	prompt := fmt.Sprintf("Generate a %d-day meal plan with %d calories per day.",
		req.Days, req.TargetCalories)

	if len(fridgeItems) > 0 {
		items := make([]string, len(fridgeItems))
		for i, item := range fridgeItems {
			items[i] = fmt.Sprintf("%s (%s)", item.Name, item.Quantity)
		}
		prompt += "\n\nAvailable ingredients: " + strings.Join(items, ", ")
	}

	return prompt
}

func parseMealPlan(response string, days int) []dto.DayMeal {
	// Simplified parsing - real implementation would parse AI response
	plan := make([]dto.DayMeal, days)
	for i := 0; i < days; i++ {
		plan[i] = dto.DayMeal{
			Day:           fmt.Sprintf("Day %d", i+1),
			Breakfast:     "Breakfast item",
			Lunch:         "Lunch item",
			Dinner:        "Dinner item",
			TotalCalories: 2000,
		}
	}
	return plan
}

func calculateTotalCalories(plan []dto.DayMeal) float64 {
	total := 0.0
	for _, day := range plan {
		total += day.TotalCalories
	}
	return total
}

func calculateAvgCalories(plan []dto.DayMeal) float64 {
	if len(plan) == 0 {
		return 0
	}
	return calculateTotalCalories(plan) / float64(len(plan))
}

func buildRecipePrompt(title, language string) string {
	return fmt.Sprintf("Generate a complete recipe for: %s (language: %s)", title, language)
}

func parseRecipeJSON(response string) (*dto.GeneratedRecipe, error) {
	// Simplified - real implementation would parse JSON
	return &dto.GeneratedRecipe{
		Title:       "Generated Recipe",
		Description: "AI generated recipe",
		Difficulty:  "intermediate",
		Time:        30,
		Portions:    4,
	}, nil
}

func buildFridgeRecommendationsPrompt(req dto.FridgeRecommendationsRequest, fridgeItems []dto.AvailableIngredientDTO) string {
	items := make([]string, len(fridgeItems))
	for i, item := range fridgeItems {
		items[i] = fmt.Sprintf("%s (%s)", item.Name, item.Quantity)
	}

	prompt := "Suggest recipes using these ingredients: " + strings.Join(items, ", ")

	if req.Cuisine != "" {
		prompt += fmt.Sprintf("\nCuisine: %s", req.Cuisine)
	}
	if req.MaxTime > 0 {
		prompt += fmt.Sprintf("\nMax cooking time: %d minutes", req.MaxTime)
	}

	return prompt
}

func parseFridgeRecommendations(response string) []dto.FridgeRecommendation {
	// Simplified parsing
	return []dto.FridgeRecommendation{
		{
			RecipeName:      "Sample Recipe",
			Description:     "A delicious recipe",
			MatchPercentage: 85,
			MissingItems:    []string{},
			PrepTime:        30,
			Difficulty:      "easy",
		},
	}
}
