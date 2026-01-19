package service

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"gorm.io/gorm"
)

// RecipeMatchService handles recipe matching logic
type RecipeMatchService struct {
	db *gorm.DB
}

func NewRecipeMatchService(db *gorm.DB) *RecipeMatchService {
	return &RecipeMatchService{db: db}
}

// RecipeMatch represents a recipe with match score
type RecipeMatch struct {
	Recipe             models.RecipeCatalog `json:"recipe"`
	MatchScore         float64              `json:"matchScore"` // 0-100
	MatchedIngredients []MatchedIngredient  `json:"matchedIngredients"`
	MissingIngredients []MissingIngredient  `json:"missingIngredients"`
	CostToComplete     float64              `json:"costToComplete"` // PLN to buy missing ingredients
	HasExpiringItems   bool                 `json:"hasExpiringItems"`
	ExpiringItemsCount int                  `json:"expiringItemsCount"`
	CanMakeNow         bool                 `json:"canMakeNow"` // All required ingredients available

	// Economy calculations (clear semantics for frontend)
	UsedValue       float64 `json:"usedValue"`       // PLN: cost of ingredients used from fridge
	SavedMoney      float64 `json:"savedMoney"`      // PLN: money saved by having ingredients (= usedValue, "Wartość z lodówki")
	TotalRecipeCost float64 `json:"totalRecipeCost"` // PLN: full recipe cost (usedValue + costToComplete)
	WasteRiskSaved  float64 `json:"wasteRiskSaved"`  // PLN: value of expiring items used (prevents food waste)
}

type MatchedIngredient struct {
	IngredientID   string             `json:"ingredientId"`
	Name           string             `json:"name"`
	Required       float64            `json:"required"`
	Available      float64            `json:"available"`
	Unit           string             `json:"unit"`
	IsExpiringSoon bool               `json:"isExpiringSoon"`
	ExpiresAt      *time.Time         `json:"expiresAt,omitempty"`
	Optional       bool               `json:"optional"` // Is this ingredient optional for the recipe?
	Ingredient     *models.Ingredient `json:"-"`        // Full ingredient for localization (not exported to JSON)
}

type MissingIngredient struct {
	IngredientID  string             `json:"ingredientId"`
	Name          string             `json:"name"`
	Required      float64            `json:"required"`
	Unit          string             `json:"unit"`
	EstimatedCost float64            `json:"estimatedCost"` // PLN
	Optional      bool               `json:"optional"`
	Ingredient    *models.Ingredient `json:"-"` // Full ingredient for localization (not exported to JSON)
}

// MatchRecipesWithFridge finds recipes that match user's fridge contents
func (s *RecipeMatchService) MatchRecipesWithFridge(
	userID string,
	filters RecipeFilters,
) ([]RecipeMatch, error) {
	// 1. Load user's fridge with prices
	fridgeItems, err := s.loadFridgeWithPrices(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to load fridge: %w", err)
	}

	if len(fridgeItems) == 0 {
		return []RecipeMatch{}, nil
	}

	// 2. Build ingredient map for fast lookup by ingredientId
	fridgeMap := make(map[string]*FridgeItem)
	for i := range fridgeItems {
		// Use ingredientId as key for precise matching
		fridgeMap[fridgeItems[i].ID] = &fridgeItems[i]

		// Also add normalized name as fallback for compatibility
		key := normalizeIngredientName(fridgeItems[i].Name)
		if _, exists := fridgeMap[key]; !exists {
			fridgeMap[key] = &fridgeItems[i]
		}
	}

	// 3. Load recipes from catalog with filters
	recipes, err := s.loadRecipesWithFilters(filters)
	if err != nil {
		return nil, fmt.Errorf("failed to load recipes: %w", err)
	}

	// 4. Calculate match score for each recipe
	matches := make([]RecipeMatch, 0, len(recipes))
	for _, recipe := range recipes {
		match := s.calculateRecipeMatch(recipe, fridgeMap)

		// Apply minimum score threshold
		if match.MatchScore < filters.MinScore {
			continue
		}

		// Apply cookable filter if requested
		if filters.OnlyCookable && !match.CanMakeNow {
			continue
		}

		matches = append(matches, match)
	}

	// 5. Sort by: canCookNow DESC -> score DESC -> costToComplete ASC -> timeMinutes ASC
	sortRecipeMatches(matches)

	// 6. Return top N results
	if filters.Limit > 0 && len(matches) > filters.Limit {
		matches = matches[:filters.Limit]
	}

	return matches, nil
}

