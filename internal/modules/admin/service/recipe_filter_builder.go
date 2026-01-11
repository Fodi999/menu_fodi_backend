package service

import (
	"fmt"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"gorm.io/gorm"
)

// ApplyRecipeFilters - декларативный фильтр-билдер
// Применяет фильтры к GORM query независимо друг от друга
func ApplyRecipeFilters(db *gorm.DB, filter RecipeFilter) *gorm.DB {
	query := db.Model(&models.RecipeCatalog{})

	// 1. Status filter
	if filter.Status != nil && *filter.Status != "" {
		query = query.Where("status = ?", *filter.Status)
	}

	// 2. Category filter
	if filter.Category != nil && *filter.Category != "" {
		query = query.Where("category = ?", *filter.Category)
	}

	// 3. Difficulty filter
	if filter.Difficulty != nil && *filter.Difficulty != "" {
		query = query.Where("difficulty = ?", *filter.Difficulty)
	}

	// 4. Time filters
	if filter.TimeLte != nil {
		query = query.Where("\"timeMinutes\" <= ?", *filter.TimeLte)
	}
	if filter.TimeGte != nil {
		query = query.Where("\"timeMinutes\" >= ?", *filter.TimeGte)
	}

	// 5. Calories filters (from nutrition_profile JSONB)
	if filter.CaloriesLte != nil {
		query = query.Where("(\"nutritionProfile\"->>'calories')::int <= ?", *filter.CaloriesLte)
	}
	if filter.CaloriesGte != nil {
		query = query.Where("(\"nutritionProfile\"->>'calories')::int >= ?", *filter.CaloriesGte)
	}

	// 6. Ingredient filter (JOIN with recipe_ingredients)
	if len(filter.IngredientIDs) > 0 {
		query = query.
			Joins("JOIN recipe_ingredients ri ON ri.recipe_id = \"Recipe\".id").
			Where("ri.ingredient_id IN ?", filter.IngredientIDs).
			Group("\"Recipe\".id")
	}

	// 7. Source type filter (from source JSONB)
	if filter.SourceType != nil && *filter.SourceType != "" {
		query = query.Where("source->>'type' = ?", *filter.SourceType)
	}

	// 8. Author ID filter (from source JSONB)
	if filter.AuthorID != nil {
		query = query.Where("source->>'authorId' = ?", filter.AuthorID.String())
	}

	// 9. Sorting (строго ограниченный список)
	query = applySorting(query, filter.Sort)

	// 10. Pagination (ВСЕГДА на backend)
	limit := filter.Limit
	if limit > 50 {
		limit = 50
	}
	offset := (filter.Page - 1) * limit

	query = query.Limit(limit).Offset(offset)

	return query
}

// applySorting - применяет сортировку (whitelist)
func applySorting(query *gorm.DB, sort string) *gorm.DB {
	switch sort {
	case "newest":
		return query.Order("\"createdAt\" DESC")
	case "oldest":
		return query.Order("\"createdAt\" ASC")
	case "time_asc":
		return query.Order("\"timeMinutes\" ASC")
	case "time_desc":
		return query.Order("\"timeMinutes\" DESC")
	case "name_asc":
		return query.Order("title ASC")
	case "name_desc":
		return query.Order("title DESC")
	case "popular":
		// TODO: добавить поле views или popularity
		return query.Order("\"createdAt\" DESC")
	default:
		// Default: newest first
		return query.Order("\"createdAt\" DESC")
	}
}

// GetFilteredRecipesCount - получить общее количество (для meta.total)
func GetFilteredRecipesCount(db *gorm.DB, filter RecipeFilter) (int64, error) {
	var count int64

	// Применяем те же фильтры, но без pagination
	query := db.Model(&models.RecipeCatalog{})

	if filter.Status != nil && *filter.Status != "" {
		query = query.Where("status = ?", *filter.Status)
	}
	if filter.Category != nil && *filter.Category != "" {
		query = query.Where("category = ?", *filter.Category)
	}
	if filter.Difficulty != nil && *filter.Difficulty != "" {
		query = query.Where("difficulty = ?", *filter.Difficulty)
	}
	if filter.TimeLte != nil {
		query = query.Where("\"timeMinutes\" <= ?", *filter.TimeLte)
	}
	if filter.TimeGte != nil {
		query = query.Where("\"timeMinutes\" >= ?", *filter.TimeGte)
	}
	if filter.CaloriesLte != nil {
		query = query.Where("(\"nutritionProfile\"->>'calories')::int <= ?", *filter.CaloriesLte)
	}
	if filter.CaloriesGte != nil {
		query = query.Where("(\"nutritionProfile\"->>'calories')::int >= ?", *filter.CaloriesGte)
	}
	if len(filter.IngredientIDs) > 0 {
		query = query.
			Joins("JOIN recipe_ingredients ri ON ri.recipe_id = \"Recipe\".id").
			Where("ri.ingredient_id IN ?", filter.IngredientIDs).
			Group("\"Recipe\".id")
	}
	if filter.SourceType != nil && *filter.SourceType != "" {
		query = query.Where("source->>'type' = ?", *filter.SourceType)
	}
	if filter.AuthorID != nil {
		query = query.Where("source->>'authorId' = ?", filter.AuthorID.String())
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count recipes: %w", err)
	}

	return count, nil
}
