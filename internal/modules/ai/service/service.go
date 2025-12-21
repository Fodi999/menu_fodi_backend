package service

import (
	"errors"
	"fmt"
	"strings"
	"encoding/json"

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
	
	// SMART KITCHEN: Create Restaurant Recipe from Fridge
	CreateRecipeFromFridge(userID string, language string, fridgeItems []dto.FridgeItemDTO) (*dto.CreateRecipeFromFridgeResponse, error)
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

	// 🔧 СПЕЦИАЛЬНАЯ ОБРАБОТКА для budget_review: считаем на backend, AI только комментирует
	var budgetSummary string
	if req.Goal == "budget_review" {
		budgetSummary = calculateBudgetSummary(fridgeItems, language)
	}

	// System prompt на выбранном языке + КОМПАКТНЫЙ список доступных продуктов
	baseSystemPrompt := prompts.FridgeSystemPrompt[language]
	strictSystemPrompt := fmt.Sprintf(`%s

DOSTĘPNE PRODUKTY (używaj TYLKO tych):
%s

ZAKAZ dodawania innych składników!`, 
		baseSystemPrompt,
		ingredientsList)

	// Строим user prompt с учётом языка
	var prompt string
	if req.Goal == "budget_review" && budgetSummary != "" {
		// Для budget_review используем готовую аналитику
		goalPrompt := ""
		if goalTexts, ok := prompts.GoalPrompts["budget_review"]; ok {
			if text, ok := goalTexts[language]; ok {
				goalPrompt = text
			}
		}
		prompt = fmt.Sprintf(`%s

%s

Teraz skomentuj te dane i dodaj praktyczne porady jak zaoszczędzić.`, goalPrompt, budgetSummary)
	} else {
		prompt = buildFridgeAnalysisPrompt(req.Goal, language, fridgeItems)
	}

	// ВАЖНО: Проверяем что промпт не пустой
	if strings.TrimSpace(prompt) == "" {
		return "", fmt.Errorf("empty prompt for goal: %s, language: %s", req.Goal, language)
	}

	// Отправляем в Groq AI
	response, err := s.groqClient.SimpleChat(strictSystemPrompt, prompt)
	if err != nil {
		return "", fmt.Errorf("AI analysis failed: %w", err)
	}

	// 🔧 ПАРСИНГ JSON (Phase 3A)
	parsedJSON, isJSON, parseErr := parseAIResponse(response, req.Goal)
	
	// Если AI вернул валидный JSON - возвращаем его
	if isJSON && parseErr == nil {
		return parsedJSON, nil
	}
	
	// Если AI вернул невалидный JSON или текст - используем fallback

	// ВАЖНО: Проверяем что AI вернул непустой ответ (fallback logic)
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

// parseAIResponse пытается распарсить JSON ответ от AI
// Возвращает: (jsonString, isJSON, error)
func parseAIResponse(response string, goal string) (string, bool, error) {
	response = strings.TrimSpace(response)
	
	// Проверка 1: Ответ пустой
	if response == "" {
		return "", false, fmt.Errorf("empty response")
	}
	
	// Проверка 2: Попытка найти JSON (может быть обёрнут в markdown)
	// Убираем markdown блоки если есть
	if strings.Contains(response, "```json") {
		start := strings.Index(response, "```json")
		end := strings.LastIndex(response, "```")
		if start != -1 && end != -1 && end > start {
			response = response[start+7 : end]
			response = strings.TrimSpace(response)
		}
	} else if strings.Contains(response, "```") {
		// Убираем обычные ``` блоки
		start := strings.Index(response, "```")
		end := strings.LastIndex(response, "```")
		if start != -1 && end != -1 && end > start {
			response = response[start+3 : end]
			response = strings.TrimSpace(response)
		}
	}
	
	// Проверка 3: Выглядит ли как JSON?
	if !strings.HasPrefix(response, "{") || !strings.HasSuffix(response, "}") {
		return response, false, fmt.Errorf("response is not JSON")
	}
	
	// Проверка 4: Валидный ли JSON?
	var testJSON map[string]interface{}
	if err := json.Unmarshal([]byte(response), &testJSON); err != nil {
		return response, false, fmt.Errorf("invalid JSON: %w", err)
	}
	
	// Проверка 5: Есть ли поле error?
	if errorMsg, ok := testJSON["error"].(string); ok && errorMsg != "" {
		return response, true, fmt.Errorf("AI returned error: %s", errorMsg)
	}
	
	return response, true, nil
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

// calculateBudgetSummary считает финансовую аналитику НА BACKEND (AI только комментирует)
func calculateBudgetSummary(items []dto.FridgeItemDTO, language string) string {
	if len(items) == 0 {
		return ""
	}

	// Считаем общую стоимость и находим самые дорогие продукты
	totalValue := 0.0
	type expensiveItem struct {
		name       string
		totalPrice float64
		daysLeft   int
		risk       string
	}
	expensiveItems := make([]expensiveItem, 0)

	for _, item := range items {
		// Calculate total price: quantity * pricePerUnit
		if item.PricePerUnit != nil && *item.PricePerUnit > 0 && item.Quantity > 0 {
			itemTotalPrice := item.Quantity * (*item.PricePerUnit)
			totalValue += itemTotalPrice
			
			risk := "ok"
			daysLeft := 999
			if item.DaysLeft != nil {
				daysLeft = *item.DaysLeft
				if daysLeft <= 2 {
					risk = "critical"
				} else if daysLeft <= 5 {
					risk = "warning"
				}
			}
			
			expensiveItems = append(expensiveItems, expensiveItem{
				name:       item.Name,
				totalPrice: itemTotalPrice,
				daysLeft:   daysLeft,
				risk:       risk,
			})
		}
	}

	// Сортируем по цене (самые дорогие первыми)
	for i := 0; i < len(expensiveItems)-1; i++ {
		for j := i + 1; j < len(expensiveItems); j++ {
			if expensiveItems[j].totalPrice > expensiveItems[i].totalPrice {
				expensiveItems[i], expensiveItems[j] = expensiveItems[j], expensiveItems[i]
			}
		}
	}

	// Берём топ-3 самых дорогих
	topExpensive := expensiveItems
	if len(topExpensive) > 3 {
		topExpensive = topExpensive[:3]
	}

	// Считаем потенциальные потери (дорогие продукты с коротким сроком)
	potentialLoss := 0.0
	criticalExpensive := make([]expensiveItem, 0)
	for _, item := range expensiveItems {
		if item.risk == "critical" || item.risk == "warning" {
			potentialLoss += item.totalPrice
			criticalExpensive = append(criticalExpensive, item)
		}
	}

	// Форматируем результат в зависимости от языка
	currency := "PLN"
	if len(items) > 0 {
		currency = items[0].Currency
	}

	summaryTemplates := map[string]string{
		"pl": `**ANALIZA BUDŻETU (obliczona przez system):**

📊 PODSUMOWANIE:
- Całkowita wartość lodówki: %.2f %s
- Liczba produktów z ceną: %d
- Średnia wartość produktu: %.2f %s

💰 NAJDROŻSZE PRODUKTY:
%s

⚠️ RYZYKO STRAT:
- Produkty kończące się (≤5 dni): %.2f %s
- Potencjalna strata jeśli nie wykorzystasz: %.2f %s
%s`,
		"en": `**BUDGET ANALYSIS (calculated by system):**

📊 SUMMARY:
- Total fridge value: %.2f %s
- Products with price: %d
- Average product value: %.2f %s

💰 MOST EXPENSIVE PRODUCTS:
%s

⚠️ LOSS RISK:
- Expiring products (≤5 days): %.2f %s
- Potential loss if not used: %.2f %s
%s`,
		"ru": `**АНАЛИЗ БЮДЖЕТА (рассчитано системой):**

📊 ИТОГО:
- Общая стоимость холодильника: %.2f %s
- Продуктов с ценой: %d
- Средняя стоимость продукта: %.2f %s

💰 САМЫЕ ДОРОГИЕ ПРОДУКТЫ:
%s

⚠️ РИСК ПОТЕРЬ:
- Истекающие продукты (≤5 дней): %.2f %s
- Потенциальная потеря если не использовать: %.2f %s
%s`,
	}

	template := summaryTemplates[language]
	
	// Форматируем список дорогих продуктов
	expensiveList := ""
	for i, item := range topExpensive {
		riskEmoji := ""
		if item.risk == "critical" {
			riskEmoji = " 🚨"
		} else if item.risk == "warning" {
			riskEmoji = " ⚠️"
		}
		expensiveList += fmt.Sprintf("%d. %s: %.2f %s (zostało %d dni)%s\n", 
			i+1, item.name, item.totalPrice, currency, item.daysLeft, riskEmoji)
	}

	// Форматируем список критических продуктов
	criticalList := ""
	if len(criticalExpensive) > 0 {
		labels := map[string]string{
			"pl": "\n\n🔴 PRODUKTY DO NATYCHMIASTOWEGO WYKORZYSTANIA:",
			"en": "\n\n🔴 PRODUCTS TO USE IMMEDIATELY:",
			"ru": "\n\n🔴 ПРОДУКТЫ ДЛЯ НЕМЕДЛЕННОГО ИСПОЛЬЗОВАНИЯ:",
		}
		criticalList = labels[language]
		for _, item := range criticalExpensive {
			criticalList += fmt.Sprintf("\n- %s: %.2f %s (zostało %d dni)", 
				item.name, item.totalPrice, currency, item.daysLeft)
		}
	}

	avgValue := 0.0
	if len(expensiveItems) > 0 {
		avgValue = totalValue / float64(len(expensiveItems))
	}

	return fmt.Sprintf(template,
		totalValue, currency,
		len(expensiveItems),
		avgValue, currency,
		expensiveList,
		potentialLoss, currency,
		potentialLoss, currency,
		criticalList)
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
		if item.PricePerUnit != nil && *item.PricePerUnit > 0 && item.Quantity > 0 {
			totalPrice := item.Quantity * (*item.PricePerUnit)
			priceInfo = fmt.Sprintf(" [%.2f %s]", totalPrice, item.Currency)
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

// CreateRecipeFromFridge creates a restaurant-grade recipe from fridge products only
func (s *aiService) CreateRecipeFromFridge(userID string, language string, fridgeItems []dto.FridgeItemDTO) (*dto.CreateRecipeFromFridgeResponse, error) {
	// 1. Validate language
	language = prompts.NormalizeLanguage(language)
	
	// 2. Check if fridge is empty
	if len(fridgeItems) == 0 {
		messages := map[string]string{
			"pl": "Twoja lodówka jest pusta. Dodaj produkty, aby stworzyć przepis!",
			"en": "Your fridge is empty. Add products to create a recipe!",
			"ru": "Твой холодильник пуст. Добавь продукты, чтобы создать рецепт!",
		}
		return &dto.CreateRecipeFromFridgeResponse{
			Success: false,
			Message: messages[language],
		}, nil
	}
	
	// 3. Filter and prioritize products by expiry
	type PrioritizedProduct struct {
		Item     dto.FridgeItemDTO
		Priority int // 1=critical, 2=warning, 3=ok
	}
	
	var products []PrioritizedProduct
	for _, item := range fridgeItems {
		if item.Quantity <= 0 {
			continue // Skip products with zero quantity
		}
		if item.Status == "expired" {
			continue // Skip expired products
		}
		
		priority := 3 // default: ok
		switch item.Status {
		case "critical":
			priority = 1
		case "warning":
			priority = 2
		}
		
		products = append(products, PrioritizedProduct{
			Item:     item,
			Priority: priority,
		})
	}
	
	if len(products) == 0 {
		messages := map[string]string{
			"pl": "Brak dostępnych produktów do użycia. Wszystkie produkty są przeterminowane lub ich ilość wynosi 0.",
			"en": "No available products to use. All products are expired or have zero quantity.",
			"ru": "Нет доступных продуктов для использования. Все продукты просрочены или их количество равно 0.",
		}
		return &dto.CreateRecipeFromFridgeResponse{
			Success: false,
			Message: messages[language],
		}, nil
	}
	
	// Sort by priority (critical first)
	for i := 0; i < len(products)-1; i++ {
		for j := i + 1; j < len(products); j++ {
			if products[j].Priority < products[i].Priority {
				products[i], products[j] = products[j], products[i]
			}
		}
	}
	
	// 4. Build "kitchen context" for AI with formatted product list
	var fridgeContext strings.Builder
	for idx, prod := range products {
		item := prod.Item
		
		// Format expiry date
		expiryText := ""
		priorityLabel := ""
		if item.DaysLeft != nil {
			daysLeft := *item.DaysLeft
			if daysLeft <= 2 {
				priorityLabel = " (PRIORITY)"
				switch language {
				case "pl":
					expiryText = fmt.Sprintf("termin: %d dzień/dni", daysLeft)
				case "en":
					expiryText = fmt.Sprintf("expiry: %d day(s)", daysLeft)
				case "ru":
					expiryText = fmt.Sprintf("срок: %d день/дней", daysLeft)
				}
			} else {
				switch language {
				case "pl":
					expiryText = fmt.Sprintf("termin: %d dni", daysLeft)
				case "en":
					expiryText = fmt.Sprintf("expiry: %d days", daysLeft)
				case "ru":
					expiryText = fmt.Sprintf("срок: %d дней", daysLeft)
				}
			}
		}
		
		fridgeContext.WriteString(fmt.Sprintf("\n%d. %s%s\n", idx+1, item.Name, priorityLabel))
		fridgeContext.WriteString(fmt.Sprintf("   ilość: %.0f %s\n", item.Quantity, item.Unit))
		if expiryText != "" {
			fridgeContext.WriteString(fmt.Sprintf("   %s\n", expiryText))
		}
	}
	
	// 5. Get prompt template for language
	promptTemplate, ok := prompts.RestaurantRecipePrompt[language]
	if !ok {
		promptTemplate = prompts.RestaurantRecipePrompt["pl"] // fallback to Polish
	}
	
	// Inject product list into prompt
	prompt := fmt.Sprintf(promptTemplate, fridgeContext.String())
	
	// 6. Call AI with retry + self-repair mechanism
	response, err := s.groqClient.SimpleChat("", prompt)
	if err != nil {
		return nil, fmt.Errorf("AI recipe generation failed: %w", err)
	}
	
	// 7. Parse JSON response with self-repair retry
	parsedJSON, isJSON, parseErr := parseAIResponse(response, "create_recipe")
	
	if !isJSON || parseErr != nil {
		// 🔄 RETRY: Try to repair invalid JSON with AI
		fmt.Printf("[AI][RETRY] First attempt failed, trying self-repair...\n")
		fmt.Printf("[AI][RETRY] Original error: %v\n", parseErr)
		fmt.Printf("[AI][RETRY] Raw response length: %d chars\n", len(response))
		
		repairPrompt := fmt.Sprintf(`You are a JSON repair API.

The following response is invalid JSON. Fix it and return ONLY valid JSON matching this schema:

{
  "name": "string",
  "description": "string",
  "ingredientsUsed": [{"name": "string", "quantity": number, "unit": "string"}],
  "ingredientsMissing": [{"name": "string", "quantity": number, "unit": "string"}],
  "steps": ["string"],
  "cookingTime": number,
  "chefTips": ["string"],
  "expiryPriority": "critical|warning|ok",
  "economy": {
    "usedFromFridge": boolean,
    "estimatedExtraCost": number,
    "currency": "PLN"
  }
}

CRITICAL RULES:
1. Return ONLY valid JSON
2. NO markdown, NO comments, NO explanations
3. If JSON is incomplete, complete it logically
4. All numbers must be numbers (not strings)
5. All required fields must be present

INVALID RESPONSE TO FIX:
%s

Return ONLY the fixed JSON:`, response)

		repairedResponse, repairErr := s.groqClient.SimpleChat("", repairPrompt)
		if repairErr != nil {
			fmt.Printf("[AI][RETRY] Repair call failed: %v\n", repairErr)
			return &dto.CreateRecipeFromFridgeResponse{
				Success: false,
				Message: "Failed to generate recipe in valid format. Please try again.",
			}, nil
		}
		
		// Try parsing repaired response
		parsedJSON, isJSON, parseErr = parseAIResponse(repairedResponse, "create_recipe")
		
		if !isJSON || parseErr != nil {
			fmt.Printf("[AI][RETRY] Self-repair also failed\n")
			fmt.Printf("[AI][RETRY] Repaired response: %s\n", repairedResponse)
			fmt.Printf("[AI][RETRY] Parse error: %v\n", parseErr)
			return &dto.CreateRecipeFromFridgeResponse{
				Success: false,
				Message: "Failed to generate recipe in valid format. Please try again.",
			}, nil
		}
		
		fmt.Printf("[AI][RETRY] ✅ Self-repair succeeded!\n")
	}
	
	// 8. Decode into RestaurantRecipe
	var recipe dto.RestaurantRecipe
	if err := json.Unmarshal([]byte(parsedJSON), &recipe); err != nil {
		fmt.Printf("[AI][ERROR] Failed to unmarshal JSON into RestaurantRecipe\n")
		fmt.Printf("[AI][ERROR] Parsed JSON: %s\n", parsedJSON)
		fmt.Printf("[AI][ERROR] Unmarshal error: %v\n", err)
		return &dto.CreateRecipeFromFridgeResponse{
			Success: false,
			Message: "Failed to parse recipe data. Please try again.",
		}, nil
	}
	
	// Validate critical fields
	if recipe.Name == "" {
		fmt.Printf("[AI][ERROR] Recipe name is empty\n")
		fmt.Printf("[AI][ERROR] Parsed JSON: %s\n", parsedJSON)
		return &dto.CreateRecipeFromFridgeResponse{
			Success: false,
			Message: "Recipe missing required field: name.",
		}, nil
	}
	
	fmt.Printf("[AI][SUCCESS] Recipe parsed successfully: %s\n", recipe.Name)
	fmt.Printf("[AI][SUCCESS] IngredientsUsed: %d, IngredientsMissing: %d\n", 
		len(recipe.IngredientsUsed), len(recipe.IngredientsMissing))
	fmt.Printf("[AI][SUCCESS] Economy: %+v\n", recipe.Economy)
	
	// 9. Build list of used products with cost calculation
	usedProducts := make([]dto.UsedProductInfo, 0)
	totalUsedCost := 0.0
	currency := "PLN" // default
	
	fmt.Printf("[AI][ECONOMY DEBUG] Starting cost calculation for %d products\n", len(products))
	
	for _, prod := range products {
		// For simplicity, assume AI used all critical/warning products
		if prod.Priority <= 2 {
			// Calculate cost of used product
			usedCost := 0.0
			pricePerUnit := 0.0
			
			// DEBUG: Log each product before calculation
			priceStr := "NULL"
			if prod.Item.PricePerUnit != nil {
				priceStr = fmt.Sprintf("%.6f", *prod.Item.PricePerUnit)
			}
			fmt.Printf("[ECONOMY] Product: %s | qty=%.2f %s | price=%s | priority=%d\n",
				prod.Item.Name, prod.Item.Quantity, prod.Item.Unit, priceStr, prod.Priority)
			
			if prod.Item.PricePerUnit != nil && *prod.Item.PricePerUnit > 0 {
				pricePerUnit = *prod.Item.PricePerUnit
				usedCost = prod.Item.Quantity * pricePerUnit
				totalUsedCost += usedCost
				
				if prod.Item.Currency != "" {
					currency = prod.Item.Currency
				}
				
				// DEBUG: Log successful calculation
				fmt.Printf("[ECONOMY] ✅ Calculated cost: %.2f %s (%.2f × %.6f)\n",
					usedCost, currency, prod.Item.Quantity, pricePerUnit)
			} else {
				// DEBUG: Log when price is missing
				fmt.Printf("[ECONOMY] ⚠️ NO PRICE DATA - skipping cost calculation\n")
			}
			
			usedProducts = append(usedProducts, dto.UsedProductInfo{
				Name:         prod.Item.Name,
				QuantityUsed: prod.Item.Quantity,
				Unit:         prod.Item.Unit,
				PricePerUnit: pricePerUnit,
				UsedCost:     usedCost,
				Currency:     currency,
				DaysLeft:     prod.Item.DaysLeft,
			})
		}
	}
	
	fmt.Printf("[AI][ECONOMY DEBUG] Total products processed: %d, Total cost: %.2f %s\n",
		len(usedProducts), totalUsedCost, currency)
	
	// 10. Calculate economy and override AI's estimatedExtraCost
	// AI may return pantry cost, but we trust backend calculation more
	estimatedExtraCost := 0.0
	if recipe.Economy != nil && recipe.Economy.EstimatedExtraCost > 0 {
		estimatedExtraCost = recipe.Economy.EstimatedExtraCost
	}
	
	savedMoney := totalUsedCost - estimatedExtraCost
	if savedMoney < 0 {
		savedMoney = 0 // Can't have negative savings
	}
	
	// 🔥 CRITICAL DEBUG: Log BEFORE setting economy
	fmt.Printf("[AI][ECONOMY] ⚠️ BEFORE override - recipe.Economy = %+v\n", recipe.Economy)
	fmt.Printf("[AI][ECONOMY] ⚠️ About to set: UsedValue=%.2f, SavedMoney=%.2f, Currency=%s\n",
		totalUsedCost, savedMoney, currency)
	
	// ALWAYS override economy block with backend-calculated values (even if prices missing)
	// This ensures frontend always receives economy structure
	recipe.Economy = &dto.RecipeEconomy{
		UsedFromFridge:     len(usedProducts) > 0,
		UsedValue:          totalUsedCost,          // 0.0 if no prices
		EstimatedExtraCost: estimatedExtraCost,
		SavedMoney:         savedMoney,             // 0.0 if no prices
		Currency:           currency,
	}
	
	// 🔥 CRITICAL DEBUG: Log AFTER setting economy
	fmt.Printf("[AI][ECONOMY] ✅ AFTER override - recipe.Economy = %+v\n", recipe.Economy)
	fmt.Printf("[AI][ECONOMY] ✅ Memory address of recipe: %p\n", &recipe)
	fmt.Printf("[AI][ECONOMY] ✅ Memory address of recipe.Economy: %p\n", recipe.Economy)
	// 🔥 CRITICAL DEBUG: Log AFTER setting economy
	fmt.Printf("[AI][ECONOMY] ✅ AFTER override - recipe.Economy = %+v\n", recipe.Economy)
	fmt.Printf("[AI][ECONOMY] ✅ Memory address of recipe: %p\n", &recipe)
	fmt.Printf("[AI][ECONOMY] ✅ Memory address of recipe.Economy: %p\n", recipe.Economy)
	
	fmt.Printf("[AI][ECONOMY] Used cost: %.2f %s, Extra cost: %.2f %s, Saved: %.2f %s (prices available: %d products)\n",
		totalUsedCost, currency, estimatedExtraCost, currency, savedMoney, currency, len(usedProducts))
	
	// 11. Return successful result
	fmt.Printf("[AI][ECONOMY] 🚀 About to return response with recipe at address: %p\n", &recipe)
	return &dto.CreateRecipeFromFridgeResponse{
		Success:      true,
		Recipe:       &recipe,
		UsedProducts: usedProducts,
	}, nil
}
