package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/ai"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/database"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/dmitrijfomin/menu-fodifood/backend/pkg/utils"
)

// AnalyzeRecipeRequest запрос на анализ рецепта
type AnalyzeRecipeRequest struct {
	RecipeName   string `json:"recipeName" binding:"required"`
	Ingredients  string `json:"ingredients" binding:"required"`
	Instructions string `json:"instructions" binding:"required"`
	Language     string `json:"language"` // pl, ua, en
}

// MentorChatRequest запрос к AI наставнику
type MentorChatRequest struct {
	Question string `json:"question" binding:"required"`
	Language string `json:"language"` // pl, ua, en
}

// EstimatePriceRequest запрос на оценку цены
type EstimatePriceRequest struct {
	RecipeName  string `json:"recipeName" binding:"required"`
	Ingredients string `json:"ingredients" binding:"required"`
	PortionSize int    `json:"portionSize"`
	Language    string `json:"language"`
}

// AnalyzeStepRequest запрос на анализ шага рецепта
type AnalyzeStepRequest struct {
	Step     string `json:"step" binding:"required"`
	Language string `json:"language"`
}

// AnalyzeRecipeHandler POST /api/ai/analyze - анализ рецепта
func AnalyzeRecipeHandler(w http.ResponseWriter, r *http.Request) {
	var req AnalyzeRecipeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"message": "Invalid request body",
		})
		return
	}

	if req.Language == "" {
		req.Language = "pl"
	}

	analyzer := ai.NewRecipeAnalyzer(req.Language)
	analysis, err := analyzer.AnalyzeRecipe(req.RecipeName, req.Ingredients, req.Instructions)
	if err != nil {
		log.Printf("[AI] Error analyzing recipe: %v", err)
		utils.RespondWithJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"status":  "error",
			"message": "Failed to analyze recipe",
		})
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status": "ok",
		"data":   analysis,
	})
}

// MentorChatHandler POST /api/mentor/chat - чат с AI наставником
func MentorChatHandler(w http.ResponseWriter, r *http.Request) {
	var req MentorChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"message": "Invalid request body",
		})
		return
	}

	if req.Language == "" {
		req.Language = "pl"
	}

	mentor := ai.NewMentorChat(req.Language)
	answer, err := mentor.Ask(req.Question)
	if err != nil {
		log.Printf("[MENTOR] Error: %v", err)
		utils.RespondWithJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"status":  "error",
			"message": "Failed to get mentor response",
		})
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status": "ok",
		"data": map[string]interface{}{
			"question": req.Question,
			"answer":   answer,
			"language": req.Language,
		},
	})
}

// EstimatePriceHandler POST /api/ai/estimate-price - оценка цены блюда
func EstimatePriceHandler(w http.ResponseWriter, r *http.Request) {
	var req EstimatePriceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"message": "Invalid request body",
		})
		return
	}

	if req.Language == "" {
		req.Language = "pl"
	}
	if req.PortionSize == 0 {
		req.PortionSize = 1
	}

	estimator := ai.NewPriceEstimator(req.Language)
	estimation, err := estimator.EstimatePrice(req.RecipeName, req.Ingredients, req.PortionSize)
	if err != nil {
		log.Printf("[PRICE] Error estimating price: %v", err)
		utils.RespondWithJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"status":  "error",
			"message": "Failed to estimate price",
		})
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status": "ok",
		"data":   estimation,
	})
}

// AnalyzeStepHandler POST /api/mentor/analyze-step - анализ шага рецепта
func AnalyzeStepHandler(w http.ResponseWriter, r *http.Request) {
	var req AnalyzeStepRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"message": "Invalid request body",
		})
		return
	}

	if req.Language == "" {
		req.Language = "pl"
	}

	mentor := ai.NewMentorChat(req.Language)

	// Формируем специфичный запрос для анализа шага
	question := ""
	switch req.Language {
	case "pl":
		question = "Czy ten krok prawidłowy? Jakie wskazówki masz dla tego kroku: " + req.Step
	case "ua":
		question = "Чи правильний цей крок? Які поради ви маєте для цього кроку: " + req.Step
	case "en":
		question = "Is this step correct? What tips do you have for this step: " + req.Step
	default:
		question = "Czy ten krok prawidłowy? Jakie wskazówki masz dla tego kroku: " + req.Step
	}

	answer, err := mentor.Ask(question)
	if err != nil {
		log.Printf("[MENTOR] Error analyzing step: %v", err)
		utils.RespondWithJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"status":  "error",
			"message": "Failed to analyze step",
		})
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status": "ok",
		"data": map[string]interface{}{
			"comment": answer,
		},
	})
}