// calculateRecipeMatch calculates match score and details
func (s *RecipeMatchService) calculateRecipeMatch(
	recipe models.RecipeCatalog,
	fridgeMap map[string]*FridgeItem,
) RecipeMatch {
	match := RecipeMatch{
		Recipe:             recipe,
		MatchedIngredients: []MatchedIngredient{},
		MissingIngredients: []MissingIngredient{},
		CostToComplete:     0,
		HasExpiringItems:   false,
		ExpiringItemsCount: 0,
		UsedValue:          0,
		SavedMoney:         0,
		TotalRecipeCost:    0,
		WasteRiskSaved:     0,
	}

	requiredCount := 0
	matchedCount := 0
	optionalMatchedCount := 0

	// DEBUG: Log recipe matching attempt
	fmt.Printf("🔍 Matching recipe: %s (canonicalName=%s, ingredients=%d)\n",
		recipe.ID.String(), recipe.CanonicalName, len(recipe.Ingredients))

	// CRITICAL: Skip recipes without ingredients (invalid data)
	if len(recipe.Ingredients) == 0 {
		fmt.Printf("  ⚠️  SKIPPED: Recipe has no ingredients (invalid data)\n\n")
		match.CanMakeNow = false
		match.MatchScore = 0
		return match
	}

	for _, recipeIng := range recipe.Ingredients {
		fmt.Printf("  → Ingredient: id=%s, quantity=%.2f %s, optional=%v\n",
			recipeIng.IngredientID, recipeIng.Quantity, recipeIng.Unit, recipeIng.Optional)

		// CRITICAL: Skip ingredients with invalid quantity (data quality issue)
		if recipeIng.Quantity <= 0 {
			fmt.Printf("    ⚠️  INVALID: quantity <= 0 (skipping this ingredient)\n")
			continue
		}

		if recipeIng.Optional {
			// Optional ingredients don't affect core match score
			fridgeItem := s.findIngredientInFridge(recipeIng, fridgeMap)
			if fridgeItem != nil {
				optionalMatchedCount++
				match.MatchedIngredients = append(match.MatchedIngredients, MatchedIngredient{
					IngredientID:   recipeIng.IngredientID,
					Name:           recipeIng.Ingredient.Name,
					Required:       recipeIng.Quantity,
					Available:      fridgeItem.Quantity,
					Unit:           recipeIng.Unit,
					IsExpiringSoon: fridgeItem.IsExpiringSoon,
					ExpiresAt:      fridgeItem.ExpiresAt,
					Optional:       true,
					Ingredient:     &recipeIng.Ingredient, // Store full ingredient for localization
				})
			}
			continue
		}

		requiredCount++
		fridgeItem := s.findIngredientInFridge(recipeIng, fridgeMap)

		if fridgeItem != nil && fridgeItem.Quantity >= recipeIng.Quantity {
			// Ingredient available in sufficient quantity
			matchedCount++
			fmt.Printf("    ✅ MATCHED: found in fridge (available=%.2f %s, required=%.2f %s)\n",
				fridgeItem.Quantity, fridgeItem.Unit, recipeIng.Quantity, recipeIng.Unit)

			// Calculate value of used ingredient
			ingredientValue := recipeIng.Quantity * fridgeItem.PricePerUnit
			match.UsedValue += ingredientValue

			// Track expiring items value (waste prevention)
			if fridgeItem.IsExpiringSoon {
				match.HasExpiringItems = true
				match.ExpiringItemsCount++
				match.WasteRiskSaved += ingredientValue
			}

			matched := MatchedIngredient{
				IngredientID:   recipeIng.IngredientID,
				Name:           recipeIng.Ingredient.Name,
				Required:       recipeIng.Quantity,
				Available:      fridgeItem.Quantity,
				Unit:           recipeIng.Unit,
				IsExpiringSoon: fridgeItem.IsExpiringSoon,
				ExpiresAt:      fridgeItem.ExpiresAt,
				Optional:       false,
				Ingredient:     &recipeIng.Ingredient, // Store full ingredient for localization
			}
			match.MatchedIngredients = append(match.MatchedIngredients, matched)
		} else {
			// Ingredient missing or insufficient
			if fridgeItem != nil {
				fmt.Printf("    ❌ INSUFFICIENT: found but not enough (available=%.2f, required=%.2f)\n",
					fridgeItem.Quantity, recipeIng.Quantity)
			} else {
				fmt.Printf("    ❌ NOT FOUND: ingredient not in fridge\n")
			}

			pricePerUnit := 0.0
			if recipeIng.Ingredient.DefaultPricePerUnit != nil {
				pricePerUnit = *recipeIng.Ingredient.DefaultPricePerUnit
			}
			estimatedCost := roundToTwoDecimals(recipeIng.Quantity * pricePerUnit)
			missing := MissingIngredient{
				IngredientID:  recipeIng.IngredientID,
				Name:          recipeIng.Ingredient.Name,
				Required:      recipeIng.Quantity,
				Unit:          recipeIng.Unit,
				EstimatedCost: estimatedCost,
				Optional:      recipeIng.Optional,
				Ingredient:    &recipeIng.Ingredient, // Store full ingredient for localization
			}
			match.MissingIngredients = append(match.MissingIngredients, missing)
			match.CostToComplete += estimatedCost
		}
	}

	// Calculate base match score
	if requiredCount > 0 {
		match.MatchScore = (float64(matchedCount) / float64(requiredCount)) * 100
	}

	// Bonus for optional ingredients
	if len(recipe.Ingredients) > requiredCount && optionalMatchedCount > 0 {
		optionalBonus := (float64(optionalMatchedCount) / float64(len(recipe.Ingredients)-requiredCount)) * 5
		match.MatchScore += optionalBonus
	}

	// Bonus for using expiring items (prioritize waste reduction)
	if match.HasExpiringItems {
		expiryBonus := float64(match.ExpiringItemsCount) * 2.0
		match.MatchScore += expiryBonus
	}

	// Cap at 100
	if match.MatchScore > 100 {
		match.MatchScore = 100
	}

	// Round score to 2 decimals
	match.MatchScore = roundToTwoDecimals(match.MatchScore)

	// Round economy values to 2 decimals
	match.CostToComplete = roundToTwoDecimals(match.CostToComplete)
	match.UsedValue = roundToTwoDecimals(match.UsedValue)
	match.WasteRiskSaved = roundToTwoDecimals(match.WasteRiskSaved)

	// Calculate economy semantics (clear for frontend)
	match.SavedMoney = match.UsedValue                                                 // Money saved by having ingredients (already rounded)
	match.TotalRecipeCost = roundToTwoDecimals(match.UsedValue + match.CostToComplete) // Full recipe cost

	// Determine if can make now
	match.CanMakeNow = (matchedCount == requiredCount)

	// DEBUG: Log final result
	fmt.Printf("  📊 RESULT: score=%.2f, matched=%d/%d, canMakeNow=%v\n\n",
		match.MatchScore, matchedCount, requiredCount, match.CanMakeNow)

	return match
}

