package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/ai"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/database"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/dmitrijfomin/menu-fodifood/backend/pkg/utils"
	"github.com/google/uuid"
)

// Repository instance
var chefMentorRepo = database.NewChefMentorRepository()

// ChefMentorSessionRequest with session ID
type ChefMentorSessionRequest struct {
	Message   string `json:"message"`
	SessionID string `json:"sessionId,omitempty"` // Optional: если есть - продолжаем, если нет - новая сессия
	Language  string `json:"language,omitempty"`
}

// ChefMentorSessionResponse with session management
type ChefMentorSessionResponse struct {
	SessionID        string       `json:"sessionId"`
	Message          string       `json:"message"`
	Recipe           *RecipeDraft `json:"recipe"`
	NextQuestion     string       `json:"nextQuestion"`
	IsComplete       bool         `json:"isComplete"`
	SuggestedActions []string     `json:"suggestedActions,omitempty"`
	QuickReplies     []string     `json:"quickReplies,omitempty"` // NEW: контекстные быстрые ответы
}

// ChefMentorSessionHandler with database-backed session management
// POST /api/ai/chef-mentor/session
func ChefMentorSessionHandler(w http.ResponseWriter, r *http.Request) {
	var req ChefMentorSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid input")
		return
	}

	// Get or create session from database
	var dbSession *models.ChefMentorSession
	var err error

	if req.SessionID != "" {
		// Retrieve existing session
		dbSession, err = chefMentorRepo.GetSession(req.SessionID)
		if err != nil {
			utils.RespondWithError(w, http.StatusNotFound, "Session not found")
			return
		}
	} else {
		// Create new session
		lang := req.Language
		if lang == "" {
			lang = "ua"
		}
		dbSession, err = chefMentorRepo.CreateSession(nil, lang) // nil userID for anonymous
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create session")
			return
		}
	}

	sessionIDStr := dbSession.ID.String()

	// Get conversation history from database
	dbMessages, err := chefMentorRepo.GetMessages(sessionIDStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to load messages")
		return
	}

	// Parse current recipe from database
	currentRecipe := &RecipeDraft{}
	if len(dbSession.Recipe) > 0 {
		recipeJSON, _ := json.Marshal(dbSession.Recipe)
		json.Unmarshal(recipeJSON, currentRecipe)
	}

	// Build system prompt
	systemPrompt := buildMentorSystemPrompt(dbSession.Language)
	contextPrompt := buildRecipeContext(currentRecipe, dbSession.Language)

	// Create AI messages
	client := ai.NewGroqClient()
	messages := []ai.GroqMessage{
		{
			Role:    "system",
			Content: systemPrompt + "\n\n" + contextPrompt,
		},
	}

	// Add history from database
	for _, msg := range dbMessages {
		messages = append(messages, ai.GroqMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// Add current user message
	messages = append(messages, ai.GroqMessage{
		Role:    "user",
		Content: req.Message,
	})

	// 🧠 Smart context detection: handle greetings & commands BEFORE AI
	contextResponse := detectUserIntent(req.Message, currentRecipe, dbSession.Language)
	var assistantMessage string
	
	if contextResponse != "" {
		// Use pre-defined response for greetings/commands
		assistantMessage = contextResponse
	} else {
		// Get AI response for recipe creation
		response, err := client.Chat(messages, 0.7, 1000)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "AI service error")
			return
		}
		assistantMessage = response.Choices[0].Message.Content
	}

	// Save messages to database
	if err := chefMentorRepo.SaveMessage(dbSession.ID, "user", req.Message); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to save user message")
		return
	}
	if err := chefMentorRepo.SaveMessage(dbSession.ID, "assistant", assistantMessage); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to save assistant message")
		return
	}

	// Extract recipe updates using improved logic
	currentRecipe = smartExtractRecipeUpdates(assistantMessage, currentRecipe, req.Message, dbSession.Language)

	// Calculate nutrition and cost if ingredients were added
	if len(currentRecipe.Ingredients) > 0 {
		calculateRecipeMetrics(currentRecipe)
	}

	// Generate human-like commentary if metrics calculated
	enhancedMessage := assistantMessage
	if currentRecipe.Calories > 0 {
		commentary := generateNutritionCommentary(currentRecipe, dbSession.Language)
		if commentary != "" {
			enhancedMessage = assistantMessage + "\n\n" + commentary
		}
	}

	// Update session with new recipe data
	if err := chefMentorRepo.UpdateSession(sessionIDStr, currentRecipe, nil); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update session")
		return
	}

	// Determine next question
	nextQuestion := determineNextQuestion(currentRecipe, dbSession.Language)
	isComplete := isRecipeComplete(currentRecipe)

	// Mark complete if done AND auto-save recipe to AI Culinary OS
	if isComplete {
		chefMentorRepo.MarkComplete(sessionIDStr)
		
		// 🚀 AUTO-SAVE TO AI CULINARY OS
		savedRecipeID, err := saveCompletedRecipeToDatabase(dbSession.ID, dbSession.UserID, currentRecipe, dbSession.Language)
		if err == nil && savedRecipeID != "" {
			// Add success message to response
			successMsg := map[string]string{
				"ua": fmt.Sprintf("\n\n✅ Рецепт автоматично збережено в AI Culinary OS! ID: %s", savedRecipeID[:8]),
				"en": fmt.Sprintf("\n\n✅ Recipe automatically saved to AI Culinary OS! ID: %s", savedRecipeID[:8]),
				"ru": fmt.Sprintf("\n\n✅ Рецепт автоматически сохранён в AI Culinary OS! ID: %s", savedRecipeID[:8]),
				"pl": fmt.Sprintf("\n\n✅ Przepis automatycznie zapisany w AI Culinary OS! ID: %s", savedRecipeID[:8]),
			}
			if msg, ok := successMsg[dbSession.Language]; ok {
				enhancedMessage += msg
			}
		}
	}

	// Generate contextual quick replies
	quickReplies := generateQuickReplies(currentRecipe, dbSession.Language)

	// Build response
	mentorResponse := ChefMentorSessionResponse{
		SessionID:        sessionIDStr,
		Message:          enhancedMessage, // Use enhanced message with commentary
		Recipe:           currentRecipe,
		NextQuestion:     nextQuestion,
		IsComplete:       isComplete,
		SuggestedActions: getSuggestedActions(currentRecipe, dbSession.Language),
		QuickReplies:     quickReplies,
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status": "success",
		"data":   mentorResponse,
	})
}

