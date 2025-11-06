package ai

import (
	"encoding/json"
	"fmt"
	"log"
)

// RecipeAnalysis результат анализа рецепта
type RecipeAnalysis struct {
	RecipeName     string   `json:"recipeName"`
	Rating         float64  `json:"rating"`         // 1-10
	ChefComment    string   `json:"chefComment"`    // комментарий шефа
	TasteBalance   string   `json:"tasteBalance"`   // sweet/salty/umami/sour/bitter
	Difficulty     string   `json:"difficulty"`     // easy/medium/hard
	EstimatedPrice float64  `json:"estimatedPrice"` // цена блюда
	Category       string   `json:"category"`       // категория
	Improvements   []string `json:"improvements"`   // рекомендации
	Keywords       []string `json:"keywords"`       // ключевые слова
	Allergens      []string `json:"allergens"`      // аллергены
	NutritionalTip string   `json:"nutritionalTip"` // совет по питанию
}

// RecipeAnalyzer анализатор рецептов с AI
type RecipeAnalyzer struct {
	client *GroqClient
	lang   string // pl, ua, en
}

// NewRecipeAnalyzer создаёт новый анализатор
func NewRecipeAnalyzer(lang string) *RecipeAnalyzer {
	if lang == "" {
		lang = "pl" // default польский
	}
	return &RecipeAnalyzer{
		client: NewGroqClient(),
		lang:   lang,
	}
}

// AnalyzeRecipe анализирует рецепт через Groq AI
func (ra *RecipeAnalyzer) AnalyzeRecipe(recipeName, ingredients, instructions string) (*RecipeAnalysis, error) {
	systemPrompt := ra.getSystemPrompt()
	userMessage := ra.buildUserMessage(recipeName, ingredients, instructions)

	log.Printf("[AI] 🧠 Analyzing recipe: %s (lang: %s)", recipeName, ra.lang)

	response, err := ra.client.SimpleChat(systemPrompt, userMessage)
	if err != nil {
		log.Printf("[AI] ❌ Groq API error: %v", err)
		return ra.getFallbackAnalysis(recipeName), nil // возвращаем fallback вместо ошибки
	}

	analysis, err := ra.parseAnalysis(response, recipeName)
	if err != nil {
		log.Printf("[AI] ⚠️ Failed to parse AI response: %v, using fallback", err)
		return ra.getFallbackAnalysis(recipeName), nil
	}

	log.Printf("[AI] ✅ Recipe analyzed: %s - Rating: %.1f/10", recipeName, analysis.Rating)
	return analysis, nil
}

// getSystemPrompt возвращает system prompt для AI в зависимости от языка
func (ra *RecipeAnalyzer) getSystemPrompt() string {
	prompts := map[string]string{
		"pl": `Jesteś doświadczonym szefem kuchni polskiej i japońskiej. 
Analizujesz przepisy kulinarne i dajesz profesjonalną ocenę.
Odpowiadaj TYLKO w formacie JSON bez dodatkowego tekstu.
Format: {"rating": 8.5, "chefComment": "Świetny balans smaku", "tasteBalance": "umami-sweet", "difficulty": "medium", "estimatedPrice": 45.50, "category": "Sushi", "improvements": ["Dodaj więcej imbiru", "Użyj lepszego ryżu"], "keywords": ["łosoś", "awokado", "ryż"], "allergens": ["ryby", "sezam"], "nutritionalTip": "Bogaty w omega-3"}`,

		"ua": `Ти досвідчений шеф-кухар української та японської кухні.
Аналізуй рецепти і давай професійну оцінку.
Відповідай ЛИШЕ у форматі JSON без додаткового тексту.
Формат: {"rating": 8.5, "chefComment": "Чудовий баланс смаку", "tasteBalance": "umami-sweet", "difficulty": "medium", "estimatedPrice": 45.50, "category": "Суші", "improvements": ["Додай більше імбиру", "Використай кращий рис"], "keywords": ["лосось", "авокадо", "рис"], "allergens": ["риба", "кунжут"], "nutritionalTip": "Багатий на омега-3"}`,

		"en": `You are an experienced Polish and Japanese cuisine chef.
Analyze culinary recipes and give professional assessment.
Respond ONLY in JSON format without additional text.
Format: {"rating": 8.5, "chefComment": "Great taste balance", "tasteBalance": "umami-sweet", "difficulty": "medium", "estimatedPrice": 45.50, "category": "Sushi", "improvements": ["Add more ginger", "Use better rice"], "keywords": ["salmon", "avocado", "rice"], "allergens": ["fish", "sesame"], "nutritionalTip": "Rich in omega-3"}`,
	}

	if prompt, ok := prompts[ra.lang]; ok {
		return prompt
	}
	return prompts["pl"] // fallback
}

// buildUserMessage создаёт сообщение пользователя
func (ra *RecipeAnalyzer) buildUserMessage(recipeName, ingredients, instructions string) string {
	templates := map[string]string{
		"pl": fmt.Sprintf(`Przeanalizuj przepis:
Nazwa: %s
Składniki: %s
Instrukcje: %s

Oceń przepis i zwróć JSON.`, recipeName, ingredients, instructions),

		"ua": fmt.Sprintf(`Проаналізуй рецепт:
Назва: %s
Інгредієнти: %s
Інструкції: %s

Оціни рецепт і поверни JSON.`, recipeName, ingredients, instructions),

		"en": fmt.Sprintf(`Analyze the recipe:
Name: %s
Ingredients: %s
Instructions: %s

Rate the recipe and return JSON.`, recipeName, ingredients, instructions),
	}

	if msg, ok := templates[ra.lang]; ok {
		return msg
	}
	return templates["pl"]
}

// parseAnalysis парсит JSON ответ от AI
func (ra *RecipeAnalyzer) parseAnalysis(response, recipeName string) (*RecipeAnalysis, error) {
	var analysis RecipeAnalysis
	if err := json.Unmarshal([]byte(response), &analysis); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	analysis.RecipeName = recipeName
	return &analysis, nil
}

// getFallbackAnalysis возвращает базовую оценку, если AI недоступен
func (ra *RecipeAnalyzer) getFallbackAnalysis(recipeName string) *RecipeAnalysis {
	comments := map[string]string{
		"pl": "Przepis wygląda dobrze! Użyj świeżych składników dla najlepszego smaku.",
		"ua": "Рецепт виглядає добре! Використовуй свіжі інгредієнти для найкращого смаку.",
		"en": "Recipe looks good! Use fresh ingredients for the best taste.",
	}

	comment := comments[ra.lang]
	if comment == "" {
		comment = comments["pl"]
	}

	return &RecipeAnalysis{
		RecipeName:     recipeName,
		Rating:         7.0,
		ChefComment:    comment,
		TasteBalance:   "balanced",
		Difficulty:     "medium",
		EstimatedPrice: 30.0,
		Category:       "Unknown",
		Improvements:   []string{"Use fresh ingredients", "Follow cooking time"},
		Keywords:       []string{"recipe"},
		Allergens:      []string{},
		NutritionalTip: "Balanced nutrition",
	}
}