// findIngredientInFridge finds ingredient in fridge by ingredientId (primary) or normalized name (fallback)
func (s *RecipeMatchService) findIngredientInFridge(
	recipeIng models.CatalogIngredient,
	fridgeMap map[string]*FridgeItem,
) *FridgeItem {
	// 1. Try exact match by ingredientId (MOST RELIABLE)
	if recipeIng.IngredientID != "" {
		if item, ok := fridgeMap[recipeIng.IngredientID]; ok {
			return item
		}
	}

	// 2. Try by ingredient key (legacy compatibility)
	if item, ok := fridgeMap[recipeIng.IngredientKey]; ok {
		return item
	}

	// 3. Try normalized name match (fallback for backwards compatibility)
	key := normalizeIngredientName(recipeIng.Ingredient.Name)
	if item, ok := fridgeMap[key]; ok {
		return item
	}

	// 4. Try fuzzy match (contains) - last resort
	for fridgeKey, item := range fridgeMap {
		if strings.Contains(fridgeKey, key) || strings.Contains(key, fridgeKey) {
			return item
		}
	}

	return nil
}

// FridgeItem represents user's fridge item with calculated data
type FridgeItem struct {
	ID             string
	Name           string
	Quantity       float64
	Unit           string
	PricePerUnit   float64
	ExpiresAt      *time.Time
	IsExpiringSoon bool // Within 3 days
}

