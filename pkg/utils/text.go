package utils

import (
	"strings"
	"unicode"
)

// Capitalize делает первую букву заглавной, остальные оставляет как есть
// Работает корректно с UTF-8 (поддержка кириллицы, польских диакритиков и т.д.)
//
// Примеры:
//   - "bacon" → "Bacon"
//   - "бекон" → "Бекон"
//   - "łosoś" → "Łosoś"
//   - "EGGS" → "EGGS" (не меняет остальные буквы)
//   - "" → ""
//   - "  test  " → "Test" (убирает пробелы по краям)
func Capitalize(s string) string {
	if s == "" {
		return s
	}

	// Убираем пробелы по краям
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}

	// Конвертируем в руны для корректной работы с UTF-8
	runes := []rune(s)

	// Делаем первую руну заглавной
	if len(runes) > 0 {
		runes[0] = unicode.ToUpper(runes[0])
	}

	return string(runes)
}

// CapitalizeWords делает первую букву каждого слова заглавной
// Используется для составных названий типа "olive oil" → "Olive Oil"
//
// Примеры:
//   - "olive oil" → "Olive Oil"
//   - "sea salt" → "Sea Salt"
//   - "красный перец" → "Красный Перец"
func CapitalizeWords(s string) string {
	if s == "" {
		return s
	}

	s = strings.TrimSpace(s)
	words := strings.Fields(s) // Разбиваем по пробелам

	for i, word := range words {
		words[i] = Capitalize(word)
	}

	return strings.Join(words, " ")
}

// CapitalizeSentence capitalizes the first letter of a sentence
// Alias for Capitalize for semantic clarity when working with sentences
func CapitalizeSentence(s string) string {
	return Capitalize(s)
}

// CapitalizeTitle properly capitalizes a recipe title
// - First letter uppercase
// - Rest as provided (preserving proper nouns)
// - Trims extra whitespace
func CapitalizeTitle(s string) string {
	s = strings.TrimSpace(s)
	s = NormalizeWhitespace(s)
	return Capitalize(s)
}

// NormalizeWhitespace removes extra spaces and normalizes whitespace
// Example: "яичница  с   беконом" → "яичница с беконом"
func NormalizeWhitespace(s string) string {
	// Replace multiple spaces with single space
	words := strings.Fields(s)
	return strings.Join(words, " ")
}

// CapitalizeSteps capitalizes each step in a recipe
func CapitalizeSteps(steps []string) []string {
	normalized := make([]string, len(steps))
	for i, step := range steps {
		normalized[i] = Capitalize(step)
	}
	return normalized
}

// CleanRecipeText performs comprehensive text cleaning for recipes
// - Removes extra whitespace
// - Capitalizes first letter
// - Removes trailing dots if present
func CleanRecipeText(s string) string {
	// Normalize whitespace
	s = NormalizeWhitespace(s)
	
	// Capitalize first letter
	s = Capitalize(s)
	
	// Remove trailing dot if it's the only punctuation at the end
	s = strings.TrimSuffix(strings.TrimSpace(s), ".")
	
	return s
}

// SanitizeRecipeInput sanitizes user input for recipe creation
// This is the first line of defense before AI processing
func SanitizeRecipeInput(input string) string {
	// 1. Trim whitespace
	input = strings.TrimSpace(input)
	
	// 2. Normalize whitespace
	input = NormalizeWhitespace(input)
	
	return input
}

// ValidateRecipeTitle checks if a recipe title is valid
func ValidateRecipeTitle(title string) (bool, string) {
	title = strings.TrimSpace(title)
	
	if title == "" {
		return false, "Title cannot be empty"
	}
	
	if len(title) < 3 {
		return false, "Title must be at least 3 characters"
	}
	
	if len(title) > 200 {
		return false, "Title must be less than 200 characters"
	}
	
	// Check if title contains only whitespace
	if strings.TrimSpace(title) == "" {
		return false, "Title cannot contain only whitespace"
	}
	
	return true, ""
}
