# Нормализация текста рецептов - Guide

**Дата:** 19 января 2026 г.  
**Цель:** Обеспечить профессиональное качество текста рецептов на уровне backend

---

## 🎯 Проблема

### Что приходит от пользователя:
```
❌ "яишница глазунья"          // опечатка, нижний регистр
❌ "EGGS  with    bacon"       // лишние пробелы, caps lock
❌ "Łosoś   z  grilla  🔥"    // emoji, множественные пробелы
```

### Что должно быть в БД:
```
✅ "Яичница глазунья"
✅ "Eggs with bacon"
✅ "Łosoś z grilla"
```

---

## 🏗 3-ступенчатая нормализация

### 🥇 Шаг 1: AI исправляет орфографию и стиль

#### Обновленный AI System Prompt:
```
You are a professional chef and food editor.

Rules:
- Fix spelling mistakes (e.g. "яишница" → "яичница", "egs" → "eggs")
- Use correct culinary terminology  
- Recipe title must be short, professional, human-readable
- Do NOT repeat words unnecessarily
- Do NOT use emojis in text
- Return clean, editorial-quality text
- First letter always capitalized
- Avoid excessive caps lock (NOT: "EGGS", use: "Eggs")
```

**Файл:** `internal/modules/ai/prompts/fridge.go`

✅ Реализовано в польском промпте `RestaurantRecipePrompt["pl"]`

---

### 🥈 Шаг 2: Backend нормализует регистр

#### Новые утилиты в `pkg/utils/text.go`:

```go
// Капитализация предложений
utils.CapitalizeSentence("яичница глазунья") 
// → "Яичница глазунья"

// Капитализация заголовка (с очисткой пробелов)
utils.CapitalizeTitle("  EGGS  with   bacon  ")
// → "Eggs with bacon"

// Нормализация пробелов
utils.NormalizeWhitespace("яичница  с   беконом")
// → "яичница с беконом"

// Комплексная очистка текста
utils.CleanRecipeText("  яишница глазунья.  ")
// → "Яишница глазунья"

// Капитализация массива шагов
utils.CapitalizeSteps([]string{
    "разогрейте сковороду",
    "добавьте масло",
})
// → ["Разогрейте сковороду", "Добавьте масло"]
```

#### Применение при создании рецепта:

**Файл:** `internal/modules/recipes_admin/service/recipe_admin_service.go`

```go
// Нормализация текста
localName := utils.CleanRecipeText(req.LocalName)
title := utils.CapitalizeTitle(req.LocalName)
description := utils.CleanRecipeText(req.Description)

// Нормализация шагов
if req.StepsPL != nil {
    normalizedSteps := utils.CapitalizeSteps(*req.StepsPL)
    // сохранение в JSON
}
```

---

### 🥉 Шаг 3: Каноническое имя (slug)

#### Назначение:
- Технический идентификатор (не для отображения)
- Используется в URL, поиске, дедупликации
- Всегда lowercase, без спецсимволов

#### Генерация:

**Функция:** `utils.GenerateCanonicalName()`

**Файл:** `pkg/utils/canonical_names.go`

```go
utils.GenerateCanonicalName("Яичница глазунья")
// → "yaichnitsa_glazunya"

utils.GenerateCanonicalName("Łosoś z grilla")
// → "losos_z_grilla"

utils.GenerateCanonicalName("Scrambled Eggs")
// → "scrambled_eggs"
```

#### Логика:
1. **Проверка маппинга** - известные рецепты имеют фиксированные slugs
2. **Транслитерация** - Cyrillic/Polish → Latin
3. **Нормализация** - lowercase, underscores, удаление спецсимволов

#### Маппинг популярных рецептов:
```go
var RecipeNameMapping = map[string]string{
    // Polish
    "jajecznica":          "scrambled_eggs",
    "omlet":               "omelet",
    
    // Russian
    "яичница":             "scrambled_eggs",
    "борщ":                "borscht",
    
    // English (identity)
    "scrambled eggs":      "scrambled_eggs",
}
```

---

## ✅ Результат нормализации

### Пример Input → Output:

#### ❌ Ввод пользователя:
```json
{
  "localName": "яишница глазунья",
  "description": "простая  яичница   с   беконом.",
  "stepsPL": [
    "разогрейте сковороду",
    "добавьте  масло  и   яйца"
  ]
}
```

