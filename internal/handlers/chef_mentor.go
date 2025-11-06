package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/ai"
	"github.com/dmitrijfomin/menu-fodifood/backend/pkg/utils"
)

// ChefMentorRequest represents a user message in the conversation
type ChefMentorRequest struct {
	Message         string                 `json:"message"`          // User's message
	Language        string                 `json:"language"`         // ui, en, ru, pl
	ConversationHistory []ConversationMessage `json:"history,omitempty"` // Previous messages
	CurrentRecipe   *RecipeDraft           `json:"currentRecipe,omitempty"` // Draft being built
}

// ConversationMessage represents one message in the chat
type ConversationMessage struct {
	Role    string `json:"role"`    // "user" or "assistant"
	Content string `json:"content"` // Message text
}

// RecipeDraft represents the recipe being built step-by-step
type RecipeDraft struct {
	Title        string             `json:"title,omitempty"`
	Description  string             `json:"description,omitempty"`
	Category     string             `json:"category,omitempty"`
	Difficulty   string             `json:"difficulty,omitempty"`
	Time         int                `json:"time,omitempty"`
	Portions     int                `json:"portions,omitempty"`
	Ingredients  []RecipeIngredient `json:"ingredients,omitempty"`
	Steps        []string           `json:"steps,omitempty"`
	GrossWeight  int                `json:"grossWeight,omitempty"`
	NetWeight    int                `json:"netWeight,omitempty"`
	Calories     int                `json:"calories,omitempty"`
	Protein      float64            `json:"protein,omitempty"`
	Fats         float64            `json:"fats,omitempty"`
	Carbs        float64            `json:"carbs,omitempty"`
	Yield        int                `json:"yield,omitempty"`
	Cost         float64            `json:"cost,omitempty"`
	TokensReward int                `json:"tokensReward,omitempty"`
	IsComplete   bool               `json:"isComplete,omitempty"`
}

// ChefMentorResponse represents the assistant's response
type ChefMentorResponse struct {
	Message       string       `json:"message"`       // Assistant's response
	Recipe        *RecipeDraft `json:"recipe"`        // Updated recipe draft
	NextQuestion  string       `json:"nextQuestion"`  // Suggested next question
	IsComplete    bool         `json:"isComplete"`    // Recipe is ready
	SuggestedActions []string  `json:"suggestedActions,omitempty"` // Quick actions
}

// ChefMentorHandler is the interactive AI chef assistant
// POST /api/ai/chef-mentor
func ChefMentorHandler(w http.ResponseWriter, r *http.Request) {
	var req ChefMentorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid input")
		return
	}

	// Default language to Ukrainian
	if req.Language == "" {
		req.Language = "ua"
	}

	// Initialize recipe draft if not exists
	if req.CurrentRecipe == nil {
		req.CurrentRecipe = &RecipeDraft{}
	}

	// Build system prompt based on language
	systemPrompt := buildMentorSystemPrompt(req.Language)
	
	// Build context with recipe state
	contextPrompt := buildRecipeContext(req.CurrentRecipe, req.Language)

	// Create conversation for AI
	client := ai.NewGroqClient()
	
	messages := []ai.GroqMessage{
		{
			Role:    "system",
			Content: systemPrompt + "\n\n" + contextPrompt,
		},
	}

	// Add conversation history
	for _, msg := range req.ConversationHistory {
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

	// Get AI response
	response, err := client.Chat(messages, 0.7, 1000)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "AI service error")
		return
	}

	assistantMessage := response.Choices[0].Message.Content

	// Try to extract recipe updates from AI response
	updatedRecipe := extractRecipeUpdates(assistantMessage, req.CurrentRecipe, req.Message)

	// Determine next question and completion status
	nextQuestion := determineNextQuestion(updatedRecipe, req.Language)
	isComplete := isRecipeComplete(updatedRecipe)

	// Build response
	mentorResponse := ChefMentorResponse{
		Message:          assistantMessage,
		Recipe:           updatedRecipe,
		NextQuestion:     nextQuestion,
		IsComplete:       isComplete,
		SuggestedActions: getSuggestedActions(updatedRecipe, req.Language),
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status": "success",
		"data":   mentorResponse,
	})
}

