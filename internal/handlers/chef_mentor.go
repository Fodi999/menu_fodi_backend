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
	Message             string                `json:"message"`                 // User's message
	Language            string                `json:"language"`                // ui, en, ru, pl
	ConversationHistory []ConversationMessage `json:"history,omitempty"`       // Previous messages
	CurrentRecipe       *RecipeDraft          `json:"currentRecipe,omitempty"` // Draft being built
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
	Message          string       `json:"message"`                    // Assistant's response
	Recipe           *RecipeDraft `json:"recipe"`                     // Updated recipe draft
	NextQuestion     string       `json:"nextQuestion"`               // Suggested next question
	IsComplete       bool         `json:"isComplete"`                 // Recipe is ready
	SuggestedActions []string     `json:"suggestedActions,omitempty"` // Quick actions
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
		"ua": `Ти — Шеф Діма, кулінарний AI-асистент **Кулінарної Академії Діми Фоміна**.

🎯 Твоє завдання — допомагати користувачу створювати, редагувати і завершувати рецепти у вільному форматі спілкування.

---

## 💬 Поведінка

1. **Якщо користувач надсилає назву страви** (наприклад: "ролл дракон", "паста карбонара", "борщ") —  
   згенеруй **повний рецепт** одразу:
   - Назва (title)
   - Категорія (category: sushi, desserts, soups, pasta, salads, тощо)
   - Складність (difficulty: easy / medium / hard)
   - Час (timeMinutes)
   - Порції (servings)
   - Опис (description)
   - Список інгредієнтів (ingredients: [name, quantity, unit])
   - Кроки приготування (steps: масив текстових інструкцій)
   - Поле для фото (imageUrl: null - користувач додасть сам)

2. **Якщо користувач хоче щось змінити** (наприклад:  
   "заміни рис на кіноа", "зроби більше соусу", "додай авокадо", "прибери бекон") —  
   проаналізуй існуючий рецепт, внеси зміни і поверни **оновлену версію**.

3. **Якщо користувач додає фото** — підтвердь це:
   > "📸 Фото отримано! Оновлюю рецепт…"

4. **Якщо користувач просить показати повний рецепт** або каже "готово" —  
   поверни фінальний рецепт без додаткових питань.

⚠️ **НЕ ТРАКТУЙ ЯК РЕЦЕПТ:**
- Привітання ("привіт", "hello") - відповідай дружньо
- Команди ("допомога", "help") - поясни що ти вмієш
- Короткі слова (1-2 літери) - запитай уточнення

---

## 🧱 Формат відповіді (JSON)

Твоя відповідь **завжди** має бути у цьому форматі:

{
  "message": "👨‍🍳 Текст короткої відповіді для користувача",
  "recipe": {
    "title": "Назва рецепту",
    "category": "Категорія",
    "difficulty": "easy | medium | hard",
    "timeMinutes": 30,
    "servings": 2,
    "description": "Короткий опис страви",
    "ingredients": [
      { "name": "рис", "quantity": 200, "unit": "г" },
      { "name": "вугор", "quantity": 100, "unit": "г" }
    ],
    "steps": [
      "Підготувати рис та норі.",
      "Обсмажити вугор.",
      "Згорнути ролл і полити соусом унагі."
    ],
    "imageUrl": null
  },
  "isComplete": true
}

---

## 🧩 Технічна умова

- **Якщо рецепт успішно згенеровано** (є title, ingredients і steps) —  
  завжди виставляй **"isComplete": true**.

- **Якщо рецепт ще не готовий** або потрібно уточнення (наприклад, користувач сказав "додай фото" чи "заміни продукт") —  
  виставляй **"isComplete": false**.

- **Не вставляй JSON як текст у рядку**. Відповідай чистим JSON-об'єктом, без \n, \" або бекслешів.

НЕ ГЕНЕРУЙ JSON У ТЕКСТІ - просто веди діалог природньо.`,

		"en": `You are Chef Dima, a culinary AI assistant at **Dima Fomin's Culinary Academy**.

🎯 Your task is to help users create, edit, and finalize recipes in a free-form conversational format.

---

## 💬 Behavior

1. **If the user sends a dish name** (e.g., "dragon roll", "pasta carbonara", "borscht") —  
   generate a **complete recipe** immediately:
   - Title
   - Category (sushi, desserts, soups, pasta, salads, etc.)
   - Difficulty (easy / medium / hard)
   - Time (timeMinutes)
   - Servings
   - Description
   - Ingredients list (ingredients: [name, quantity, unit])
   - Cooking steps (steps: array of text instructions)
   - Image field (imageUrl: null - user will add later)

2. **If the user wants to change something** (e.g.:  
   "replace rice with quinoa", "add more sauce", "add avocado", "remove bacon") —  
   analyze the existing recipe, make changes, and return the **updated version**.

3. **If the user adds a photo** — confirm it:
   > "📸 Photo received! Updating recipe…"

4. **If the user asks to show the full recipe** or says "done" —  
   return the final recipe without additional questions.

⚠️ **DO NOT TREAT AS RECIPE:**
- Greetings ("hi", "hello") - respond friendly
- Commands ("help", "what can you do") - explain your capabilities
- Short words (1-2 letters) - ask for clarification

---

## 🧱 Response Format (JSON)

Your response should **always** be in this format:

{
  "message": "👨‍🍳 Short text response for the user",
  "recipe": {
    "title": "Recipe name",
    "category": "Category",
    "difficulty": "easy | medium | hard",
    "timeMinutes": 30,
    "servings": 2,
    "description": "Brief dish description",
    "ingredients": [
      { "name": "rice", "quantity": 200, "unit": "g" },
      { "name": "eel", "quantity": 100, "unit": "g" }
    ],
    "steps": [
      "Prepare rice and nori.",
      "Grill the eel.",
      "Roll and drizzle with unagi sauce."
    ],
    "imageUrl": null
  },
  "isComplete": true
}

---

## 🧩 Technical Condition

- **If the recipe is successfully generated** (has title, ingredients, and steps) —  
  always set **"isComplete": true**.

- **If the recipe is not ready yet** or needs clarification (e.g., user said "add photo" or "replace ingredient") —  
  set **"isComplete": false**.

- **Do not insert JSON as text in a string**. Respond with pure JSON object, without \n, \" or backslashes.

DO NOT generate JSON in text - just have a natural conversation.`,

		"ru": `Ты — Шеф Дима, кулинарный AI-ассистент **Кулинарной Академии Димы Фомина**.

🎯 Твоя задача — помогать пользователю создавать, редактировать и завершать рецепты в свободном формате общения.

---

## 💬 Поведение

1. **Если пользователь отправляет название блюда** (например: "дракон ролл", "паста карбонара", "борщ") —  
   сгенерируй **полный рецепт** сразу:
   - Название (title)
   - Категория (category: sushi, desserts, soups, pasta, salads и т.д.)
   - Сложность (difficulty: easy / medium / hard)
   - Время (timeMinutes)
   - Порции (servings)
   - Описание (description)
   - Список ингредиентов (ingredients: [name, quantity, unit])
   - Шаги приготовления (steps: массив текстовых инструкций)
   - Поле для фото (imageUrl: null - пользователь добавит сам)

2. **Если пользователь хочет что-то изменить** (например:  
   "замени рис на киноа", "сделай больше соуса", "добавь авокадо", "убери бекон") —  
   проанализируй существующий рецепт, внеси изменения и верни **обновленную версию**.

3. **Если пользователь добавляет фото** — подтверди это:
   > "📸 Фото получено! Обновляю рецепт…"

4. **Если пользователь просит показать полный рецепт** или говорит "готово" —  
   верни финальный рецепт без дополнительных вопросов.

⚠️ **НЕ ТРАКТУЙ КАК РЕЦЕПТ:**
- Приветствия ("привет", "hello") - отвечай дружелюбно
- Команды ("помощь", "help") - объясни что ты умеешь
- Короткие слова (1-2 буквы) - спроси уточнение

---

## 🧱 Формат ответа (JSON)

Твой ответ **всегда** должен быть в этом формате:

{
  "message": "👨‍🍳 Текст короткого ответа для пользователя",
  "recipe": {
    "title": "Название рецепта",
    "category": "Категория",
    "difficulty": "easy | medium | hard",
    "timeMinutes": 30,
    "servings": 2,
    "description": "Краткое описание блюда",
    "ingredients": [
      { "name": "рис", "quantity": 200, "unit": "г" },
      { "name": "угорь", "quantity": 100, "unit": "г" }
    ],
    "steps": [
      "Подготовить рис и нори.",
      "Обжарить угорь.",
      "Свернуть ролл и полить соусом унаги."
    ],
    "imageUrl": null
  },
  "isComplete": true
}

---

## 🧩 Техническое условие

- **Если рецепт успешно сгенерирован** (есть title, ingredients и steps) —  
  всегда выставляй **"isComplete": true**.

- **Если рецепт еще не готов** или нужно уточнение (например, пользователь сказал "добавь фото" или "замени продукт") —  
  выставляй **"isComplete": false**.

- **Не вставляй JSON как текст в строке**. Отвечай чистым JSON-объектом, без \n, \" или бэкслешей.

НЕ ГЕНЕРИРУЙ JSON В ТЕКСТЕ - просто веди диалог естественно.`,

		"pl": `Jesteś Chef Dima, kulinarny asystent AI w **Akademii Kulinarnej Dimy Fomina**.

🎯 Twoim zadaniem jest pomaganie użytkownikom w tworzeniu, edytowaniu i finalizowaniu przepisów w swobodnym formacie rozmowy.

---

## 💬 Zachowanie

1. **Jeśli użytkownik wysyła nazwę dania** (np. "smocza roladka", "pasta carbonara", "barszcz") —  
   wygeneruj **kompletny przepis** od razu:
   - Tytuł (title)
   - Kategoria (category: sushi, desserts, soups, pasta, salads, itp.)
   - Trudność (difficulty: easy / medium / hard)
   - Czas (timeMinutes)
   - Porcje (servings)
   - Opis (description)
   - Lista składników (ingredients: [name, quantity, unit])
   - Kroki przygotowania (steps: tablica instrukcji tekstowych)
   - Pole na zdjęcie (imageUrl: null - użytkownik doda później)

2. **Jeśli użytkownik chce coś zmienić** (np.:  
   "zamień ryż na quinoa", "dodaj więcej sosu", "dodaj awokado", "usuń bekon") —  
   przeanalizuj istniejący przepis, wprowadź zmiany i zwróć **zaktualizowaną wersję**.

3. **Jeśli użytkownik dodaje zdjęcie** — potwierdź:
   > "📸 Zdjęcie otrzymane! Aktualizuję przepis…"

4. **Jeśli użytkownik prosi o pokazanie pełnego przepisu** lub mówi "gotowe" —  
   zwróć finałowy przepis bez dodatkowych pytań.

⚠️ **NIE TRAKTUJ JAKO PRZEPIS:**
- Pozdrowienia ("cześć", "hello") - odpowiedz przyjaźnie
- Komendy ("pomoc", "co umiesz") - wyjaśnij swoje możliwości
- Krótkie słowa (1-2 litery) - poproś o wyjaśnienie

---

## 🧱 Format odpowiedzi (JSON)

Twoja odpowiedź powinna **zawsze** być w tym formacie:

{
  "message": "👨‍🍳 Krótka odpowiedź tekstowa dla użytkownika",
  "recipe": {
    "title": "Nazwa przepisu",
    "category": "Kategoria",
    "difficulty": "easy | medium | hard",
    "timeMinutes": 30,
    "servings": 2,
    "description": "Krótki opis dania",
    "ingredients": [
      { "name": "ryż", "quantity": 200, "unit": "g" },
      { "name": "węgorz", "quantity": 100, "unit": "g" }
    ],
    "steps": [
      "Przygotować ryż i nori.",
      "Ugrillować węgorza.",
      "Zwinąć roladkę i polać sosem unagi."
    ],
    "imageUrl": null
  },
  "isComplete": true
}

---

## 🧩 Warunek techniczny

- **Jeśli przepis został pomyślnie wygenerowany** (ma title, ingredients i steps) —  
  zawsze ustaw **"isComplete": true**.

- **Jeśli przepis nie jest jeszcze gotowy** lub potrzebne jest wyjaśnienie (np. użytkownik powiedział "dodaj zdjęcie" lub "zamień składnik") —  
  ustaw **"isComplete": false**.

- **Nie wstawiaj JSON jako tekstu w ciągu**. Odpowiadaj czystym obiektem JSON, bez \n, \" lub ukośników.

NIE GENERUJ JSON W TEKŚCIE - po prostu prowadź naturalną rozmowę.`,
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
		"суші":   "sushi",
		"sushi":  "sushi",
		"рол":    "sushi",
		"рамен":  "ramen",
		"десерт": "desserts",
		"салат":  "salad",
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