// loadFridgeWithPrices loads user's fridge items with prices
func (s *RecipeMatchService) loadFridgeWithPrices(userID string) ([]FridgeItem, error) {
	var dbItems []models.UserFridgeItem
	err := s.db.
		Preload("Ingredient").
		Where("user_id = ?", userID).
		Find(&dbItems).Error

	if err != nil {
		return nil, err
	}

	items := make([]FridgeItem, 0, len(dbItems))

	for _, item := range dbItems {
		if item.Ingredient == nil {
			continue
		}

		pricePerUnit := 0.0
		if item.CurrentPricePerUnit != nil {
			pricePerUnit = *item.CurrentPricePerUnit
		} else if item.Ingredient.DefaultPricePerUnit != nil {
			pricePerUnit = *item.Ingredient.DefaultPricePerUnit
		}

		isExpiringSoon := false
		if item.ExpiresAt != nil {
			daysUntilExpiry := time.Until(*item.ExpiresAt).Hours() / 24
			isExpiringSoon = daysUntilExpiry <= 3 && daysUntilExpiry > 0
		}

		items = append(items, FridgeItem{
			ID:             item.Ingredient.ID, // Use ingredientId, not fridgeItemId
			Name:           item.Ingredient.Name,
			Quantity:       item.Quantity,
			Unit:           item.Ingredient.Unit,
			PricePerUnit:   pricePerUnit,
			ExpiresAt:      item.ExpiresAt,
			IsExpiringSoon: isExpiringSoon,
		})
	}

	return items, nil
}

// RecipeFilters for filtering recipe catalog
type RecipeFilters struct {
	Country          string   // "Poland", "Italy", etc.
	Category         string   // "main", "dessert", etc.
	Difficulty       string   // "easy", "medium", "hard"
	MaxTime          int      // Maximum time in minutes
	ExcludeAllergens []string // ["gluten", "lactose"]
	IncludeDietTags  []string // ["vegetarian", "keto"]
	MinScore         float64  // Minimum match score (0-100), default: 0
	OnlyCookable     bool     // Only show recipes that can be cooked now (all required ingredients available)
	Limit            int      // Max results
	ExcludeRecipeIds []string // Recipe UUIDs to exclude from results (for "show next" functionality)
}