// buildMentorSystemPrompt creates the AI Chef Mentor system prompt
func buildMentorSystemPrompt(language string) string {
	prompts := map[string]string{
		"ua": `Ти — **AI Chef Mentor**, офіційний асистент **Кулінарної Академії Діми Фоміна**.
Твоя мета — допомогти користувачу створити професійний, структурований рецепт крок за кроком, як справжній шеф навчає студента.

ПРАВИЛА СПІЛКУВАННЯ:
- Завжди розмовляй природньо та дружньо ("Чудово! Додаймо наступний крок.")
- Відповідай коротко (2-4 речення)
- Задавай по одному чіткому питанню за раз
- Коли користувач дає інформацію, підтверджуй та переходь до наступного кроку
- Якщо інформації достатньо, пропонуй завершити рецепт

ЩО ПОТРІБНО ЗІБРАТИ:
1. Назва страви
2. Категорія (sushi, ramen, desserts, fusion, тощо)
3. Складність (easy, intermediate, hard)
4. Час приготування (хвилини)
5. Кількість порцій
6. Інгредієнти (назва, кількість, одиниця)
7. Кроки приготування (детально)

НЕ ГЕНЕРУЙ JSON У ВІДПОВІДІ - просто веди діалог природньо.`,

		"en": `You are **AI Chef Mentor**, the official assistant of **Dima Fomin's AI Culinary Academy**.
Your goal is to help users create professional, structured recipes step-by-step — just like a chef would teach a student.

COMMUNICATION RULES:
- Always speak naturally and friendly ("Great! Let's add the next step.")
- Keep answers short (2-4 sentences)
- Ask one clear question at a time
- When user provides info, acknowledge and move to next step
- When enough info is gathered, offer to complete the recipe

WHAT TO COLLECT:
1. Dish name
2. Category (sushi, ramen, desserts, fusion, etc.)
3. Difficulty (easy, intermediate, hard)
4. Cooking time (minutes)
5. Servings
6. Ingredients (name, amount, unit)
7. Cooking steps (detailed)

DO NOT generate JSON in responses - just have a natural conversation.`,

		"ru": `Ты — **AI Chef Mentor**, официальный ассистент **Кулинарной Академии Димы Фомина**.
Твоя цель — помочь пользователю создать профессиональный, структурированный рецепт шаг за шагом, как настоящий шеф учит студента.

ПРАВИЛА ОБЩЕНИЯ:
- Всегда говори естественно и дружелюбно ("Отлично! Добавим следующий шаг.")
- Отвечай коротко (2-4 предложения)
- Задавай по одному чёткому вопросу за раз
- Когда пользователь даёт информацию, подтверждай и переходи к следующему шагу
- Когда информации достаточно, предложи завершить рецепт

ЧТО НУЖНО СОБРАТЬ:
1. Название блюда
2. Категория (sushi, ramen, desserts, fusion, и т.д.)
3. Сложность (easy, intermediate, hard)
4. Время приготовления (минуты)
5. Количество порций
6. Ингредиенты (название, количество, единица)
7. Шаги приготовления (детально)

НЕ ГЕНЕРИРУЙ JSON В ОТВЕТАХ - просто веди диалог естественно.`,

		"pl": `Jesteś **AI Chef Mentor**, oficjalnym asystentem **Akademii Kulinarnej Dimy Fomina**.
Twoim celem jest pomóc użytkownikowi stworzyć profesjonalny, ustrukturyzowany przepis krok po kroku — tak jak prawdziwy szef uczy ucznia.

ZASADY KOMUNIKACJI:
- Zawsze mów naturalnie i przyjaźnie ("Świetnie! Dodajmy kolejny krok.")
- Odpowiadaj krótko (2-4 zdania)
- Zadawaj po jednym jasnym pytaniu na raz
- Gdy użytkownik poda informacje, potwierdź i przejdź do następnego kroku
- Gdy wystarczy informacji, zaproponuj ukończenie przepisu

CO TRZEBA ZEBRAĆ:
1. Nazwa dania
2. Kategoria (sushi, ramen, desery, fusion, itp.)
3. Trudność (łatwy, średni, trudny)
4. Czas przygotowania (minuty)
5. Liczba porcji
6. Składniki (nazwa, ilość, jednostka)
7. Kroki przygotowania (szczegółowo)

NIE GENERUJ JSON W ODPOWIEDZIACH - po prostu prowadź naturalną rozmowę.`,
	}

	prompt, ok := prompts[language]
	if !ok {
		prompt = prompts["ua"] // Default to Ukrainian
	}
	return prompt
}

// buildRecipeContext shows AI the current recipe state
func buildRecipeContext(recipe *RecipeDraft, language string) string {
	if recipe == nil || (recipe.Title == "" && len(recipe.Ingredients) == 0) {
		return "ПОТОЧНИЙ СТАН РЕЦЕПТУ: Порожній, збір інформації тільки починається."
	}

	var context strings.Builder
	context.WriteString("ПОТОЧНИЙ СТАН РЕЦЕПТУ:\n")
	
	if recipe.Title != "" {
		context.WriteString(fmt.Sprintf("- Назва: %s\n", recipe.Title))
	}
	if recipe.Category != "" {
		context.WriteString(fmt.Sprintf("- Категорія: %s\n", recipe.Category))
	}
	if recipe.Difficulty != "" {
		context.WriteString(fmt.Sprintf("- Складність: %s\n", recipe.Difficulty))
	}
	if recipe.Time > 0 {
		context.WriteString(fmt.Sprintf("- Час: %d хв\n", recipe.Time))
	}
	if recipe.Portions > 0 {
		context.WriteString(fmt.Sprintf("- Порцій: %d\n", recipe.Portions))
	}
	if len(recipe.Ingredients) > 0 {
		context.WriteString(fmt.Sprintf("- Інгредієнти: %d штук\n", len(recipe.Ingredients)))
	}
	if len(recipe.Steps) > 0 {
		context.WriteString(fmt.Sprintf("- Кроки: %d штук\n", len(recipe.Steps)))
	}

	context.WriteString("\nНА ОСНОВІ ЦЬОГО продовжуй діалог природньо.")
	return context.String()
}