// GetSessionHandler retrieves session state from database
// GET /api/ai/chef-mentor/session?id=xxx
func GetSessionHandler(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("id")
	if sessionID == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Session ID required")
		return
	}

	// Get session from database
	dbSession, err := chefMentorRepo.GetSession(sessionID)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Session not found")
		return
	}

	// Get messages
	dbMessages, err := chefMentorRepo.GetMessages(sessionID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to load messages")
		return
	}

	// Parse recipe
	currentRecipe := &RecipeDraft{}
	if len(dbSession.Recipe) > 0 {
		recipeJSON, _ := json.Marshal(dbSession.Recipe)
		json.Unmarshal(recipeJSON, currentRecipe)
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"sessionId":    dbSession.ID.String(),
			"language":     dbSession.Language,
			"recipe":       currentRecipe,
			"history":      dbMessages,
			"isComplete":   dbSession.IsComplete,
			"createdAt":    dbSession.CreatedAt,
			"lastActivity": dbSession.LastActivity,
		},
	})
}

// DeleteSessionHandler deletes session from database
// DELETE /api/ai/chef-mentor/session?id=xxx
func DeleteSessionHandler(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("id")
	if sessionID == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Session ID required")
		return
	}

	// Delete from database
	if err := chefMentorRepo.DeleteSession(sessionID); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to delete session")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "success",
		"message": "Session deleted",
	})
}

