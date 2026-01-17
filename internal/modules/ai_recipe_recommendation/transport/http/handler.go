package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/ai_recipe_recommendation/service"
	"gorm.io/gorm"
)

// AIRecipeHandler - handler для AI рекомендаций рецептов (архитектура 2025)
type AIRecipeHandler struct {
	db           *gorm.DB
	matchService *service.RecipeMatchService
}

// NewAIRecipeHandler - конструктор
func NewAIRecipeHandler(db *gorm.DB, matchService *service.RecipeMatchService) *AIRecipeHandler {
	return &AIRecipeHandler{
		db:           db,
		matchService: matchService,
	}
}

// RecommendationResponse - финальный ответ клиенту (пункт 7)
type RecommendationResponse struct {
	Success bool                   `json:"success"`
	Data    *RecommendationData    `json:"data,omitempty"`
	Error   string                 `json:"error,omitempty"`
}

// RecommendationData - полная структура данных
type RecommendationData struct {
	Recipe RecipeData           `json:"recipe"`
	AI     service.AIResponse   `json:"ai"`
}

// RecipeData - информация о рецепте
type RecipeData struct {
	ID                 string   `json:"id"`
	CanonicalName      string   `json:"canonicalName"`      // 1️⃣ Единый ключ (например: "scrambled_eggs")
	DisplayName        string   `json:"displayName"`        // Локализованное название
	CanCookNow         bool     `json:"canCookNow"`
	Scenario           string   `json:"scenario"`           // 5️⃣ "CAN_COOK_NOW" | "NEED_MORE" | "ALMOST_READY"
	Confidence         string   `json:"confidence"`         // 5️⃣ "EXACT_MATCH" | "HIGH" | "MEDIUM" | "LOW"
	MatchRatio         float64  `json:"matchRatio"`
	Ingredients        []string `json:"ingredients"`        // 2️⃣ Нормализованные (toLowerCase)
	MissingIngredients []string `json:"missingIngredients"` // 4️⃣ Недостающие ингредиенты
}

// GetRecommendation - GET /api/ai-recipe/recommendation
// Главный endpoint: backend решает, AI объясняет
func (h *AIRecipeHandler) GetRecommendation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Получить userID из контекста (auth middleware)
	userID, ok := r.Context().Value("userID").(string)
	if !ok || userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(RecommendationResponse{
			Success: false,
			Error:   "Unauthorized",
		})
		return
	}

	// 1️⃣ ЯЗЫК БРАТЬ ТОЛЬКО ИЗ БД (НЕ ИЗ FRONTEND)
	userLang := h.getUserLanguageFromDB(userID)

	// 2️⃣ BACKEND САМ ВЫБИРАЕТ РЕЦЕПТ (НЕ AI)
	match, err := h.matchService.FindBestRecipe(r.Context(), userID, userLang)
	if err != nil {
		w.WriteHeader(http.StatusOK) // Not 500 - это нормальная ситуация
		json.NewEncoder(w).Encode(RecommendationResponse{
			Success: false,
			Error:   "No suitable recipes found. Add more ingredients to your fridge.",
		})
		return
	}

	// 3️⃣ BACKEND ГОТОВИТ DTO ДЛЯ AI
	aiContext := service.PrepareAIContext(match, userLang)

	// 4️⃣ + 5️⃣ SYSTEM PROMPT + USER PROMPT
	systemPrompt := service.BuildSystemPrompt(aiContext.Language)
	userPrompt := service.BuildUserPrompt(aiContext)

	// TODO: Вызов OpenAI API с systemPrompt и userPrompt
	// Сейчас возвращаем mock-ответ для демонстрации архитектуры
	aiResponse := service.AIResponse{
		Title:           h.generateTitle(userLang, match.CanCookNow),
		Reason:          h.generateReason(userLang, match),
		IngredientsUsed: match.UserIngredients,
		// 3️⃣ Confidence убран из AI - он ТОЛЬКО в recipe
	}

	_ = systemPrompt // TODO: использовать в OpenAI API
	_ = userPrompt   // TODO: использовать в OpenAI API

	// 7️⃣ BACKEND ВОЗВРАЩАЕТ ГОТОВЫЙ РЕЗУЛЬТАТ В FRONTEND
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(RecommendationResponse{
		Success: true,
		Data: &RecommendationData{
			Recipe: RecipeData{
				ID:                 match.RecipeID,
				CanonicalName:      match.CanonicalName,
				DisplayName:        match.DisplayName,
				CanCookNow:         match.CanCookNow,
				Scenario:           match.Scenario,
				Confidence:         match.Confidence,
				MatchRatio:         match.MatchRatio,
				Ingredients:        match.UserIngredients,
				MissingIngredients: match.MissingIngredients,
			},
			AI: aiResponse,
		},
	})
}

// getUserLanguageFromDB - ПУНКТ 1: язык ТОЛЬКО из БД
func (h *AIRecipeHandler) getUserLanguageFromDB(userID string) string {
	var user models.User
	err := h.db.Select("settings").Where("id = ?", userID).First(&user).Error
	if err != nil {
		return "pl" // default fallback
	}

	lang := string(user.Settings.Language)
	if lang == "" {
		return "pl"
	}

	return lang // "ru" | "pl" | "en"
}

// generateTitle - генерация заголовка (TODO: заменить на AI)
func (h *AIRecipeHandler) generateTitle(lang string, canCook bool) string {
	if canCook {
		switch lang {
		case "ru":
			return "Можно готовить сейчас"
		case "pl":
			return "Możesz gotować teraz"
		default:
			return "You can cook now"
		}
	}

	switch lang {
	case "ru":
		return "Нужно больше ингредиентов"
	case "pl":
		return "Potrzebujesz więcej składników"
	default:
		return "Need more ingredients"
	}
}

// generateReason - генерация объяснения (TODO: заменить на AI)
// 2️⃣ Mock БЕЗ повторения цифр (естественный язык)
func (h *AIRecipeHandler) generateReason(lang string, match *service.RecipeMatch) string {
	if match.CanCookNow {
		switch lang {
		case "ru":
			return "У вас есть все необходимые ингредиенты для приготовления этого блюда."
		case "pl":
			return "Masz wszystkie niezbędne składniki do przygotowania tego dania."
		default:
			return "You have all the necessary ingredients to prepare this dish."
		}
	}

	switch lang {
	case "ru":
		return fmt.Sprintf("Вам не хватает нескольких ингредиентов для приготовления %s.", match.DisplayName)
	case "pl":
		return fmt.Sprintf("Brakuje Ci kilku składników do przygotowania %s.", match.DisplayName)
	default:
		return fmt.Sprintf("You're missing a few ingredients for %s.", match.DisplayName)
	}
}
