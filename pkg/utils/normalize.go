package utils

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// ============================================================================
// НОРМАЛИЗАЦИЯ НАЗВАНИЙ ПРОДУКТОВ
// Цель: одинаковые продукты всегда дают одинаковый canonicalKey
// ============================================================================

var (
	// Регулярка для удаления всего кроме букв, цифр и пробелов
	regexNonAlphaNum = regexp.MustCompile(`[^a-zа-я0-9\s]`)
	
	// Регулярка для схлопывания множественных пробелов
	regexMultiSpace = regexp.MustCompile(`\s+`)
)

// NormalizeName нормализует название продукта для поиска и сравнения
// Используется для:
// 1. Создания normalizedName в IngredientAlias (для поиска дублей)
// 2. Поиска существующих продуктов при добавлении
//
// Правила нормализации:
// - trim
// - toLowerCase
// - ё → е
// - удалить спецсимволы (оставить только буквы, цифры, пробелы)
// - схлопнуть множественные пробелы
// - удалить диакритику (ą→a, ó→o, etc.)
func NormalizeName(name string) string {
	if name == "" {
		return ""
	}

	// 1. Trim
	result := strings.TrimSpace(name)

	// 2. Lowercase
	result = strings.ToLower(result)

	// 3. Заменить ё на е
	result = strings.ReplaceAll(result, "ё", "е")

	// 4. Удалить диакритику (ą→a, ó→o, ę→e, etc.)
	result = removeDiacritics(result)

	// 5. Удалить все спецсимволы кроме букв, цифр и пробелов
	result = regexNonAlphaNum.ReplaceAllString(result, "")

	// 6. Схлопнуть множественные пробелы в один
	result = regexMultiSpace.ReplaceAllString(result, " ")

	// 7. Финальный trim
	result = strings.TrimSpace(result)

	return result
}

// GenerateCanonicalKey генерирует уникальный ключ для продукта
// Используется только при создании нового CanonicalIngredient
//
// Правила:
// - как NormalizeName, но пробелы заменяются на дефис
// - максимум 100 символов
//
// Примеры:
//   "Лук репчатый" → "лук-репчатый"
//   "Onion" → "onion"
//   "Pierś z kurczaka" → "piers-z-kurczaka"
func GenerateCanonicalKey(name string) string {
	normalized := NormalizeName(name)
	
	// Заменяем пробелы на дефисы
	key := strings.ReplaceAll(normalized, " ", "-")
	
	// Ограничиваем длину
	if len(key) > 100 {
		key = key[:100]
	}
	
	// Убираем trailing дефисы
	key = strings.TrimRight(key, "-")
	
	return key
}

// removeDiacritics удаляет диакритические знаки (ą→a, ó→o, ę→e)
func removeDiacritics(s string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, _ := transform.String(t, s)
	return result
}

// ============================================================================
// ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ
// ============================================================================

// AreNamesEqual проверяет, являются ли два названия одинаковыми после нормализации
func AreNamesEqual(name1, name2 string) bool {
	return NormalizeName(name1) == NormalizeName(name2)
}

// ExtractBaseWord извлекает базовое слово из составного названия
// Пример: "Лук репчатый" → "лук", "Pierś z kurczaka" → "piers"
// Полезно для поиска похожих продуктов
func ExtractBaseWord(name string) string {
	normalized := NormalizeName(name)
	words := strings.Fields(normalized)
	if len(words) > 0 {
		return words[0]
	}
	return normalized
}

// SimilarityScore вычисляет схожесть двух названий (0.0 - 1.0)
// Простая реализация на основе Levenshtein distance
func SimilarityScore(name1, name2 string) float64 {
	n1 := NormalizeName(name1)
	n2 := NormalizeName(name2)
	
	if n1 == n2 {
		return 1.0
	}
	
	distance := levenshteinDistance(n1, n2)
	maxLen := max(len(n1), len(n2))
	
	if maxLen == 0 {
		return 1.0
	}
	
	return 1.0 - float64(distance)/float64(maxLen)
}

// levenshteinDistance вычисляет расстояние Левенштейна
func levenshteinDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	// Создаем матрицу
	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
		matrix[i][0] = i
	}
	for j := 0; j <= len(s2); j++ {
		matrix[0][j] = j
	}

	// Заполняем матрицу
	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 1
			if s1[i-1] == s2[j-1] {
				cost = 0
			}
			matrix[i][j] = min(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len(s1)][len(s2)]
}

func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
