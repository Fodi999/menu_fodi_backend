package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/dmitrijfomin/menu-fodifood/backend/pkg/utils"
	"gorm.io/gorm"
)

// RecipeMatchService - сервис для подбора рецептов (БЕЗ AI)
type RecipeMatchService struct {
	db *gorm.DB
}

// NewRecipeMatchService - конструктор
func NewRecipeMatchService(db *gorm.DB) *RecipeMatchService {
	return &RecipeMatchService{
		db: db,
	}
}

// RecipeMatch - результат подбора рецепта
type RecipeMatch struct {
	RecipeID           string   `json:"recipeId"`
	CanonicalName      string   `json:"canonicalName"` // 1️⃣ Единый ключ (например: "scrambled_eggs")
	DisplayName        string   `json:"displayName"`   // Локализованное название
	TotalIngredients   int      `json:"totalIngredients"`
	MatchedCount       int      `json:"matchedCount"`
	MatchRatio         float64  `json:"matchRatio"`
	CanCookNow         bool     `json:"canCookNow"`
	Scenario           string   `json:"scenario"`           // 5️⃣ "CAN_COOK_NOW" | "NEED_MORE" | "ALMOST_READY"
	Confidence         string   `json:"confidence"`         // 5️⃣ "EXACT_MATCH" | "HIGH" | "MEDIUM" | "LOW"
	UserIngredients    []string `json:"userIngredients"`    // 2️⃣ Нормализованные названия (toLowerCase)
	MissingIngredients []string `json:"missingIngredients"` // 4️⃣ Недостающие ингредиенты
	MissingCount       int      `json:"missingCount"`
}

// recipeMatchResult - внутренняя структура для SQL запроса
type recipeMatchResult struct {
	RecipeID      string
	CanonicalName string // может быть NULL для user recipes
	Title         string // всегда есть
	LocalName     string // всегда есть
	Total         int
	Matched       int
}

// FindBestRecipe - BACKEND САМ ВЫБИРАЕТ ЛУЧШИЙ РЕЦЕПТ
// Пункт 2: Backend решает, не AI
func (s *RecipeMatchService) FindBestRecipe(ctx context.Context, userID string, lang string) (*RecipeMatch, error) {
	// 2.1 Получить продукты пользователя из холодильника
	var userIngredientIDs []string
	err := s.db.Raw(`
		SELECT DISTINCT ingredient_id 
		FROM user_fridge_items 
		WHERE user_id = ? AND quantity > 0
	`, userID).Pluck("ingredient_id", &userIngredientIDs).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get user ingredients: %w", err)
	}

	if len(userIngredientIDs) == 0 {
		return nil, fmt.Errorf("no ingredients in fridge")
	}

	// 2.2 Посчитать совпадения по каждому рецепту
	var results []recipeMatchResult
	err = s.db.Raw(`
		SELECT 
			r.id AS recipe_id,
			COALESCE(r."canonicalName", '') AS canonical_name,
			r."title" AS title,
			r."localName" AS local_name,
			COUNT(ri.id) AS total,
			COUNT(ri.id) FILTER (
				WHERE ri."ingredientId" IN (?)
			) AS matched
		FROM "Recipe" r
		JOIN "RecipeIngredient" ri ON r.id = ri."recipeId"
		GROUP BY r.id, r."canonicalName", r."title", r."localName"
		HAVING COUNT(ri.id) > 0
		ORDER BY 
			COUNT(ri.id) FILTER (WHERE ri."ingredientId" IN (?)) DESC,
			COUNT(ri.id) ASC
		LIMIT 1
	`, userIngredientIDs, userIngredientIDs).Scan(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to match recipes: %w", err)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no recipes match your ingredients")
	}

	best := results[0]

	// 2.3 Правило (rules engine) - минимум 70% совпадения
	matchRatio := float64(best.Matched) / float64(best.Total)
	canCookNow := matchRatio >= 0.7

	// 5️⃣ Определить сценарий для UI
	var scenario string
	if canCookNow {
		scenario = "CAN_COOK_NOW"
	} else if matchRatio >= 0.5 {
		scenario = "ALMOST_READY" // 50-69% - почти готово
	} else {
		scenario = "NEED_MORE" // < 50% - нужно больше
	}

	// 1️⃣ Получить НОРМАЛИЗОВАННЫЕ названия ингредиентов (через GetName)
	var ingredients []models.Ingredient
	err = s.db.Where("id IN (?)", userIngredientIDs).Find(&ingredients).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get ingredient names: %w", err)
	}

	// 1️⃣ Применить нормализацию через GetName(lang) + ToLower
	ingredientNames := make([]string, 0, len(ingredients))
	for _, ing := range ingredients {
		// 2️⃣ Нормализация: toLowerCase для консистентности
		ingredientNames = append(ingredientNames, normalizeIngredientName(ing.GetName(lang)))
	}

	// 1️⃣ ВСЕГДА генерировать правильный canonical name (английский slug)
	// Canonical name НИКОГДА не берётся из БД напрямую (может быть локализован)
	canonicalName := utils.GenerateCanonicalName(best.LocalName)
	if canonicalName == "" || canonicalName == best.LocalName {
		// Если не нашли в мапе - попробуем title
		canonicalName = utils.GenerateCanonicalName(best.Title)
	}

	// 2️⃣ Локализовать название рецепта (displayName)
	displayName := localizeRecipeName(best, lang)

	// 3️⃣ Получить недостающие ингредиенты (для ALMOST_READY сценария)
	missingIngredients := []string{}
	if !canCookNow && matchRatio >= 0.5 {
		// TODO: получить реальные недостающие ингредиенты из RecipeIngredient
		// Пока оставляем пустым
	}

	// 5️⃣ Confidence как enum
	confidence := calculateConfidence(matchRatio)

	return &RecipeMatch{
		RecipeID:           best.RecipeID,
		CanonicalName:      canonicalName,
		DisplayName:        displayName,
		TotalIngredients:   best.Total,
		MatchedCount:       best.Matched,
		MatchRatio:         matchRatio,
		CanCookNow:         canCookNow,
		Scenario:           scenario,
		Confidence:         confidence,
		UserIngredients:    ingredientNames,
		MissingIngredients: missingIngredients,
		MissingCount:       best.Total - best.Matched,
	}, nil
}