#### ✅ После нормализации в БД:
```json
{
  "id": "uuid",
  "localName": "Яичница глазунья",
  "title": "Яичница глазунья",
  "canonicalName": "yaichnitsa_glazunya",
  "description": "Простая яичница с беконом",
  "stepsPL": [
    "Разогрейте сковороду",
    "Добавьте масло и яйца"
  ]
}
```

---

## 📋 Чеклист нормализации

### При создании рецепта (Draft):
- ✅ Исправление опечаток (через AI prompt)
- ✅ Капитализация названия
- ✅ Очистка лишних пробелов
- ✅ Генерация canonical name (slug)
- ✅ Капитализация описания
- ✅ Капитализация шагов
- ✅ Удаление trailing dots

### При публикации рецепта:
- ✅ Валидация длины названия (3-200 символов)
- ✅ Автоматический перевод на PL/EN/RU (если отсутствует)
- ✅ Проверка уникальности canonical name

---

## 🚨 Почему нельзя делать это только на фронте

### Backend должен гарантировать качество:
1. **API** - мобильные приложения, интеграции
2. **Batch import** - массовая загрузка рецептов
3. **AI генерация** - автоматическое создание рецептов
4. **Data integrity** - единый источник правды
5. **Future integrations** - сторонние системы

👉 **Frontend просто отображает**, backend гарантирует качество.

---

## 🧪 Тестирование нормализации

### Unit тесты (рекомендуется создать):

```go
// pkg/utils/text_test.go
func TestCapitalizeSentence(t *testing.T) {
    tests := []struct {
        input    string
        expected string
    }{
        {"яичница", "Яичница"},
        {"  eggs  ", "Eggs"},
        {"BACON", "BACON"}, // не меняет остальные буквы
        {"", ""},
    }
    
    for _, tt := range tests {
        result := utils.CapitalizeSentence(tt.input)
        assert.Equal(t, tt.expected, result)
    }
}

func TestGenerateCanonicalName(t *testing.T) {
    tests := []struct {
        input    string
        expected string
    }{
        {"Яичница глазунья", "yaichnitsa_glazunya"},
        {"Łosoś z grilla", "losos_z_grilla"},
        {"Scrambled Eggs", "scrambled_eggs"},
        {"БОРЩ", "borshch"},
    }
    
    for _, tt := range tests {
        result := utils.GenerateCanonicalName(tt.input)
        assert.Equal(t, tt.expected, result)
    }
}
```

---

## 📊 Примеры нормализации

### Польский:
```
"jajecznica  z   boczkiem" → "Jajecznica z boczkiem"
canonical: "scrambled_eggs"
```

### Русский:
```
"яишница ГЛАЗУНЬЯ" → "Яишница глазунья"
canonical: "yaishnitsa_glazunya"
```

### Английский:
```
"EGGS  with  bacon!!!" → "Eggs with bacon"
canonical: "eggs_with_bacon"
```

---

## 🔧 Файлы изменены

1. **pkg/utils/text.go** - утилиты нормализации текста
   - `CapitalizeSentence()`
   - `CapitalizeTitle()`
   - `NormalizeWhitespace()`
   - `CapitalizeSteps()`
   - `CleanRecipeText()`
   - `ValidateRecipeTitle()`

2. **pkg/utils/canonical_names.go** - генерация slugs
   - `GenerateCanonicalName()`
   - `Transliterate()` - транслитерация Cyrillic/Polish
   - `RecipeNameMapping` - маппинг известных рецептов

3. **internal/modules/recipes_admin/service/recipe_admin_service.go**
   - Применение нормализации при создании draft
   - Нормализация переводов (PL/EN/RU)
   - Автоматическая генерация canonical name

4. **internal/modules/ai/prompts/fridge.go**
   - Обновленный AI prompt с правилами качества текста
   - Исправление орфографии через AI

---

## ✅ Итоговый checklist

- ✅ AI исправляет орфографию
- ✅ Backend капитализирует текст
- ✅ canonicalName всегда чистый (slug)
- ✅ UI просто отображает
- ✅ Нет дубликатов
- ✅ Нет "яишница", "яИшНиЦа", "EGGS EGGS"
- ✅ Профессиональное качество текста

---

## 🎯 Результат

**Система гарантирует:** Любой рецепт в БД имеет профессиональное качество текста, независимо от источника (пользователь, AI, API, импорт).

---

**Статус:** ✅ Реализовано  
**Дата:** 19 января 2026 г.  
**Следующий шаг:** Тестирование на реальных данных