// loadRecipesWithFilters loads recipes from catalog with filters
func (s *RecipeMatchService) loadRecipesWithFilters(filters RecipeFilters) ([]models.RecipeCatalog, error) {
	query := s.db.Model(&models.RecipeCatalog{}).
		Preload("Ingredients.Ingredient").
		Preload("Allergens").
		Preload("DietTags")

	// Apply filters
	if filters.Country != "" {
		query = query.Where("country = ?", filters.Country)
	}
	if filters.Category != "" {
		query = query.Where("category = ?", filters.Category)
	}
	if filters.Difficulty != "" {
		query = query.Where("difficulty = ?", filters.Difficulty)
	}
	if filters.MaxTime > 0 {
		query = query.Where("\"timeMinutes\" <= ?", filters.MaxTime)
	}

	// Exclude allergens
	if len(filters.ExcludeAllergens) > 0 {
		query = query.Where(`id NOT IN (
			SELECT "recipeId" FROM "RecipeAllergen" ra
			JOIN "Allergen" a ON a.id = ra."allergenId"
			WHERE a.name IN ?
		)`, filters.ExcludeAllergens)
	}

	// Include diet tags
	if len(filters.IncludeDietTags) > 0 {
		query = query.Where(`id IN (
			SELECT "recipeId" FROM "RecipeDietTag" rdt
			JOIN "DietTag" dt ON dt.id = rdt."dietTagId"
			WHERE dt.name IN ?
		)`, filters.IncludeDietTags)
	}

	// Exclude specific recipe IDs (for "show next" functionality)
	if len(filters.ExcludeRecipeIds) > 0 {
		query = query.Where("id NOT IN ?", filters.ExcludeRecipeIds)
	}

	var recipes []models.RecipeCatalog
	err := query.Find(&recipes).Error
	return recipes, err
}

// normalizeIngredientName normalizes ingredient name for matching
func normalizeIngredientName(name string) string {
	normalized := strings.ToLower(name)
	normalized = strings.TrimSpace(normalized)
	// Remove plural forms (basic)
	normalized = strings.TrimSuffix(normalized, "y") // ziemniaky -> ziemniak
	normalized = strings.TrimSuffix(normalized, "i") // pomidori -> pomidor
	return normalized
}

// sortRecipeMatches sorts matches by: canCookNow DESC -> score DESC -> costToComplete ASC -> timeMinutes ASC
func sortRecipeMatches(matches []RecipeMatch) {
	// Simple bubble sort (good enough for small datasets)
	for i := 0; i < len(matches); i++ {
		for j := i + 1; j < len(matches); j++ {
			// Primary sort: canCookNow (recipes you can cook NOW go first)
			if matches[i].CanMakeNow != matches[j].CanMakeNow {
				if matches[j].CanMakeNow {
					matches[i], matches[j] = matches[j], matches[i]
				}
				continue
			}

			// Secondary sort: score (higher is better)
			scoreI := matches[i].MatchScore
			scoreJ := matches[j].MatchScore

			// Bonus for expiring items (prioritize waste reduction)
			if matches[i].HasExpiringItems {
				scoreI += 5
			}
			if matches[j].HasExpiringItems {
				scoreJ += 5
			}

			if scoreI != scoreJ {
				if scoreJ > scoreI {
					matches[i], matches[j] = matches[j], matches[i]
				}
				continue
			}

			// Tertiary sort: costToComplete (cheaper is better)
			if matches[i].CostToComplete != matches[j].CostToComplete {
				if matches[j].CostToComplete < matches[i].CostToComplete {
					matches[i], matches[j] = matches[j], matches[i]
				}
				continue
			}

			// Quaternary sort: timeMinutes (faster is better)
			if matches[i].Recipe.TimeMinutes > matches[j].Recipe.TimeMinutes {
				matches[i], matches[j] = matches[j], matches[i]
			}
		}
	}
}

// roundToTwoDecimals rounds a float to 2 decimal places (for money)
func roundToTwoDecimals(value float64) float64 {
	return math.Round(value*100) / 100
}

