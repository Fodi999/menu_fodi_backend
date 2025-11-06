package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/ai"
	"github.com/dmitrijfomin/menu-fodifood/backend/pkg/utils"
)

// RecipeGenerationRequest represents the input for recipe generation
type RecipeGenerationRequest struct {
	Title    string `json:"title"`
	Language string `json:"language"` // "pl", "en", "ru", "ua"
}

// GeneratedRecipe represents the AI-generated recipe structure
type GeneratedRecipe struct {
	Title       string             `json:"title"`
	Description string             `json:"description"`
	Category    string             `json:"category"`
	Difficulty  string             `json:"difficulty"` // "beginner", "intermediate", "advanced"
	Time        int                `json:"time"`       // minutes
	Portions    int                `json:"portions"`
	Ingredients []RecipeIngredient `json:"ingredients"`
	Steps       []string           `json:"steps"`

	// Nutrition & Metrics
	GrossWeight  int     `json:"grossWeight"`  // Брутто (г)
	NetWeight    int     `json:"netWeight"`    // Нетто (г)
	Calories     int     `json:"calories"`     // ккал
	Protein      float64 `json:"protein"`      // Белки (г)
	Fats         float64 `json:"fats"`         // Жиры (г)
	Carbs        float64 `json:"carbs"`        // Углеводы (г)
	RecipeYield  int     `json:"yield"`        // Выход (г)
	Cost         float64 `json:"cost"`         // Себестоимость (PLN)
	TokensReward int     `json:"tokensReward"` // ChefTokens награда

	ImageUrl string `json:"imageUrl,omitempty"`
}

// RecipeIngredient represents an ingredient in the recipe
type RecipeIngredient struct {
	Name   string  `json:"name"`
	Amount float64 `json:"amount"`
	Unit   string  `json:"unit"`  // "г", "мл", "шт", "ст.л.", etc.
	Gross  float64 `json:"gross"` // Брутто вес (г)
	Net    float64 `json:"net"`   // Нетто вес (г)
}

// GenerateRecipeHandler generates a complete recipe based on dish title
// POST /api/ai/recipe-helper
func GenerateRecipeHandler(w http.ResponseWriter, r *http.Request) {
	var req RecipeGenerationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input
	if req.Title == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Title is required")
		return
	}

	// Default language to Polish if not specified
	if req.Language == "" {
		req.Language = "pl"
	}

	// Get language-specific prompt
	prompt := buildRecipePrompt(req.Title, req.Language)

	// Call Groq API
	client := ai.NewGroqClient()

	messages := []ai.GroqMessage{
		{
			Role:    "system",
			Content: "You are a professional chef specializing in Japanese and fusion cuisine. Generate recipes in valid JSON format only.",
		},
		{
			Role:    "user",
			Content: prompt,
		},
	}

	response, err := client.Chat(messages, 0.7, 2000)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to generate recipe: "+err.Error())
		return
	}

	// Extract content from response
	if len(response.Choices) == 0 {
		utils.RespondWithError(w, http.StatusInternalServerError, "No response from AI")
		return
	}

	aiContent := response.Choices[0].Message.Content

	// Parse JSON response from AI
	var recipe GeneratedRecipe
	if err := json.Unmarshal([]byte(aiContent), &recipe); err != nil {
		// If AI didn't return valid JSON, try to extract it
		recipe = GeneratedRecipe{
			Title:       req.Title,
			Description: aiContent,
			Category:    "other",
			Difficulty:  "intermediate",
			Time:        30,
			Portions:    4,
			Ingredients: []RecipeIngredient{},
			Steps:       []string{"Szczegóły przepisu zostaną uzupełnione."},
		}
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status": "success",
		"data":   recipe,
	})
}