// ReviewRecipeHandler POST /api/ai/review-recipe - AI оценивает рецепт ученика
func ReviewRecipeHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RecipeID string `json:"recipeId"`
		Language string `json:"language"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"message": "Invalid request body",
		})
		return
	}

	if req.Language == "" {
		req.Language = "pl"
	}

	// Получаем рецепт из базы
	var recipe models.PersonalRecipe
	if err := database.DB.Where("id = ?", req.RecipeID).First(&recipe).Error; err != nil {
		utils.RespondWithJSON(w, http.StatusNotFound, map[string]interface{}{
			"status":  "error",
			"message": "Recipe not found",
		})
		return
	}

	// Формируем данные для AI анализа
	ingredients := strings.Join(recipe.Ingredients, ", ")
	steps := strings.Join(recipe.Steps, ". ")

	analyzer := ai.NewRecipeAnalyzer(req.Language)
	analysis, err := analyzer.AnalyzeRecipe(recipe.Title, ingredients, steps)
	if err != nil {
		log.Printf("[AI REVIEW] Error: %v", err)
		utils.RespondWithJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"status":  "error",
			"message": "Failed to analyze recipe",
		})
		return
	}

	// Обновляем рейтинг рецепта в базе
	recipe.Rating = analysis.Rating
	database.DB.Save(&recipe)

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status": "ok",
		"data": map[string]interface{}{
			"recipeId":       recipe.ID,
			"rating":         analysis.Rating,
			"chefComment":    analysis.ChefComment,
			"tasteBalance":   analysis.TasteBalance,
			"difficulty":     analysis.Difficulty,
			"improvements":   analysis.Improvements,
			"estimatedPrice": analysis.EstimatedPrice,
		},
	})
}

// CritiqueRecipeHandler POST /api/ai/critique - глубокий AI-анализ рецепта
func CritiqueRecipeHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RecipeID string `json:"recipeId"`
		Language string `json:"language"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"message": "Invalid request body",
		})
		return
	}

	if req.Language == "" {
		req.Language = "pl"
	}

	// Получаем рецепт из базы
	var recipe models.PersonalRecipe
	if err := database.DB.Where("id = ?", req.RecipeID).First(&recipe).Error; err != nil {
		utils.RespondWithJSON(w, http.StatusNotFound, map[string]interface{}{
			"status":  "error",
			"message": "Recipe not found",
		})
		return
	}

	// Формируем запрос к AI для детального анализа
	ingredients := strings.Join(recipe.Ingredients, ", ")
	steps := strings.Join(recipe.Steps, ". ")

	systemPrompt := getCritiqueSystemPrompt(req.Language)
	userMessage := fmt.Sprintf(`Przeprowadź głęboką krytykę przepisu:
Nazwa: %s
Składniki: %s
Instrukcje: %s

Oceń w skali 0-10 następujące aspekty:
- Smak (taste)
- Prezentacja (presentation)
- Technika (technique)
- Kreatywność (creativity)
- Zdrowie (health)

Zwróć JSON format:
{
  "overallRating": 8.5,
  "taste": 9.0,
  "presentation": 7.5,
  "technique": 8.0,
  "creativity": 9.5,
  "health": 7.0,
  "masterComment": "Szczegółowy komentarz mistrza kuchni...",
  "strengths": ["mocna strona 1", "mocna strona 2"],
  "weaknesses": ["słabość 1", "słabość 2"],
  "suggestions": ["sugestia poprawy 1", "sugestia 2"]
}`, recipe.Title, ingredients, steps)

	log.Printf("[AI CRITIQUE] 🎨 Detailed critique for: %s (lang: %s)", recipe.Title, req.Language)

	client := ai.NewGroqClient()
	response, err := client.SimpleChat(systemPrompt, userMessage)
	if err != nil {
		log.Printf("[AI CRITIQUE] Error: %v", err)
		utils.RespondWithJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"status":  "error",
			"message": "Failed to critique recipe",
		})
		return
	}

	// Парсим ответ
	var critique map[string]interface{}
	if err := json.Unmarshal([]byte(response), &critique); err != nil {
		log.Printf("[AI CRITIQUE] Failed to parse: %v", err)
		// Fallback критика
		critique = getFallbackCritique(req.Language, recipe.Title)
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status": "ok",
		"data":   critique,
	})
}

// getCritiqueSystemPrompt возвращает system prompt для детального анализа
func getCritiqueSystemPrompt(lang string) string {
	prompts := map[string]string{
		"pl": `Jesteś mistrzem kuchni polskiej i japońskiej z 30-letnim doświadczeniem.
Przeprowadzasz profesjonalną krytykę kulinarną z głęboką analizą smaku, techniki i prezentacji.
Odpowiadaj TYLKO w formacie JSON bez dodatkowego tekstu.`,

		"ua": `Ти майстер польської та японської кухні з 30-річним досвідом.
Проводь професійну кулінарну критику з глибоким аналізом смаку, техніки та презентації.
Відповідай ЛИШЕ у форматі JSON без додаткового тексту.`,

		"en": `You are a master chef of Polish and Japanese cuisine with 30 years of experience.
Conduct professional culinary critique with deep analysis of taste, technique, and presentation.
Respond ONLY in JSON format without additional text.`,
	}

	if prompt, ok := prompts[lang]; ok {
		return prompt
	}
	return prompts["pl"]
}

// getFallbackCritique возвращает базовую критику при недоступности AI
func getFallbackCritique(lang, recipeName string) map[string]interface{} {
	return map[string]interface{}{
		"overallRating": 7.5,
		"taste":         7.5,
		"presentation":  7.0,
		"technique":     7.5,
		"creativity":    8.0,
		"health":        7.0,
		"masterComment": "Recipe looks promising. Use fresh ingredients for best results.",
		"strengths":     []string{"Good ingredient selection", "Clear instructions"},
		"weaknesses":    []string{"Could improve presentation", "Add more seasoning details"},
		"suggestions":   []string{"Use fresh herbs", "Focus on plating"},
	}
}
