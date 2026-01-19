package utils

import (
	"regexp"
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
	fallback := Transliterate(normalized)

	return fallback
}

// Transliterate converts Cyrillic and Polish characters to Latin
// and creates a clean URL-friendly slug
// Examples:
//   - "Яичница глазунья" → "yaichnitsa_glazunya"
//   - "Łosoś z grilla" → "losos_z_grilla"
//   - "Борщ украинский" → "borshch_ukrainskiy"
func Transliterate(s string) string {
	s = strings.ToLower(s)
	s = strings.TrimSpace(s)

	// Cyrillic to Latin mapping
	cyrillicMap := map[rune]string{
		'а': "a", 'б': "b", 'в': "v", 'г': "g", 'д': "d",
		'е': "e", 'ё': "yo", 'ж': "zh", 'з': "z", 'и': "i",
		'й': "y", 'к': "k", 'л': "l", 'м': "m", 'н': "n",
		'о': "o", 'п': "p", 'р': "r", 'с': "s", 'т': "t",
		'у': "u", 'ф': "f", 'х': "kh", 'ц': "ts", 'ч': "ch",
		'ш': "sh", 'щ': "shch", 'ъ': "", 'ы': "y", 'ь': "",
		'э': "e", 'ю': "yu", 'я': "ya",
	}

	// Polish to Latin mapping (remove diacritics)
	polishMap := map[rune]string{
		'ą': "a", 'ć': "c", 'ę': "e", 'ł': "l",
		'ń': "n", 'ó': "o", 'ś': "s", 'ź': "z", 'ż': "z",
	}

	var result strings.Builder
	for _, r := range s {
		if lat, ok := cyrillicMap[r]; ok {
			result.WriteString(lat)
		} else if lat, ok := polishMap[r]; ok {
			result.WriteString(lat)
		} else {
			result.WriteRune(r)
		}
	}

	transliterated := result.String()

	// Replace spaces and special characters with underscores
	transliterated = strings.ReplaceAll(transliterated, " ", "_")
	transliterated = strings.ReplaceAll(transliterated, "-", "_")

	// Remove any non-alphanumeric characters (except underscores)
	reg := regexp.MustCompile(`[^a-z0-9_]+`)
	transliterated = reg.ReplaceAllString(transliterated, "_")

	// Remove duplicate underscores
	for strings.Contains(transliterated, "__") {
		transliterated = strings.ReplaceAll(transliterated, "__", "_")
	}

	// Trim underscores from start and end
	transliterated = strings.Trim(transliterated, "_")

	return transliterated
}
