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