// GetBestRecommendation возвращает ОДИН лучший рецепт для сценария "Что я могу приготовить СЕЙЧАС?"
// RULES ENGINE v1: ЖЁСТКИЕ ПРАВИЛА БЕЗ КОМПРОМИССОВ
//
// Алгоритм:
// 1. Фильтр: ТОЛЬКО рецепты с coverage == 100% (canMakeNow == true)
// 2. Сортировка: professional > ai → меньше времени → больше expiringSoon
// 3. Выбор: TOP-1
//
// ❗ Никаких partial, никаких "почти", никаких компромиссов
// ❗ Пользователь хочет готовить СЕЙЧАС - без докупок
func (s *RecipeMatchService) GetBestRecommendation(
	userID string,
	limit int,
	excludeRecipeIds []string,
) (*RecipeMatch, error) {
	// ВАЖНО: limit здесь НЕ ограничивает количество результатов для пользователя (всегда 1)
	// limit указывает, сколько ТОП-кандидатов рассматривать ПОСЛЕ фильтрации
	if limit <= 0 {
		limit = 20 // default: смотрим больше кандидатов для лучшего выбора
	}

	// 1. Используем матчинг с ЖЁСТКИМИ фильтрами
	filters := RecipeFilters{
		MinScore:         100,              // ТОЛЬКО 100% coverage (все обязательные ингредиенты есть)
		OnlyCookable:     true,             // ТОЛЬКО рецепты, которые можно приготовить СЕЙЧАС
		Limit:            limit * 2,        // Загружаем больше для лучшего выбора после фильтрации
		ExcludeRecipeIds: excludeRecipeIds, // Исключаем уже показанные рецепты
	}

	matches, err := s.MatchRecipesWithFridge(userID, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to match recipes: %w", err)
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("no recipes available with 100%% coverage")
	}

	// 2. Дополнительная сортировка с приоритетом professional источников
	// Правила (в порядке приоритета):
	// - canMakeNow == true (уже отфильтровано)
	// - coverage == 100% (уже отфильтровано через MinScore)
	// - source.type == "professional" > "ai" > "user"
	// - меньше времени приготовления
	// - больше expiring items (экономим продукты)
	sort.SliceStable(matches, func(i, j int) bool {
		// 1. Приоритет: professional > ai > user
		sourceI := s.getSourcePriority(matches[i].Recipe)
		sourceJ := s.getSourcePriority(matches[j].Recipe)
		if sourceI != sourceJ {
			return sourceI < sourceJ // меньшее значение = выше приоритет
		}

		// 2. Меньше времени приготовления
		if matches[i].Recipe.TimeMinutes != matches[j].Recipe.TimeMinutes {
			return matches[i].Recipe.TimeMinutes < matches[j].Recipe.TimeMinutes
		}

		// 3. Больше expiring items (waste prevention)
		if matches[i].ExpiringItemsCount != matches[j].ExpiringItemsCount {
			return matches[i].ExpiringItemsCount > matches[j].ExpiringItemsCount
		}

		// 4. Выше match score (опционально)
		return matches[i].MatchScore > matches[j].MatchScore
	})

	// 3. Возвращаем первый (лучший) рецепт
	return &matches[0], nil
}

// getSourcePriority возвращает приоритет источника рецепта (меньше = выше приоритет)
func (s *RecipeMatchService) getSourcePriority(recipe models.RecipeCatalog) int {
	// Parse source JSON
	var source map[string]interface{}
	if err := json.Unmarshal(recipe.Source, &source); err != nil {
		return 999 // unknown source - lowest priority
	}

	sourceType, ok := source["type"].(string)
	if !ok {
		return 999
	}

	switch sourceType {
	case "professional":
		return 1 // HIGHEST priority
	case "ai":
		return 2
	case "user":
		return 3
	default:
		return 999 // unknown - lowest priority
	}
}

// GetRecipeByID returns full recipe details by ID (with ingredients)
func (s *RecipeMatchService) GetRecipeByID(recipeID string) (*models.RecipeCatalog, error) {
	var recipe models.RecipeCatalog

	err := s.db.
		Preload("Ingredients").            // Load recipe ingredients
		Preload("Ingredients.Ingredient"). // Load ingredient details
		Preload("Allergens").              // Load allergens
		Preload("DietTags").               // Load diet tags
		Preload("RecipeSteps", func(db *gorm.DB) *gorm.DB {
			return db.Order("step_number ASC") // Load steps in order
		}).
		Where("id = ?", recipeID).
		First(&recipe).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("recipe not found")
		}
		return nil, fmt.Errorf("failed to get recipe: %w", err)
	}

	return &recipe, nil
}