// smartExtractRecipeUpdates - improved extraction with NLP-like logic
func smartExtractRecipeUpdates(aiResponse string, currentRecipe *RecipeDraft, userMessage string, language string) *RecipeDraft {
	if currentRecipe == nil {
		currentRecipe = &RecipeDraft{}
	}

	// Call original extraction
	currentRecipe = extractRecipeUpdates(aiResponse, currentRecipe, userMessage)

	// Enhanced extraction logic
	lowerMsg := strings.ToLower(userMessage)

	// Difficulty extraction (multi-language) - priority order
	difficultyChecks := []struct {
		keywords []string
		value    string
	}{
		{[]string{"легк", "easy", "łatw"}, "easy"},
		{[]string{"серед", "intermediate", "średni"}, "intermediate"},
		{[]string{"склад", "висок", "hard", "trudny", "difficult"}, "hard"},
	}

	for _, check := range difficultyChecks {
		for _, keyword := range check.keywords {
			if contains(lowerMsg, keyword) {
				currentRecipe.Difficulty = check.value
				goto difficultyDone
			}
		}
	}
difficultyDone:

	// Time extraction - улучшенная логика с regex
	if contains(lowerMsg, "хв") || contains(lowerMsg, "минут") || contains(lowerMsg, "min") {
		re := regexp.MustCompile(`(\d+)\s*(хв|минут|min)`)
		if match := re.FindStringSubmatch(userMessage); len(match) > 1 {
			var time int
			fmt.Sscanf(match[1], "%d", &time)
			if time > 0 && time < 500 {
				currentRecipe.Time = time
			}
		}
	}

	// Portions extraction с regex
	if contains(lowerMsg, "порц") || contains(lowerMsg, "serving") || contains(lowerMsg, "porcj") {
		re := regexp.MustCompile(`(\d+)\s*(порц|serving|porcj)`)
		if match := re.FindStringSubmatch(userMessage); len(match) > 1 {
			var portions int
			fmt.Sscanf(match[1], "%d", &portions)
			if portions > 0 && portions < 100 {
				currentRecipe.Portions = portions
			}
		}
	}

	// Ingredients extraction - parse "рис 100г, вугор 200г, норі 1 шт"
	ingredients := parseIngredients(userMessage, language)
	if len(ingredients) > 0 {
		// Append new ingredients (avoid duplicates)
		for _, newIng := range ingredients {
			found := false
			for i, existing := range currentRecipe.Ingredients {
				if strings.EqualFold(existing.Name, newIng.Name) {
					// Update existing ingredient
					currentRecipe.Ingredients[i] = newIng
					found = true
					break
				}
			}
			if !found {
				currentRecipe.Ingredients = append(currentRecipe.Ingredients, newIng)
			}
		}
	}

	return currentRecipe
}

// parseIngredients extracts ingredients from user message
// Supports formats: "рис 100г", "вугор 200 г", "норі 1 шт", "масло 50мл"
func parseIngredients(message string, language string) []RecipeIngredient {
	var ingredients []RecipeIngredient

	// Regex pattern: ingredient_name + number + unit
	// Examples: "рис 100г", "вугор 200 г", "норі 1 шт"
	re := regexp.MustCompile(`([а-яіїєґa-z]+)\s+(\d+(?:\.\d+)?)\s*([а-яa-zїієґ]+)?`)
	matches := re.FindAllStringSubmatch(strings.ToLower(message), -1)

	for _, match := range matches {
		if len(match) >= 3 {
			name := match[1]
			var amount float64
			fmt.Sscanf(match[2], "%f", &amount)
			
			unit := ""
			if len(match) > 3 && match[3] != "" {
				unit = match[3]
			}

			// Filter out common words that aren't ingredients
			if isLikelyIngredient(name, language) {
				// Auto-calculate gross and net
				gross := amount
				net := amount * 0.95 // 5% standard cooking loss
				
				// Adjust loss factor based on ingredient type
				lossFactor := getCookingLossFactor(name, unit)
				if lossFactor > 0 {
					net = amount * lossFactor
				}

				ingredients = append(ingredients, RecipeIngredient{
					Name:   name,
					Amount: amount,
					Unit:   unit,
					Gross:  gross,
					Net:    net,
				})
			}
		}
	}

	return ingredients
}

// getCookingLossFactor returns cooking loss coefficient for ingredient
func getCookingLossFactor(ingredient string, unit string) float64 {
	ingredient = strings.ToLower(ingredient)
	
	// Different ingredients have different loss factors
	lossFactors := map[string]float64{
		// Vegetables (high water loss)
		"морква":   0.85, // carrot
		"цибуля":   0.80, // onion
		"капуста":  0.75, // cabbage
		"картопля": 0.85, // potato
		"огірок":   0.98, // cucumber (minimal loss, often raw)
		
		// Proteins (medium loss)
		"м'ясо":  0.70, // meat
		"курка":  0.75, // chicken
		"риба":   0.80, // fish
		"вугор":  0.85, // eel
		"креветки": 0.70, // shrimp
		
		// Grains (absorb water, increase weight)
		"рис":    1.20, // rice (absorbs water)
		"макарони": 1.30, // pasta
		"греча":  1.25, // buckwheat
		
		// No cooking (100% yield)
		"норі":     1.00, // nori
		"авокадо":  0.95, // avocado (just peeling loss)
		"сир":      0.98, // cheese
		"масло":    1.00, // oil
		"соус":     1.00, // sauce
	}
	
	if factor, exists := lossFactors[ingredient]; exists {
		return factor
	}
	
	// Default: 5% loss for unknown ingredients
	return 0.95
}

