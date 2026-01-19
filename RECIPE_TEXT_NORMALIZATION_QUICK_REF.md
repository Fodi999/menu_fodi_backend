# Нормализация рецептов - Краткая справка

## 🎯 Что делает система

### Проблема:
```
❌ "яишница глазунья"        // опечатка, lowercase
❌ "EGGS  with   bacon"     // caps, лишние пробелы  
❌ "Łosoś  🔥"              // emoji
```

### Решение (автоматическое):
```
✅ "Яичница глазунья"
✅ "Eggs with bacon"
✅ "Łosoś"
```

---

## 🏗 3 шага нормализации

### 1. AI исправляет орфографию
- ✅ "яишница" → "яичница"
- ✅ "egs" → "eggs"
- ✅ Убирает emoji
- ✅ Профессиональная терминология

### 2. Backend капитализирует
- ✅ Первая буква заглавная
- ✅ Убирает лишние пробелы
- ✅ Нормализует шаги

### 3. Генерируется slug
- ✅ "Яичница глазунья" → `yaichnitsa_glazunya`
- ✅ Транслитерация Cyrillic/Polish
- ✅ URL-friendly

---

## 🔧 Новые функции

```go
// Капитализация
utils.CapitalizeSentence("яичница")     → "Яичница"
utils.CapitalizeTitle("  eggs  ")       → "Eggs"

// Очистка
utils.CleanRecipeText("  яичница.  ")   → "Яичница"
utils.NormalizeWhitespace("a  b  c")    → "a b c"

// Шаги
utils.CapitalizeSteps([]string{"step"}) → ["Step"]

// Slug
utils.GenerateCanonicalName("Борщ")     → "borshch"
utils.Transliterate("Яичница")          → "yaichnitsa"
```

---

## 📊 Примеры

### Input:
```json
{
  "localName": "яишница глазунья",
  "description": "простая  яичница."
}
```

### Output в БД:
```json
{
  "localName": "Яичница глазунья",
  "title": "Яичница глазунья",
  "canonicalName": "yaichnitsa_glazunya",
  "description": "Простая яичница"
}
```

---

## ✅ Что реализовано

- ✅ AI промпт обновлен (исправление орфографии)
- ✅ Утилиты нормализации созданы (`pkg/utils/text.go`)
- ✅ Транслитерация (`pkg/utils/canonical_names.go`)
- ✅ Применено в `RecipeAdminService`
- ✅ Автоматическая генерация slug
- ✅ Нормализация переводов (PL/EN/RU)

---

## 📁 Измененные файлы

1. `pkg/utils/text.go` - утилиты нормализации
2. `pkg/utils/canonical_names.go` - генерация slugs
3. `internal/modules/recipes_admin/service/recipe_admin_service.go` - применение
4. `internal/modules/ai/prompts/fridge.go` - обновлен AI prompt

---

**Дата:** 19 января 2026 г.  
**Статус:** ✅ Готово к использованию