// GetStats returns recipe catalog statistics (no text, only numbers)
func (s *RecipeMatchService) GetStats() (totalRecipes int, byCategory map[string]int, err error) {
	// 1. Get total count
	var count int64
	err = s.db.Model(&models.RecipeCatalog{}).Count(&count).Error
	if err != nil {
		return 0, nil, fmt.Errorf("failed to count recipes: %w", err)
	}
	totalRecipes = int(count)

	// 2. Get counts by category
	type CategoryCount struct {
		Category string
		Count    int
	}

	var categoryCounts []CategoryCount
	err = s.db.Model(&models.RecipeCatalog{}).
		Select("category, COUNT(*) as count").
		Group("category").
		Scan(&categoryCounts).Error
	if err != nil {
		return 0, nil, fmt.Errorf("failed to count by category: %w", err)
	}

	// 3. Build map
	byCategory = make(map[string]int)
	for _, cc := range categoryCounts {
		byCategory[cc.Category] = cc.Count
	}

	return totalRecipes, byCategory, nil
}

// GetFridgeItemCount returns count of user's fridge items (for context in AI responses)
func (s *RecipeMatchService) GetFridgeItemCount(userID string) (int, error) {
	var count int64
	err := s.db.Model(&models.UserFridgeItem{}).
		Where("user_id::text = ?", userID).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count fridge items: %w", err)
	}
	return int(count), nil
}

// ListRecipes returns filtered recipe list from catalog
func (s *RecipeMatchService) ListRecipes(filters RecipeFilters) ([]models.RecipeCatalog, error) {
	query := s.db.Model(&models.RecipeCatalog{})

	// Apply filters
	if filters.Country != "" {
		query = query.Where("country = ?", filters.Country)
	}
	if filters.Category != "" {
		query = query.Where("category = ?", filters.Category)
	}
	if filters.Difficulty != "" {
		query = query.Where("difficulty = ?", filters.Difficulty)
	}
	if filters.MaxTime > 0 {
		query = query.Where("\"timeMinutes\" <= ?", filters.MaxTime)
	}

	// Execute query with limit
	var recipes []models.RecipeCatalog
	limit := filters.Limit
	if limit <= 0 || limit > 100 {
		limit = 20 // Default limit
	}

	err := query.Limit(limit).Find(&recipes).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list recipes: %w", err)
	}

	return recipes, nil
}

// EnrichRecipeWithFridgeInfo checks which recipe ingredients are available in user's fridge
// Updates the inFridge field for each CatalogIngredient
func (s *RecipeMatchService) EnrichRecipeWithFridgeInfo(userID string, recipe *models.RecipeCatalog) error {
	if recipe == nil || len(recipe.Ingredients) == 0 {
		return nil // Nothing to enrich
	}

	// 1. Load user's fridge items
	var fridgeItems []models.UserFridgeItem
	err := s.db.Where("user_id::text = ?", userID).
		Preload("Ingredient"). // Load ingredient details
		Find(&fridgeItems).Error
	if err != nil {
		return fmt.Errorf("failed to load fridge items: %w", err)
	}

	// 2. Build map: ingredientID -> fridgeItem
	fridgeMap := make(map[string]*models.UserFridgeItem)
	for i := range fridgeItems {
		fridgeMap[fridgeItems[i].IngredientID] = &fridgeItems[i]
	}

	// 3. Check each recipe ingredient against fridge
	for i := range recipe.Ingredients {
		recipeIng := &recipe.Ingredients[i]

		if fridgeItem, found := fridgeMap[recipeIng.IngredientID]; found {
			// Ingredient exists in fridge
			recipeIng.FridgeQuantity = &fridgeItem.Quantity

			// Check if quantity is sufficient (same unit)
			if fridgeItem.Unit == recipeIng.Unit && fridgeItem.Quantity >= recipeIng.Quantity {
				recipeIng.InFridge = true
			} else {
				// Exists but insufficient quantity or different unit
				recipeIng.InFridge = false
			}
		} else {
			// Ingredient not in fridge
			recipeIng.InFridge = false
			recipeIng.FridgeQuantity = nil
		}
	}

	return nil
}