// 2️⃣ normalizeIngredientName - нормализация названия ингредиента
func normalizeIngredientName(name string) string {
	// Приводим к нижнему регистру для консистентности
	// "Соль каменная" -> "соль каменная"
	// "свежие яйца" -> "свежие яйца"
	return strings.ToLower(strings.TrimSpace(name))
}

// 2️⃣ localizeRecipeName - нормализация названия рецепта
func localizeRecipeName(result recipeMatchResult, lang string) string {
	// В текущей БД нет мультиязычных полей для рецептов
	// Используем title или localName
	if result.Title != "" {
		return result.Title
	}
	if result.LocalName != "" {
		return result.LocalName
	}
	// Fallback: canonical name
	return result.CanonicalName
}

// 5️⃣ calculateConfidence - confidence как enum (HIGH, MEDIUM, LOW)
func calculateConfidence(matchRatio float64) string {
	if matchRatio >= 0.9 {
		return "EXACT_MATCH" // 90-100%
	} else if matchRatio >= 0.7 {
		return "HIGH" // 70-89%
	} else if matchRatio >= 0.5 {
		return "MEDIUM" // 50-69%
	}
	return "LOW" // < 50%
}

// AIContext - DTO для AI (пункт 3)
// AI НЕ ВИДИТ БД. AI НЕ ДУМАЕТ. AI ПОЛУЧАЕТ ГОТОВЫЙ ФАКТ.
type AIContext struct {
	Language           string   `json:"language"`           // "Russian" | "Polish" | "English"
	CanonicalName      string   `json:"canonicalName"`      // 2️⃣ Единый ключ рецепта
	DisplayName        string   `json:"displayName"`        // Локализованное название
	Scenario           string   `json:"scenario"`           // 5️⃣ "CAN_COOK_NOW" | "NEED_MORE" | "ALMOST_READY"
	Ingredients        []string `json:"ingredients"`        // 1️⃣ Нормализованные названия (GetName)
	MissingIngredients []string `json:"missingIngredients"` // 3️⃣ Недостающие ингредиенты
	MatchRatio         float64  `json:"matchRatio"`         // 0.0 - 1.0
	CanCookNow         bool     `json:"canCookNow"`         // true если >= 70%
	MissingCount       int      `json:"missingCount"`
}

// PrepareAIContext - подготовить контекст для AI (пункт 3)
func PrepareAIContext(match *RecipeMatch, userLang string) *AIContext {
	return &AIContext{
		Language:           mapLanguageForAI(userLang),
		CanonicalName:      match.CanonicalName,
		DisplayName:        match.DisplayName,
		Scenario:           match.Scenario,
		Ingredients:        match.UserIngredients,
		MissingIngredients: match.MissingIngredients,
		MatchRatio:         match.MatchRatio,
		CanCookNow:         match.CanCookNow,
		MissingCount:       match.MissingCount,
	}
}

// mapLanguageForAI - маппинг языка (пункт 4)
// "ru" → "Russian", "pl" → "Polish", "en" → "English"
func mapLanguageForAI(lang string) string {
	switch lang {
	case "ru":
		return "Russian"
	case "pl":
		return "Polish"
	case "en":
		return "English"
	default:
		return "English"
	}
}

// BuildSystemPrompt - сильный system prompt (пункт 4)
func BuildSystemPrompt(language string) string {
	return fmt.Sprintf(`You are a professional culinary assistant.

CRITICAL RULES:
- Respond strictly in %s language.
- Do NOT mix languages.
- Do NOT invent ingredients.
- Do NOT suggest other recipes.
- Do NOT repeat numeric values or percentages (backend already provides them).
- Explain ONLY the provided recipe in natural language.
- Be concise and clear.
- Return ONLY valid JSON.

Expected JSON format:
{
  "title": "string",
  "reason": "string (natural language, NO numbers)",
  "ingredientsUsed": ["string"]
}`, language)
}

// BuildUserPrompt - user prompt с данными (пункт 5)
// 4️⃣ AI НЕ ДОЛЖЕН повторять цифры - объяснять естественным языком
func BuildUserPrompt(ctx *AIContext) string {
	ingredientsJSON, _ := json.Marshal(ctx.Ingredients)
	missingJSON, _ := json.Marshal(ctx.MissingIngredients)

	return fmt.Sprintf(`Recipe: %s (canonical: %s)
Scenario: %s
Ingredients available: %s
Missing ingredients: %s
Can cook now: %t

DO NOT repeat numeric ratios like "75%%" or "3 out of 4".
Explain in natural %s language why this recipe is recommended (or why they need more ingredients).`,
		ctx.DisplayName,
		ctx.CanonicalName,
		ctx.Scenario,
		string(ingredientsJSON),
		string(missingJSON),
		ctx.CanCookNow,
		ctx.Language,
	)
}

// AIResponse - ожидаемый ответ AI (пункт 6)
// 3️⃣ Confidence ТОЛЬКО в recipe, НЕ в AI response
type AIResponse struct {
	Title           string   `json:"title"`
	Reason          string   `json:"reason"` // Естественный язык БЕЗ цифр
	IngredientsUsed []string `json:"ingredientsUsed"`
}