// calculateRecipeMetrics calculates calories, cost, and yield for recipe
func calculateRecipeMetrics(recipe *RecipeDraft) {
	if len(recipe.Ingredients) == 0 {
		return
	}

	totalCalories := 0.0
	totalProtein := 0.0
	totalFats := 0.0
	totalCarbs := 0.0
	totalCost := 0.0
	totalYield := 0.0

	for _, ing := range recipe.Ingredients {
		// Get nutrition data per 100g
		nutrition := getNutritionData(ing.Name)
		
		// Calculate based on net weight (in grams)
		netGrams := ing.Net
		if ing.Unit == "кг" || ing.Unit == "kg" {
			netGrams = ing.Net * 1000
		} else if ing.Unit == "мл" || ing.Unit == "ml" {
			netGrams = ing.Net // assume 1ml = 1g for liquids
		} else if ing.Unit == "шт" {
			netGrams = ing.Net * 50 // assume 1 piece = 50g average
		}

		// Calculate nutrition (per 100g base)
		multiplier := netGrams / 100.0
		totalCalories += nutrition.Calories * multiplier
		totalProtein += nutrition.Protein * multiplier
		totalFats += nutrition.Fats * multiplier
		totalCarbs += nutrition.Carbs * multiplier
		
		// Calculate cost
		totalCost += getCostPerGram(ing.Name) * netGrams
		
		// Sum yield
		totalYield += netGrams
	}

	// Update recipe
	recipe.Calories = int(totalCalories)
	recipe.Protein = totalProtein
	recipe.Fats = totalFats
	recipe.Carbs = totalCarbs
	recipe.Cost = totalCost
	recipe.Yield = int(totalYield)
	recipe.NetWeight = int(totalYield)
	recipe.GrossWeight = int(totalYield * 1.05) // Add 5% for gross
}

// NutritionData represents nutrition facts per 100g
type NutritionData struct {
	Calories float64
	Protein  float64
	Fats     float64
	Carbs    float64
}

// getNutritionData returns nutrition facts per 100g for ingredient
func getNutritionData(ingredient string) NutritionData {
	ingredient = strings.ToLower(ingredient)
	
	nutritionDB := map[string]NutritionData{
		// Grains
		"рис":      {Calories: 130, Protein: 2.7, Fats: 0.3, Carbs: 28.0},
		"макарони": {Calories: 157, Protein: 5.8, Fats: 0.9, Carbs: 30.9},
		"греча":    {Calories: 123, Protein: 4.5, Fats: 1.6, Carbs: 25.0},
		
		// Proteins
		"вугор":    {Calories: 184, Protein: 18.4, Fats: 11.8, Carbs: 0.0},
		"лосось":   {Calories: 208, Protein: 20.0, Fats: 13.0, Carbs: 0.0},
		"курка":    {Calories: 165, Protein: 31.0, Fats: 3.6, Carbs: 0.0},
		"свинина":  {Calories: 242, Protein: 16.0, Fats: 21.0, Carbs: 0.0},
		"креветки": {Calories: 99, Protein: 24.0, Fats: 0.3, Carbs: 0.2},
		
		// Vegetables
		"огірок":   {Calories: 15, Protein: 0.8, Fats: 0.1, Carbs: 3.6},
		"авокадо":  {Calories: 160, Protein: 2.0, Fats: 14.7, Carbs: 8.5},
		"морква":   {Calories: 41, Protein: 0.9, Fats: 0.2, Carbs: 9.6},
		"цибуля":   {Calories: 40, Protein: 1.1, Fats: 0.1, Carbs: 9.3},
		"картопля": {Calories: 77, Protein: 2.0, Fats: 0.1, Carbs: 17.0},
		
		// Seaweed & Others
		"норі":  {Calories: 35, Protein: 5.8, Fats: 0.3, Carbs: 5.1},
		"сир":   {Calories: 356, Protein: 24.0, Fats: 29.0, Carbs: 0.5},
		"масло": {Calories: 884, Protein: 0.0, Fats: 100.0, Carbs: 0.0},
		"соус":  {Calories: 50, Protein: 1.0, Fats: 2.0, Carbs: 7.0},
	}
	
	if data, exists := nutritionDB[ingredient]; exists {
		return data
	}
	
	// Default: medium calorie ingredient
	return NutritionData{Calories: 100, Protein: 5.0, Fats: 3.0, Carbs: 15.0}
}

