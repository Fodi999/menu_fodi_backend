package service

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// RecipeFilter - DTO для фильтрации рецептов
type RecipeFilter struct {
	// Search & Language
	Search string  `json:"search"` // Полнотекстовый поиск по названию
	Lang   string  `json:"lang"`   // ru | en | pl
	Status *string `json:"status"` // published | draft

	// Categories & Difficulty
	Category   *string `json:"category"`   // appetizer | main | dessert | soup | salad
	Difficulty *string `json:"difficulty"` // easy | medium | hard

	// Time & Nutrition
	TimeLte     *int `json:"timeLte"`     // <= N минут
	TimeGte     *int `json:"timeGte"`     // >= N минут
	CaloriesLte *int `json:"caloriesLte"` // <= N калорий
	CaloriesGte *int `json:"caloriesGte"` // >= N калорий

	// Ingredients (JOIN filtering)
	IngredientIDs []uuid.UUID `json:"ingredientIds"` // Фильтр по ингредиентам

	// Source & Author
	SourceType *string    `json:"sourceType"` // ai | manual | traditional
	AuthorID   *uuid.UUID `json:"authorId"`   // Создатель рецепта

	// Pagination
	Page  int `json:"page"`  // Номер страницы (default: 1)
	Limit int `json:"limit"` // Кол-во на странице (default: 20, max: 50)

	// Sorting
	Sort string `json:"sort"` // newest | popular | time_asc | time_desc | name_asc | name_desc
}

// ParseRecipeFilter - централизованный парсер query параметров
func ParseRecipeFilter(r *http.Request) RecipeFilter {
	query := r.URL.Query()
	filter := RecipeFilter{
		Search: query.Get("search"), // Полнотекстовый поиск
		Lang:   query.Get("lang"),   // ru | en | pl
		Sort:   query.Get("sort"),   // newest (default)
	}

	// Status
	if status := query.Get("status"); status != "" {
		filter.Status = &status
	}

	// Category
	if category := query.Get("category"); category != "" {
		filter.Category = &category
	}

	// Difficulty
	if difficulty := query.Get("difficulty"); difficulty != "" {
		filter.Difficulty = &difficulty
	}

	// Time filters
	if timeLte := query.Get("timeLte"); timeLte != "" {
		if val, err := strconv.Atoi(timeLte); err == nil {
			filter.TimeLte = &val
		}
	}
	if timeGte := query.Get("timeGte"); timeGte != "" {
		if val, err := strconv.Atoi(timeGte); err == nil {
			filter.TimeGte = &val
		}
	}

	// Calories filters
	if caloriesLte := query.Get("caloriesLte"); caloriesLte != "" {
		if val, err := strconv.Atoi(caloriesLte); err == nil {
			filter.CaloriesLte = &val
		}
	}
	if caloriesGte := query.Get("caloriesGte"); caloriesGte != "" {
		if val, err := strconv.Atoi(caloriesGte); err == nil {
			filter.CaloriesGte = &val
		}
	}

	// Ingredient IDs (comma-separated)
	if ingredientIDs := query.Get("ingredientIds"); ingredientIDs != "" {
		ids := strings.Split(ingredientIDs, ",")
		for _, id := range ids {
			if parsed, err := uuid.Parse(strings.TrimSpace(id)); err == nil {
				filter.IngredientIDs = append(filter.IngredientIDs, parsed)
			}
		}
	}

	// Source type
	if sourceType := query.Get("sourceType"); sourceType != "" {
		filter.SourceType = &sourceType
	}

	// Author ID
	if authorID := query.Get("authorId"); authorID != "" {
		if parsed, err := uuid.Parse(authorID); err == nil {
			filter.AuthorID = &parsed
		}
	}

	// Pagination
	filter.Page = 1
	if page := query.Get("page"); page != "" {
		if val, err := strconv.Atoi(page); err == nil && val > 0 {
			filter.Page = val
		}
	}

	filter.Limit = 20 // default
	if limit := query.Get("limit"); limit != "" {
		if val, err := strconv.Atoi(limit); err == nil && val > 0 {
			if val > 50 {
				filter.Limit = 50 // max limit
			} else {
				filter.Limit = val
			}
		}
	}

	return filter
}