// extractRecipeUpdates tries to find recipe info in user message
func extractRecipeUpdates(aiResponse string, currentRecipe *RecipeDraft, userMessage string) *RecipeDraft {
	if currentRecipe == nil {
		currentRecipe = &RecipeDraft{}
	}

	// Simple extraction logic (can be improved with NLP)
	lowerMsg := strings.ToLower(userMessage)
	
	// Detect title from longer descriptions
	if currentRecipe.Title == "" && len(userMessage) > 10 {
		// If message looks like a dish description, extract potential title
		if strings.Contains(lowerMsg, "рол") || strings.Contains(lowerMsg, "роли") ||
		   strings.Contains(lowerMsg, "суші") || strings.Contains(lowerMsg, "хочу зробити") {
			currentRecipe.Title = userMessage
		}
	}

	// Detect portions
	if strings.Contains(lowerMsg, "порц") {
		// Extract number before "порцій" or after
		var portions int
		fmt.Sscanf(userMessage, "%d", &portions)
		if portions > 0 {
			currentRecipe.Portions = portions
		}
	}

	// Detect time
	if strings.Contains(lowerMsg, "хв") || strings.Contains(lowerMsg, "минут") {
		var time int
		fmt.Sscanf(userMessage, "%d", &time)
		if time > 0 {
			currentRecipe.Time = time
		}
	}

	// Category detection
	categories := map[string]string{
		"суші": "sushi",
		"sushi": "sushi",
		"рол": "sushi",
		"рамен": "ramen",
		"десерт": "desserts",
		"салат": "salad",
	}
	
	for keyword, category := range categories {
		if strings.Contains(lowerMsg, keyword) {
			currentRecipe.Category = category
			break
		}
	}

	return currentRecipe
}

// determineNextQuestion suggests what to ask next
func determineNextQuestion(recipe *RecipeDraft, language string) string {
	questions := map[string][]string{
		"ua": {
			"Яка назва вашої страви?",
			"Яка категорія цієї страви? (суші, рамен, десерт, тощо)",
			"Яка складність приготування? (легка, середня, висока)",
			"Скільки часу потрібно на приготування? (у хвилинах)",
			"На скільки порцій розрахований рецепт?",
			"Які інгредієнти потрібні?",
			"Опишіть кроки приготування",
		},
		"en": {
			"What's the name of your dish?",
			"What category is this dish? (sushi, ramen, dessert, etc.)",
			"What's the difficulty level? (easy, intermediate, hard)",
			"How long does it take to prepare? (in minutes)",
			"How many servings does this recipe make?",
			"What ingredients are needed?",
			"Describe the cooking steps",
		},
	}

	langQuestions, ok := questions[language]
	if !ok {
		langQuestions = questions["ua"]
	}

	// Determine what's missing
	if recipe.Title == "" {
		return langQuestions[0]
	}
	if recipe.Category == "" {
		return langQuestions[1]
	}
	if recipe.Difficulty == "" {
		return langQuestions[2]
	}
	if recipe.Time == 0 {
		return langQuestions[3]
	}
	if recipe.Portions == 0 {
		return langQuestions[4]
	}
	if len(recipe.Ingredients) == 0 {
		return langQuestions[5]
	}
	if len(recipe.Steps) == 0 {
		return langQuestions[6]
	}

	return "Рецепт майже готовий! Бажаєте додати ще щось?"
}

// getSuggestedActions provides quick action buttons
func getSuggestedActions(recipe *RecipeDraft, language string) []string {
	actions := []string{}
	
	if recipe.Title == "" {
		return []string{"Розпочати новий рецепт", "Показати приклад"}
	}
	
	if len(recipe.Ingredients) == 0 {
		actions = append(actions, "Додати інгредієнт")
	}
	
	if len(recipe.Steps) == 0 {
		actions = append(actions, "Додати крок приготування")
	}
	
	if isRecipeComplete(recipe) {
		actions = append(actions, "Завершити рецепт", "Розрахувати калорії", "Оцінити вартість")
	}
	
	return actions
}