// getCostPerGram returns cost in UAH per gram for ingredient
func getCostPerGram(ingredient string) float64 {
	ingredient = strings.ToLower(ingredient)
	
	// Prices in UAH per 100g (approximate Ukrainian market prices)
	pricesPerKg := map[string]float64{
		"рис":      0.040, // 40 грн/кг
		"вугор":    0.450, // 450 грн/кг
		"лосось":   0.350, // 350 грн/кг
		"норі":     1.200, // 1200 грн/кг (expensive)
		"авокадо":  0.180, // 180 грн/кг
		"огірок":   0.050, // 50 грн/кг
		"курка":    0.120, // 120 грн/кг
		"свинина":  0.160, // 160 грн/кг
		"креветки": 0.400, // 400 грн/кг
		"сир":      0.250, // 250 грн/кг
		"масло":    0.100, // 100 грн/л
		"морква":   0.025, // 25 грн/кг
		"цибуля":   0.020, // 20 грн/кг
	}
	
	if price, exists := pricesPerKg[ingredient]; exists {
		return price / 1000.0 // convert to per gram
	}
	
	// Default: 100 грн/кг = 0.1 грн/г
	return 0.0001
}

// generateNutritionCommentary creates human-like commentary about recipe metrics
func generateNutritionCommentary(recipe *RecipeDraft, language string) string {
	if recipe.Calories == 0 || len(recipe.Ingredients) == 0 {
		return ""
	}

	templates := map[string]string{
		"ua": "📊 Цей рецепт має приблизно %d ккал, вихід %d г, собівартість %.2f грн.",
		"en": "📊 This recipe has approximately %d kcal, yield %d g, cost %.2f UAH.",
		"ru": "📊 Этот рецепт имеет примерно %d ккал, выход %d г, себестоимость %.2f грн.",
		"pl": "📊 Ten przepis ma około %d kcal, wydajność %d g, koszt %.2f UAH.",
	}

	template := templates[language]
	if template == "" {
		template = templates["ua"]
	}

	return fmt.Sprintf(template, recipe.Calories, recipe.Yield, recipe.Cost)
}

// saveCompletedRecipeToDatabase saves completed recipe to AI Culinary OS
func saveCompletedRecipeToDatabase(sessionID uuid.UUID, userID *uuid.UUID, recipe *RecipeDraft, language string) (string, error) {
	aiRecipeRepo := database.NewAIRecipeRepository()
	
	// Convert RecipeDraft to AIGeneratedRecipe
	aiRecipe, err := database.ConvertRecipeDraftToAI(recipe, sessionID, userID, language)
	if err != nil {
		return "", fmt.Errorf("failed to convert recipe: %w", err)
	}
	
	// Generate NEW UUID for recipe (before saving)
	aiRecipe.ID = uuid.New()
	
	// Generate unique share URL based on new ID
	aiRecipe.ShareURL = fmt.Sprintf("recipe-%s", aiRecipe.ID.String()[:8])
	
	// Save to database
	if err := aiRecipeRepo.SaveRecipe(aiRecipe); err != nil {
		return "", fmt.Errorf("failed to save recipe: %w", err)
	}
	
	return aiRecipe.ID.String(), nil
}