// buildRecipePrompt creates a language-specific prompt for recipe generation
func buildRecipePrompt(title string, language string) string {
	prompts := map[string]string{
		"pl": `Użytkownik wpisał nazwę potrawy: "` + title + `".

Jesteś profesjonalnym szefem kuchni specjalizującym się w kuchni japońskiej i fusion.
Stwórz kompletny przepis dla tego dania z pełnymi metrykami.

WAŻNE: Zwróć TYLKO czysty JSON bez dodatkowych komentarzy, bez markdown, bez bloków kodu.

Format JSON:
{
  "title": "nazwa dania",
  "description": "krótki opis dania (1-2 zdania)",
  "category": "sushi" | "ramen" | "appetizers" | "desserts" | "fusion" | "other",
  "difficulty": "beginner" | "intermediate" | "advanced",
  "time": liczba minut (np. 45),
  "portions": liczba porcji (np. 4),
  "grossWeight": waga surowych produktów w gramach (np. 600),
  "netWeight": waga po obróbce w gramach (np. 500),
  "calories": kaloryczność całego dania w kcal (np. 850),
  "protein": białko w gramach (np. 35.5),
  "fats": tłuszcze w gramach (np. 28.0),
  "carbs": węglowodany w gramach (np. 95.0),
  "yield": waga gotowego dania w gramach (np. 480),
  "cost": szacunkowy koszt w PLN (np. 45.50),
  "tokensReward": nagroda ChefTokens od 10 do 50 (zależnie od złożoności),
  "ingredients": [
    { 
      "name": "składnik", 
      "amount": liczba, 
      "unit": "г" | "мл" | "шт" | "ст.л.",
      "gross": waga brutto w gramach,
      "net": waga netto w gramach
    }
  ],
  "steps": [
    "Krok 1: szczegółowy opis",
    "Krok 2: szczegółowy opis"
  ]
}

Przykład:
{
  "title": "Філадельфія рол",
  "description": "Класичний рол з лососем, крем-сиром і авокадо",
  "category": "sushi",
  "difficulty": "intermediate",
  "time": 45,
  "portions": 4,
  "grossWeight": 580,
  "netWeight": 520,
  "calories": 920,
  "protein": 42.5,
  "fats": 35.0,
  "carbs": 98.0,
  "yield": 480,
  "cost": 52.00,
  "tokensReward": 25,
  "ingredients": [
    { "name": "Рис для суші", "amount": 300, "unit": "г", "gross": 300, "net": 300 },
    { "name": "Лосось", "amount": 150, "unit": "г", "gross": 180, "net": 150 }
  ],
  "steps": [
    "Відварити рис для суші",
    "Покласти рис на норі"
  ]
}`,

		"en": `User entered dish name: "` + title + `".

You are a professional chef specializing in Japanese and fusion cuisine.
Create a complete recipe for this dish with full nutritional and cost analytics.

IMPORTANT: Return ONLY pure JSON without any additional comments, without markdown, without code blocks.

JSON format:
{
  "title": "dish name",
  "description": "short description (1-2 sentences)",
  "category": "sushi" | "ramen" | "appetizers" | "desserts" | "fusion" | "other",
  "difficulty": "beginner" | "intermediate" | "advanced",
  "time": minutes (e.g. 45),
  "portions": servings (e.g. 4),
  "grossWeight": total gross weight in grams (e.g. 600),
  "netWeight": net weight after processing in grams (e.g. 500),
  "calories": total calories in kcal (e.g. 850),
  "protein": protein in grams (e.g. 35.5),
  "fats": fats in grams (e.g. 28.0),
  "carbs": carbohydrates in grams (e.g. 95.0),
  "yield": final dish weight in grams (e.g. 480),
  "cost": estimated cost in PLN (e.g. 45.50),
  "tokensReward": ChefTokens reward 10-50 based on complexity,
  "ingredients": [
    { "name": "ingredient", "amount": number, "unit": "g" | "ml" | "pcs" | "tbsp", "gross": gross weight in g, "net": net weight in g }
  ],
  "steps": [
    "Step 1: detailed description",
    "Step 2: detailed description"
  ]
}`,

		"ru": `Пользователь ввёл название блюда: "` + title + `".

Ты профессиональный шеф-повар, специализирующийся на японской и фьюжн кухне.
Создай полный рецепт для этого блюда с полной аналитикой питания и стоимости.

ВАЖНО: Верни ТОЛЬКО чистый JSON без дополнительных комментариев, без markdown, без блоков кода.

Формат JSON:
{
  "title": "название блюда",
  "description": "краткое описание (1-2 предложения)",
  "category": "sushi" | "ramen" | "appetizers" | "desserts" | "fusion" | "other",
  "difficulty": "beginner" | "intermediate" | "advanced",
  "time": минуты (напр. 45),
  "portions": порции (напр. 4),
  "grossWeight": общий вес сырых продуктов в граммах (напр. 600),
  "netWeight": вес после обработки в граммах (напр. 500),
  "calories": калорийность всего блюда в ккал (напр. 850),
  "protein": белок в граммах (напр. 35.5),
  "fats": жиры в граммах (напр. 28.0),
  "carbs": углеводы в граммах (напр. 95.0),
  "yield": вес готового блюда в граммах (напр. 480),
  "cost": ориентировочная стоимость в PLN (напр. 45.50),
  "tokensReward": награда ChefTokens от 10 до 50 в зависимости от сложности,
  "ingredients": [
    { "name": "ингредиент", "amount": число, "unit": "г" | "мл" | "шт" | "ст.л.", "gross": вес брутто в г, "net": вес нетто в г }
  ],
  "steps": [
    "Шаг 1: подробное описание",
    "Шаг 2: подробное описание"
  ]
}`,

		"ua": `Користувач ввів назву страви: "` + title + `".

Ти професійний шеф-кухар, що спеціалізується на японській та ф'южн кухні.
Створи повний рецепт для цієї страви з повною аналітикою харчування та вартості.

ВАЖЛИВО: Поверни ТІЛЬКИ чистий JSON без додаткових коментарів, без markdown, без блоків коду.

Формат JSON:
{
  "title": "назва страви",
  "description": "короткий опис (1-2 речення)",
  "category": "sushi" | "ramen" | "appetizers" | "desserts" | "fusion" | "other",
  "difficulty": "beginner" | "intermediate" | "advanced",
  "time": хвилини (напр. 45),
  "portions": порції (напр. 4),
  "grossWeight": загальна вага сирих продуктів у грамах (напр. 600),
  "netWeight": вага після обробки у грамах (напр. 500),
  "calories": калорійність усієї страви в ккал (напр. 850),
  "protein": білок у грамах (напр. 35.5),
  "fats": жири у грамах (напр. 28.0),
  "carbs": вуглеводи у грамах (напр. 95.0),
  "yield": вага готової страви у грамах (напр. 480),
  "cost": орієнтовна вартість у PLN (напр. 45.50),
  "tokensReward": нагорода ChefTokens від 10 до 50 залежно від складності,
  "ingredients": [
    { "name": "інгредієнт", "amount": число, "unit": "г" | "мл" | "шт" | "ст.л.", "gross": вага брутто в г, "net": вага нетто в г }
  ],
  "steps": [
    "Крок 1: детальний опис",
    "Крок 2: детальний опис"
  ]
}`,
	}

	// Get prompt for language, default to Polish
	prompt, ok := prompts[language]
	if !ok {
		prompt = prompts["pl"]
	}

	return prompt
}
