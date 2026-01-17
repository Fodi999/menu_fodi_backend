package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
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
	RecipeID         string   `json:"recipeId"`
	RecipeName       string   `json:"recipeName"`
	TotalIngredients int      `json:"totalIngredients"`
	MatchedCount     int      `json:"matchedCount"`
	MatchRatio       float64  `json:"matchRatio"`
	CanCookNow       bool     `json:"canCookNow"`
	UserIngredients  []string `json:"userIngredients"`    // Названия ингредиентов на нужном языке
	MissingCount     int      `json:"missingCount"`
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
	type RecipeMatchResult struct {
		RecipeID   string
		RecipeName string
		Total      int
		Matched    int
	}

	var results []RecipeMatchResult
	err = s.db.Raw(`
		SELECT 
			r.id AS recipe_id,
			r."canonicalName" AS recipe_name,
			COUNT(ri.id) AS total,
			COUNT(ri.id) FILTER (
				WHERE ri."ingredientId" IN (?)
			) AS matched
		FROM "Recipe" r
		JOIN "RecipeIngredient" ri ON r.id = ri."recipeId"
		GROUP BY r.id, r."canonicalName"
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

	// Получить названия ингредиентов на нужном языке
	var ingredients []models.Ingredient
	err = s.db.Where("id IN (?)", userIngredientIDs).Find(&ingredients).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get ingredient names: %w", err)
	}

	ingredientNames := make([]string, 0, len(ingredients))
	for _, ing := range ingredients {
		ingredientNames = append(ingredientNames, ing.GetName(lang))
	}

	return &RecipeMatch{
		RecipeID:         best.RecipeID,
		RecipeName:       best.RecipeName,
		TotalIngredients: best.Total,
		MatchedCount:     best.Matched,
		MatchRatio:       matchRatio,
		CanCookNow:       canCookNow,
		UserIngredients:  ingredientNames,
		MissingCount:     best.Total - best.Matched,
	}, nil
}

// AIContext - DTO для AI (пункт 3)
// AI НЕ ВИДИТ БД. AI НЕ ДУМАЕТ. AI ПОЛУЧАЕТ ГОТОВЫЙ ФАКТ.
type AIContext struct {
	Language     string   `json:"language"`      // "Russian" | "Polish" | "English"
	RecipeName   string   `json:"recipeName"`
	Ingredients  []string `json:"ingredients"`   // локализованные названия
	MatchRatio   float64  `json:"matchRatio"`    // 0.0 - 1.0
	CanCookNow   bool     `json:"canCookNow"`    // true если >= 70%
	MissingCount int      `json:"missingCount"`
}

// PrepareAIContext - подготовить контекст для AI (пункт 3)
func PrepareAIContext(match *RecipeMatch, userLang string) *AIContext {
	return &AIContext{
		Language:     mapLanguageForAI(userLang),
		RecipeName:   match.RecipeName,
		Ingredients:  match.UserIngredients,
		MatchRatio:   match.MatchRatio,
		CanCookNow:   match.CanCookNow,
		MissingCount: match.MissingCount,
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
- Explain ONLY the provided recipe.
- Be concise and clear.
- Return ONLY valid JSON.

Expected JSON format:
{
  "title": "string",
  "reason": "string",
  "ingredientsUsed": ["string"],
  "confidence": 0.0-1.0
}`, language)
}

// BuildUserPrompt - user prompt с данными (пункт 5)
func BuildUserPrompt(ctx *AIContext) string {
	ingredientsJSON, _ := json.Marshal(ctx.Ingredients)
	
	return fmt.Sprintf(`Recipe: %s
Ingredients available: %s
Match ratio: %.2f
Can cook now: %t
Missing ingredients: %d

Explain to the user why this recipe is recommended (or why they need more ingredients).`,
		ctx.RecipeName,
		string(ingredientsJSON),
		ctx.MatchRatio,
		ctx.CanCookNow,
		ctx.MissingCount,
	)
}

// AIResponse - ожидаемый ответ AI (пункт 6)
type AIResponse struct {
	Title           string   `json:"title"`
	Reason          string   `json:"reason"`
	IngredientsUsed []string `json:"ingredientsUsed"`
	Confidence      float64  `json:"confidence"`
}