// isRecipeComplete checks if recipe has minimum required fields
func isRecipeComplete(recipe *RecipeDraft) bool {
	if recipe == nil {
		return false
	}
	
	// Recipe is complete if it has:
	// 1. Title (and not default title)
	// 2. Category
	// 3. Difficulty  
	// 4. At least one ingredient
	hasTitle := recipe.Title != "" && 
		recipe.Title != "Новий рецепт" && 
		recipe.Title != "New Recipe" &&
		!strings.Contains(strings.ToLower(recipe.Title), "хочу") // "Хочу приготувати..." is not a real title
	
	hasCategory := recipe.Category != ""
	hasDifficulty := recipe.Difficulty != ""
	hasIngredients := len(recipe.Ingredients) > 0
	
	return hasTitle && hasCategory && hasDifficulty && hasIngredients
}

// detectUserIntent handles greetings, commands, and context before AI processing
func detectUserIntent(message string, recipe *RecipeDraft, language string) string {
	lower := strings.ToLower(strings.TrimSpace(message))
	
	// 1. Greetings (don't treat as recipe name!)
	greetings := []string{
		"привіт", "привет", "вітаю", "hello", "hi", "hey", "здравствуй", "добрий день",
		"доброго дня", "good morning", "good afternoon", "good evening", "hola", "cześć",
	}
	
	for _, greeting := range greetings {
		if lower == greeting || strings.HasPrefix(lower, greeting+" ") || strings.HasPrefix(lower, greeting+"!") {
			responses := map[string]string{
				"ua": "👋 Вітаю! Я — Шеф Діма, ваш кулінарний AI-помічник.\n\n🍣 Що будемо готувати сьогодні? Напишіть назву страви або інгредієнти, які у вас є!",
				"en": "👋 Hello! I'm Chef Dima, your culinary AI assistant.\n\n🍣 What shall we cook today? Tell me the dish name or ingredients you have!",
				"ru": "👋 Привет! Я — Шеф Дима, ваш кулинарный AI-помощник.\n\n🍣 Что будем готовить сегодня? Напишите название блюда или ингредиенты!",
				"pl": "👋 Cześć! Jestem Szef Dima, twój kulinarny asystent AI.\n\n🍣 Co będziemy dziś gotować?",
			}
			if resp, ok := responses[language]; ok {
				return resp
			}
			return responses["ua"]
		}
	}
	
	// 2. Help commands
	helpKeywords := []string{"допомога", "help", "помощь", "допоможи", "как", "що можеш", "what can"}
	for _, keyword := range helpKeywords {
		if strings.Contains(lower, keyword) {
			responses := map[string]string{
				"ua": "🧑‍🍳 **Як я можу допомогти:**\n\n✅ Створити рецепт з ваших інгредієнтів\n✅ Порахувати калорії та вартість\n✅ Підібрати складність та час\n✅ Автоматично зберегти рецепт\n\n💬 Просто напишіть назву страви або інгредієнти!",
				"en": "🧑‍🍳 **How I can help:**\n\n✅ Create recipe from your ingredients\n✅ Calculate calories and cost\n✅ Suggest difficulty and time\n✅ Auto-save your recipe\n\n💬 Just write dish name or ingredients!",
				"ru": "🧑‍🍳 **Как я могу помочь:**\n\n✅ Создать рецепт из ваших ингредиентов\n✅ Посчитать калории и стоимость\n✅ Подобрать сложность и время\n✅ Автосохранение рецепта\n\n💬 Просто напишите название блюда!",
			}
			if resp, ok := responses[language]; ok {
				return resp
			}
			return responses["ua"]
		}
	}
	
	// 3. Thank you / Goodbye
	thanksKeywords := []string{"дякую", "спасибо", "thanks", "thank you", "dziękuję"}
	for _, keyword := range thanksKeywords {
		if strings.Contains(lower, keyword) {
			responses := map[string]string{
				"ua": "😊 Будь ласка! Смачного! Якщо потрібно ще щось — звертайтесь!",
				"en": "😊 You're welcome! Enjoy your meal! Let me know if you need anything else!",
				"ru": "😊 Пожалуйста! Приятного аппетита! Обращайтесь ещё!",
			}
			if resp, ok := responses[language]; ok {
				return resp
			}
			return responses["ua"]
		}
	}
	
	goodbyeKeywords := []string{"бувай", "пока", "bye", "goodbye", "до побачення", "до свидания"}
	for _, keyword := range goodbyeKeywords {
		if strings.Contains(lower, keyword) {
			responses := map[string]string{
				"ua": "👋 До зустрічі! Приходьте ще за новими рецептами! 🍣",
				"en": "👋 Goodbye! Come back for more recipes! 🍣",
				"ru": "👋 До свидания! Приходите ещё за новыми рецептами! 🍣",
			}
			if resp, ok := responses[language]; ok {
				return resp
			}
			return responses["ua"]
		}
	}
	
	// 4. Single letter or very short non-recipe input
	if len(lower) <= 2 && !unicode.IsDigit(rune(lower[0])) {
		responses := map[string]string{
			"ua": "🤔 Вибачте, не зрозумів. Напишіть, будь ласка, назву страви або інгредієнти!",
			"en": "🤔 Sorry, I didn't understand. Please write dish name or ingredients!",
			"ru": "🤔 Извините, не понял. Напишите название блюда или ингредиенты!",
		}
		if resp, ok := responses[language]; ok {
			return resp
		}
		return responses["ua"]
	}
	
	// No special context detected - proceed with AI recipe logic
	return ""
}

