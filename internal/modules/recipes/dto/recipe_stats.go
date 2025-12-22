package dto

// RecipeStatsResponse - ответ со статистикой каталога рецептов
type RecipeStatsResponse struct {
	Success bool             `json:"success"`
	Data    *RecipeStatsData `json:"data,omitempty"`
}

// RecipeStatsData - данные статистики (БЕЗ текстов, только числа)
type RecipeStatsData struct {
	TotalRecipes int            `json:"totalRecipes"` // Всего рецептов в каталоге
	ByCategory   map[string]int `json:"byCategory"`   // Распределение по категориям
}
