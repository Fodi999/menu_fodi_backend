package service

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/ai/dto"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/ai/prompts"
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
	
	// SMART KITCHEN: AI Fridge Analysis
	AnalyzeFridge(userID string, req dto.FridgeAnalyzeRequest, fridgeItems []dto.FridgeItemDTO) (string, error)
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

// AnalyzeFridge анализирует холодильник через AI (SMART KITCHEN)
func (s *aiService) AnalyzeFridge(userID string, req dto.FridgeAnalyzeRequest, fridgeItems []dto.FridgeItemDTO) (string, error) {
	if len(fridgeItems) == 0 {
		return "Your fridge is empty. Add some items to get AI recommendations!", nil
	}

	// Нормализуем язык
	language := prompts.NormalizeLanguage(req.Language)

	// 🔴 GUARD: Проверка минимального количества продуктов для 3-дневного плана
	if req.Goal == "3_days_plan" && len(fridgeItems) < 3 {
		fallbackMessages := map[string]string{
			"pl": "Masz za mało produktów, aby ułożyć sensowny plan na 3 dni. Dodaj więcej składników do lodówki (minimum 5-7 produktów dla pełnego planu).",
			"en": "Not enough products to create a 3-day plan. Add more ingredients to your fridge (minimum 5-7 products for a complete plan).",
			"ru": "Недостаточно продуктов для составления плана на 3 дня. Добавь больше ингредиентов в холодильник (минимум 5-7 продуктов для полного плана).",
		}
		return fallbackMessages[language], nil
	}

	// Строим список доступных ингредиентов для system prompt
	ingredientsList := buildIngredientsListForPrompt(fridgeItems)

	// System prompt на выбранном языке + КОМПАКТНЫЙ список доступных продуктов
	baseSystemPrompt := prompts.FridgeSystemPrompt[language]
	strictSystemPrompt := fmt.Sprintf(`%s

DOSTĘPNE PRODUKTY (używaj TYLKO tych):
%s

ZAKAZ dodawania innych składników!`, 
		baseSystemPrompt,
		ingredientsList)

	// Строим user prompt с учётом языка
	prompt := buildFridgeAnalysisPrompt(req.Goal, language, fridgeItems)

	// ВАЖНО: Проверяем что промпт не пустой
	if strings.TrimSpace(prompt) == "" {
		return "", fmt.Errorf("empty prompt for goal: %s, language: %s", req.Goal, language)
	}

	// Отправляем в Groq AI
	response, err := s.groqClient.SimpleChat(strictSystemPrompt, prompt)
	if err != nil {
		return "", fmt.Errorf("AI analysis failed: %w", err)
	}

	// ВАЖНО: Проверяем что AI вернул непустой ответ
	if strings.TrimSpace(response) == "" {
		// Возвращаем специфичное сообщение по цели
		goalSpecificFallback := map[string]map[string]string{
			"3_days_plan": {
				"pl": "Za mało produktów w lodówce, aby ułożyć pełny plan na 3 dni. Dodaj więcej składników (mięso, warzywa, węglowodany).",
				"en": "Not enough products in the fridge to create a full 3-day plan. Add more ingredients (meat, vegetables, carbs).",
				"ru": "Недостаточно продуктов в холодильнике для создания полного плана на 3 дня. Добавь больше ингредиентов (мясо, овощи, углеводы).",
			},
			"today_meals": {
				"pl": "AI nie udało się wygenerować przepisu z dostępnych produktów. Spróbuj dodać więcej składników.",
				"en": "AI couldn't generate a recipe with available products. Try adding more ingredients.",
				"ru": "AI не смог создать рецепт из доступных продуктов. Попробуй добавить больше ингредиентов.",
			},
			"reduce_waste": {
				"pl": "AI nie może przeanalizować produktów. Sprawdź czy produkty mają poprawne daty ważności.",
				"en": "AI cannot analyze products. Check if products have valid expiry dates.",
				"ru": "AI не может проанализировать продукты. Проверь, правильно ли указаны сроки годности.",
			},
			"budget_review": {
				"pl": "AI nie może przeanalizować wydatków. Upewnij się, że produkty mają przypisane ceny.",
				"en": "AI cannot analyze expenses. Make sure products have prices assigned.",
				"ru": "AI не может проанализировать расходы. Убедись, что у продуктов указаны цены.",
			},
		}
		
		if messages, ok := goalSpecificFallback[req.Goal]; ok {
			if msg, ok := messages[language]; ok {
				return msg, nil
			}
		}
		
		return "", fmt.Errorf("AI returned empty response for goal: %s", req.Goal)
	}

	return response, nil
}

// buildIngredientsListForPrompt создаёт строгий список доступных ингредиентов
func buildIngredientsListForPrompt(items []dto.FridgeItemDTO) string {
	if len(items) == 0 {
		return "BRAK PRODUKTÓW"
	}

	ingredientsList := make([]string, 0, len(items))
	for i, item := range items {
		status := item.Status
		if item.DaysLeft != nil {
			status = fmt.Sprintf("%s, zostało %d dni", status, *item.DaysLeft)
		}
		
		ingredientsList = append(ingredientsList, fmt.Sprintf("%d. %s - %.0f %s [%s]", 
			i+1, item.Name, item.Quantity, item.Unit, status))
	}

	return strings.Join(ingredientsList, "\n")
}

// buildFridgeAnalysisPrompt строит промпт для анализа холодильника с учётом языка
func buildFridgeAnalysisPrompt(goal string, language string, items []dto.FridgeItemDTO) string {
	// Сериализуем items в строку
	itemsList := make([]string, 0, len(items))
	for _, item := range items {
		status := item.Status
		if item.DaysLeft != nil {
			status = fmt.Sprintf("%s (%d days)", status, *item.DaysLeft)
		}
		
		priceInfo := ""
		if item.TotalPrice != nil && *item.TotalPrice > 0 {
			priceInfo = fmt.Sprintf(" [%.2f %s]", *item.TotalPrice, item.Currency)
		}
		
		itemsList = append(itemsList, fmt.Sprintf("- %s: %.1f %s [%s]%s", 
			item.Name, item.Quantity, item.Unit, status, priceInfo))
	}

	// Получаем промпт цели на нужном языке
	goalPrompt := ""
	if goalTexts, ok := prompts.GoalPrompts[goal]; ok {
		if text, ok := goalTexts[language]; ok {
			goalPrompt = text
		}
	}

	// ФОЛЛБЭК: если промпт цели не найден - используем базовый
	if strings.TrimSpace(goalPrompt) == "" {
		fallbackGoals := map[string]string{
			"pl": fmt.Sprintf("\n\nCEL: %s\n\nPrzeanalizuj produkty i podaj rekomendacje.", goal),
			"en": fmt.Sprintf("\n\nGOAL: %s\n\nAnalyze products and provide recommendations.", goal),
			"ru": fmt.Sprintf("\n\nЦЕЛЬ: %s\n\nПроанализируй продукты и дай рекомендации.", goal),
		}
		goalPrompt = fallbackGoals[language]
	}

	// Формируем финальный prompt
	prompt := fmt.Sprintf(`%s

**Produkty w lodówce:**
%s

Proszę o konkretne rekomendacje.`,
		goalPrompt,
		strings.Join(itemsList, "\n"))

	return prompt
}