// isLikelyIngredient filters out common non-ingredient words
func isLikelyIngredient(word string, language string) bool {
	// Exclude common words
	excludeUA := []string{"має", "треба", "потрібно", "скільки", "додати", "взяти", "використати"}
	excludeEN := []string{"need", "add", "use", "take", "how", "much", "many"}
	excludeRU := []string{"нужно", "надо", "добавить", "взять", "сколько"}
	
	word = strings.ToLower(word)
	
	for _, excluded := range append(append(excludeUA, excludeEN...), excludeRU...) {
		if word == excluded {
			return false
		}
	}
	
	// Must be at least 3 characters
	return len(word) >= 3
}

// generateQuickReplies creates context-aware quick reply buttons
func generateQuickReplies(recipe *RecipeDraft, language string) []string {
	replies := map[string][]string{
		"ua": {
			"4 порції",
			"Середня складність",
			"30 хвилин",
			"Додати інгредієнти",
			"Пропустити",
		},
		"en": {
			"4 servings",
			"Intermediate difficulty",
			"30 minutes",
			"Add ingredients",
			"Skip",
		},
		"ru": {
			"4 порции",
			"Средняя сложность",
			"30 минут",
			"Добавить ингредиенты",
			"Пропустить",
		},
		"pl": {
			"4 porcje",
			"Średnia trudność",
			"30 minut",
			"Dodaj składniki",
			"Pomiń",
		},
	}

	langReplies, ok := replies[language]
	if !ok {
		langReplies = replies["ua"]
	}

	// Contexual filtering
	if recipe.Portions > 0 {
		// Remove portions suggestion
		langReplies = filter(langReplies, func(s string) bool {
			return !contains(s, "порц") && !contains(s, "serving") && !contains(s, "porcj")
		})
	}

	if recipe.Difficulty != "" {
		// Remove difficulty suggestion
		langReplies = filter(langReplies, func(s string) bool {
			return !contains(s, "складн") && !contains(s, "difficulty") && !contains(s, "trudność")
		})
	}

	if recipe.Time > 0 {
		// Remove time suggestion
		langReplies = filter(langReplies, func(s string) bool {
			return !contains(s, "хв") && !contains(s, "minutes") && !contains(s, "minut")
		})
	}

	// Limit to 5 suggestions
	if len(langReplies) > 5 {
		langReplies = langReplies[:5]
	}

	return langReplies
}

// Helper: case-insensitive contains
func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		(s == substr || 
		 s[:len(substr)] == substr ||
		 len(s) > len(substr) && s[len(s)-len(substr):] == substr ||
		 findInString(s, substr))
}

func findInString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Helper: filter slice
func filter(slice []string, predicate func(string) bool) []string {
	result := []string{}
	for _, item := range slice {
		if predicate(item) {
			result = append(result, item)
		}
	}
	return result
}

// CleanupOldSessions removes old sessions from database (run periodically)
func CleanupOldSessions() {
	deleted, err := chefMentorRepo.DeleteOldSessions(24 * time.Hour)
	if err != nil {
		fmt.Printf("❌ Failed to cleanup old sessions: %v\n", err)
		return
	}
	if deleted > 0 {
		fmt.Printf("🗑️ Cleaned up %d old Chef Mentor sessions\n", deleted)
	}
}

// init starts cleanup goroutine
func init() {
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		for range ticker.C {
			CleanupOldSessions()
		}
	}()
}
