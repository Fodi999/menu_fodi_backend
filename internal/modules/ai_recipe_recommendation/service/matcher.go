package service

import (
	"fmt"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"gorm.io/gorm"
)

// ============================================================================
// MATCHER - Rules Engine для подбора рецептов
// Принцип: НЕ использует AI, только факты и математику
// ============================================================================

// RecipeMatcher - сервис для matching рецептов с холодильником
type RecipeMatcher struct {
	db *gorm.DB
}

// NewRecipeMatcher - конструктор
func NewRecipeMatcher(db *gorm.DB) *RecipeMatcher {
	return &RecipeMatcher{db: db}
}

// MatchRecipes - подбирает рецепты на основе холодильника пользователя
// Шаг 1-3: Получает данные, проверяет наличие, возвращает результаты
func (m *RecipeMatcher) MatchRecipes(userID string, lang string, limit int) ([]RecipeMatchResult, error) {
	// Шаг 1: Получаем холодильник пользователя (canonical_ingredient_id)
	fridgeItems, err := m.getUserFridgeCanonicalIDs(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get fridge items: %w", err)
	}

	if len(fridgeItems) == 0 {
		return nil, fmt.Errorf("fridge is empty")
	}

	// Шаг 2: Получаем активные рецепты (status = published, canonical ingredients only)
	recipes, err := m.getActiveRecipes()
	if err != nil {
		return nil, fmt.Errorf("failed to get recipes: %w", err)
	}

	if len(recipes) == 0 {
		return nil, fmt.Errorf("no recipes available")
	}

	// Шаг 3: Rules Engine - проверяем каждый рецепт
	results := make([]RecipeMatchResult, 0, len(recipes))
	for _, recipe := range recipes {
		matchResult := m.matchSingleRecipe(recipe, fridgeItems, lang)
		results = append(results, matchResult)
	}

	// Сортируем по match_percent (DESC)
	sortByMatchPercent(results)

	// Ограничиваем количество результатов
	if limit > 0 && limit < len(results) {
		results = results[:limit]
	}

	return results, nil
}

// getUserFridgeCanonicalIDs - получает ingredient_id из холодильника пользователя
// Использует ingredient_id для matching с RecipeIngredient.ingredientId
func (m *RecipeMatcher) getUserFridgeCanonicalIDs(userID string) (map[string]bool, error) {
	var ingredientIDs []string

	err := m.db.Table("user_fridge_items").
		Select("DISTINCT ingredient_id").
		Where("user_id = ? AND quantity > 0", userID).
		Pluck("ingredient_id", &ingredientIDs).Error

	if err != nil {
		return nil, err
	}

	// Создаем set для быстрого поиска
	fridgeSet := make(map[string]bool)
	for _, id := range ingredientIDs {
		if id != "" {
			fridgeSet[id] = true
		}
	}

	return fridgeSet, nil
}

// getActiveRecipes - получает активные рецепты из каталога
func (m *RecipeMatcher) getActiveRecipes() ([]models.RecipeCatalog, error) {
	var recipes []models.RecipeCatalog

	err := m.db.
		Preload("Ingredients").
		Preload("Ingredients.Ingredient").
		Where("status = ?", "published").
		Find(&recipes).Error

	return recipes, err
}

// matchSingleRecipe - проверяет один рецепт против холодильника
// Шаг 3: Rules Engine (НЕ AI)
func (m *RecipeMatcher) matchSingleRecipe(
	recipe models.RecipeCatalog,
	fridgeSet map[string]bool,
	lang string,
) RecipeMatchResult {
	// 3.1 Обязательные ингредиенты
	requiredIngredients := recipe.Ingredients

	// 3.2 Проверка наличия
	var matched []IngredientInfo
	var missing []IngredientInfo

	for _, recipeIng := range requiredIngredients {
		ingredient := recipeIng.Ingredient
		
		// Используем ingredient.ID для matching (прямое совпадение)
		ingredientID := ingredient.ID

		ingredientInfo := IngredientInfo{
			ID:            ingredient.ID,
			CanonicalName: recipeIng.IngredientKey, // для display
			DisplayName:   ingredient.GetName(lang),
			Quantity:      recipeIng.Quantity,
			Unit:          recipeIng.Unit,
			Category:      ingredient.Category,
		}

		// Проверяем наличие в холодильнике по ingredient_id
		if ingredientID != "" && fridgeSet[ingredientID] {
			matched = append(matched, ingredientInfo)
		} else {
			missing = append(missing, ingredientInfo)
		}
	}

	// Шаг 4: Скоринг (простая формула)
	totalRequired := len(requiredIngredients)
	matchedCount := len(matched)
	missingCount := len(missing)

	matchPercent := 0.0
	if totalRequired > 0 {
		matchPercent = float64(matchedCount) / float64(totalRequired) * 100
	}

	// Шаг 5: Классификация результата
	matchStatus := classifyMatchStatus(missingCount, matchPercent)

	// Формируем результат
	return RecipeMatchResult{
		ID:                   recipe.ID.String(),
		CanonicalName:        recipe.CanonicalName,
		Title:                recipe.GetLocalizedName(lang),
		MatchPercent:         matchPercent,
		MatchStatus:          matchStatus,
		MissingCount:         missingCount,
		AvailableCount:       matchedCount,
		TotalRequired:        totalRequired,
		MissingIngredients:   missing,
		AvailableIngredients: matched,
		CookTime:             recipe.TimeMinutes,
		Portions:             recipe.Servings,
		ImageURL:             recipe.ImageUrl,
	}
}

// classifyMatchStatus - классифицирует статус matching
func classifyMatchStatus(missingCount int, matchPercent float64) string {
	switch {
	case missingCount == 0:
		return StatusReady // 🟢 готово
	case missingCount <= 2 && matchPercent >= 67:
		return StatusAlmostReady // 🟡 почти готово
	default:
		return StatusNotReady // 🔴 не хватает
	}
}

// sortByMatchPercent - сортирует результаты по match_percent (DESC)
func sortByMatchPercent(results []RecipeMatchResult) {
	// Простая bubble sort (для маленьких массивов достаточно)
	n := len(results)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if results[j].MatchPercent < results[j+1].MatchPercent {
				results[j], results[j+1] = results[j+1], results[j]
			}
		}
	}
}
