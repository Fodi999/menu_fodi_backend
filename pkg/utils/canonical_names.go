package utils

import (
	"strings"
)

// RecipeNameMapping - маппинг локализованных названий рецептов на канонические English slugs
var RecipeNameMapping = map[string]string{
	// Polish
	"jajecznica":          "scrambled_eggs",
	"omlet":               "omelet",
	"pierogi":             "pierogi",
	"bigos":               "bigos",
	"zurek":               "sour_rye_soup",
	"rosół":               "chicken_broth",
	"kotlet schabowy":     "breaded_pork_cutlet",
	"gołąbki":             "stuffed_cabbage_rolls",
	"placki ziemniaczane": "potato_pancakes",

	// Russian
	"яичница":        "scrambled_eggs",
	"жареный лосось": "fried_salmon",
	"пельмени":       "dumplings",
	"борщ":           "borscht",
	"блины":          "pancakes",
	"оливье":         "olivier_salad",

	// English (identity mapping for safety)
	"scrambled eggs": "scrambled_eggs",
	"fried salmon":   "fried_salmon",
	"salmon":         "fried_salmon",
}

// GenerateCanonicalName создаёт канонический English slug из любого названия рецепта
// Используется во всей системе для унификации:
// - Admin при создании рецепта
// - AI Recommendation при матчинге
// - Catalog при нормализации
//
// Правила:
// 1. Проверяем маппинг локализованных названий → English slug
// 2. Если не найдено, используем транслитерацию (lowercase + underscores)
// 3. Всегда возвращаем English slug (никогда не локализованное название)
//
// Примеры:
//
//	GenerateCanonicalName("Яичница") → "scrambled_eggs"
//	GenerateCanonicalName("Жареный лосось") → "fried_salmon"
//	GenerateCanonicalName("Scrambled Eggs") → "scrambled_eggs"
//	GenerateCanonicalName("Новый рецепт") → "новый_рецепт" (fallback)
func GenerateCanonicalName(title string) string {
	normalized := strings.ToLower(strings.TrimSpace(title))

	// 1. Проверяем прямой маппинг
	if canonical, exists := RecipeNameMapping[normalized]; exists {
		return canonical
	}

	// 2. Fallback: транслитерация (пробелы → underscores, lowercase)
	// Это временное решение, пока маппинг не будет полным
	fallback := strings.ToLower(strings.ReplaceAll(normalized, " ", "_"))

	return fallback
}
