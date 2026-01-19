# Кто добавляет продукты в базу: AI vs Админ

## 🎯 Краткий ответ

**Продукты добавляет AI** при создании ингредиента администратором.

**Процесс:**
1. Админ вводит название на **любом языке** (например, "лосось" или "яйца")
2. Backend отправляет запрос к **Groq AI** (модель llama-3.3-70b)
3. AI **автоматически переводит** на 3 языка + классифицирует
4. Backend сохраняет готовый ингредиент со всеми переводами в PostgreSQL

---

## 🔄 Детальный процесс создания ингредиента

### 1. Админ создает ингредиент через API

**Endpoint:** `POST /api/admin/ingredients`

**Запрос:**
```json
{
  "inputName": "лосось"
}
```

**Админ указывает ТОЛЬКО название** на одном языке - остальное делает AI.

---

### 2. Backend отправляет запрос к Groq AI

**Файл:** `internal/modules/admin/service/service.go` (функция `ClassifyIngredient`)

**System Prompt для AI:**
```
You are a multilingual ingredient classifier.
Given an ingredient name in ANY language (Russian, Polish, English, etc.),
provide ALL THREE translations plus classification.

Categories (culinary): fish, meat, egg, vegetable, fruit, dairy, grain, condiment, other
Nutrition Groups: protein, carbohydrate, fat, vegetable, fruit, dairy, condiment, other
Units: g (граммы), ml (миллилитры), pcs (штуки)

Respond ONLY with JSON:
{
  "name_pl": "polish name",
  "name_en": "english name",
  "name_ru": "russian name",
  "category": "culinary category",
  "nutrition_group": "nutritional role",
  "unit": "g or ml or pcs",
  "normalized_value": "normalized_english_singular"
}
```

**User Prompt:**
```
Classify this ingredient: "лосось"
```

---

### 3. AI генерирует ответ со всеми переводами

**Ответ AI:**
```json
{
  "name_pl": "łosoś",
  "name_en": "salmon",
  "name_ru": "лосось",
  "category": "fish",
  "nutrition_group": "protein",
  "unit": "g",
  "normalized_value": "salmon"
}
```

**AI автоматически:**
- ✅ Переводит на польский (name_pl)
- ✅ Переводит на английский (name_en)
- ✅ Переводит на русский (name_ru)
- ✅ Определяет категорию (category)
- ✅ Определяет группу питания (nutrition_group)
- ✅ Выбирает единицу измерения (unit)
- ✅ Нормализует название (normalized_value)

---

### 4. Backend валидирует ответ AI

**Файл:** `internal/modules/admin/service/service.go` (строки 991-1038)

**Проверки:**
```go
// Все переводы присутствуют?
if classification.NamePL == "" || classification.NameEN == "" || classification.NameRU == "" {
    return nil, fmt.Errorf("AI returned incomplete translations")
}

// Все поля классификации заполнены?
if classification.Category == "" || classification.NutritionGroup == "" || 
   classification.Unit == "" || classification.NormalizedValue == "" {
    return nil, fmt.Errorf("AI returned incomplete classification")
}

// Category валиден?
validCategories := map[string]bool{
    "fish": true, "meat": true, "egg": true,
    "vegetable": true, "fruit": true, "dairy": true,
    "grain": true, "condiment": true, "other": true,
}
if !validCategories[classification.Category] {
    return nil, fmt.Errorf("invalid category from AI: %s", classification.Category)
}

// Nutrition group валиден?
validNutritionGroups := map[string]bool{
    "protein": true, "carbohydrate": true, "fat": true,
    "vegetable": true, "fruit": true, "dairy": true,
    "condiment": true, "other": true,
}

// Unit валиден?
validUnits := map[string]bool{"g": true, "ml": true, "pcs": true}
```

Если хотя бы одна проверка провалена → ошибка, ингредиент не создается.

---

### 5. Backend проверяет дубликаты

**Файл:** `internal/modules/admin/service/service.go` (функция `CreateIngredientWithAI`, строка 1051)

```go
// Проверка по normalized_value
if existing, exists := s.CheckIngredientExists(classification.NormalizedValue); exists {
    return nil, fmt.Errorf("INGREDIENT_ALREADY_EXISTS: %s (id: %s)", 
        classification.NormalizedValue, existing.ID)
}
```

**Логика:**
- Ищем ингредиент с таким же `normalized_value` (например, "salmon")
- Если найден → ошибка **409 Conflict**
- Если не найден → продолжаем создание ✅

**Зачем normalized_value?**
- "Лосось" → normalized: "salmon"
- "Salmon" → normalized: "salmon"  
- "Łosoś" → normalized: "salmon"

Все варианты приводятся к одному значению → **предотвращает дубликаты** между языками.

---

### 6. Backend создает ингредиент в БД

**Файл:** `internal/modules/admin/service/service.go` (строки 1054-1079)

```go
id := uuid.New().String()
normalized := strings.ToLower(classification.NormalizedValue)

ingredient := &models.Ingredient{
    ID:              id,
    Name:            classification.NameEN, // Legacy поле
    NamePL:          &classification.NamePL,
    NameEN:          &classification.NameEN,
    NameRU:          &classification.NameRU,
    NormalizedValue: &normalized,
    Unit:            classification.Unit,
    Category:        classification.Category,
    NutritionGroup:  classification.NutritionGroup,
    AutoTranslated:  true, // ✅ Флаг: переведено AI
}

// Сохранение в PostgreSQL
if err := s.db.Create(ingredient).Error; err != nil {
    return nil, fmt.Errorf("failed to save to database: %w", err)
}
```

**База данных (таблица `Ingredient`):**
```sql
INSERT INTO "Ingredient" (
    id, name, name_pl, name_en, name_ru, 
    normalized_value, unit, category, nutrition_group, auto_translated
) VALUES (
    '<uuid>', 'salmon', 'łosoś', 'salmon', 'лосось',
    'salmon', 'g', 'fish', 'protein', true
);
```

---

## 📊 Пример: Создание ингредиента "Яйца"

### Шаг 1: Админ отправляет запрос
```bash
POST /api/admin/ingredients
Authorization: Bearer <admin-token>
{
  "inputName": "Яйца"
}
```

### Шаг 2: AI классифицирует
```
🤖 Classifying ingredient 'Яйца' via Groq AI...
```

### Шаг 3: AI возвращает JSON
```json
{
  "name_pl": "jajka",
  "name_en": "eggs",
  "name_ru": "яйца",
  "category": "egg",
  "nutrition_group": "protein",
  "unit": "pcs",
  "normalized_value": "egg"
}
```

### Шаг 4: Backend проверяет дубликаты
```
🔍 Checking if ingredient 'egg' already exists...
✅ No duplicate found
```

### Шаг 5: Backend сохраняет в БД
```sql
INSERT INTO "Ingredient" VALUES (
    '<uuid>',
    'eggs',        -- name (legacy, английское)
    'jajka',       -- name_pl (польское)
    'eggs',        -- name_en (английское)
    'яйца',        -- name_ru (русское)
    'egg',         -- normalized_value
    'pcs',         -- unit
    'egg',         -- category
    'protein',     -- nutrition_group
    true           -- auto_translated
);
```

### Шаг 6: Ответ клиенту
```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "eggs",
    "name_pl": "jajka",
    "name_en": "eggs",
    "name_ru": "яйца",
    "normalized_value": "egg",
    "unit": "pcs",
    "category": "egg",
    "nutrition_group": "protein",
    "auto_translated": true
  }
}
```

---

## 🤖 Преимущества AI-переводов

### ✅ Автоматизация
- Админ вводит **1 название** → получает **3 перевода**
- Не нужно вручную переводить каждый ингредиент
- Экономия времени: секунды вместо минут

### ✅ Консистентность
- AI использует **одну терминологию** для всех переводов
- "Salmon" всегда переводится как "łosoś" (PL) и "лосось" (RU)
- Предотвращает опечатки и разночтения

### ✅ Контекстная классификация
- AI понимает **контекст кулинарии**
- Правильно определяет категорию (fish, meat, vegetable)
- Выбирает подходящую единицу измерения (g, ml, pcs)

### ✅ Автоматическая нормализация
- "Яйца" → normalized: "egg" (единственное число)
- "Eggs" → normalized: "egg"
- "Jajka" → normalized: "egg"
- **Предотвращает дубликаты** между языками

### ✅ Валидация
- Backend проверяет **все поля** перед сохранением
- Если AI вернул некорректные данные → ошибка
- Гарантирует качество данных в БД

---

## 🔧 Технические детали

### Модель AI
- **Провайдер:** Groq Cloud
- **Модель:** llama-3.3-70b-versatile
- **Тип:** Large Language Model (LLM)
- **Обучение:** Multilingual (поддержка 50+ языков)

### Поддерживаемые языки ввода
- 🇷🇺 Русский (лосось, яйца, рис)
- 🇵🇱 Польский (łosoś, jajka, ryż)
- 🇬🇧 Английский (salmon, eggs, rice)
- 🇩🇪 Немецкий (Lachs, Eier, Reis)
- 🇫🇷 Французский (saumon, œufs, riz)
- ... и другие

AI автоматически определяет язык ввода и переводит на все три целевых языка.

### Время обработки
- **AI классификация:** ~1-2 секунды
- **Проверка дубликатов:** ~50ms
- **Сохранение в БД:** ~20ms
- **Общее время:** ~1.5-2.5 секунды

### Стоимость
- Groq API: **бесплатно** (в рамках лимитов)
- Лимит: ~6000 tokens/min
- 1 ингредиент ≈ 300 tokens
- Лимит: ~20 ингредиентов/минуту

---

## 📋 Флаг auto_translated

**В базе данных:**
```sql
auto_translated BOOLEAN DEFAULT false
```

**Значение:**
- `true` → Переведено AI автоматически ✅
- `false` → Переведено админом вручную (старые данные)

**Зачем нужен?**
- Отслеживание **источника переводов**
- Возможность **пересоздать переводы** в будущем (если AI улучшится)
- Аналитика качества данных

**Текущее состояние:**
- **Все новые ингредиенты** → `auto_translated = true`
- Старые ингредиенты (до внедрения AI) → `auto_translated = false`

---

## 🔍 Как проверить переводы в БД

### SQL запрос
```sql
SELECT 
    name,
    name_pl AS "Польский",
    name_en AS "Английский", 
    name_ru AS "Русский",
    category,
    nutrition_group,
    unit,
    auto_translated
FROM "Ingredient"
WHERE auto_translated = true
ORDER BY created_at DESC
LIMIT 10;
```

### Пример вывода
```
name      | Польский | Английский | Русский | category  | nutrition_group | unit | auto_translated
----------|----------|------------|---------|-----------|-----------------|------|----------------
salmon    | łosoś    | salmon     | лосось  | fish      | protein         | g    | true
eggs      | jajka    | eggs       | яйца    | egg       | protein         | pcs  | true
rice      | ryż      | rice       | рис     | grain     | carbohydrate    | g    | true
olive_oil | oliwa    | olive oil  | масло   | condiment | fat             | ml   | true
```

---

## 🚨 Что делать если AI ошибся?

### Проблема: Неправильный перевод
**Пример:** AI перевел "salmon" как "salmonella" (ошибка)

**Решение:**
1. **Не создавать ингредиент** через AI
2. **Удалить ошибочный** через DELETE /api/admin/ingredients/{id}
3. **Создать вручную** с правильными переводами (если будет такой endpoint)

### Проблема: Неправильная категория
**Пример:** AI классифицировал "tomato" как "vegetable" вместо "fruit"

**Решение:**
1. **Удалить ингредиент**
2. **Добавить уточнение в prompt** (если повторяется)
3. **Пересоздать** через AI

### Проблема: AI не может классифицировать
**Пример:** Экзотический ингредиент "durian"

**Ошибка:**
```
AI classification failed: timeout / invalid response
```

**Решение:**
1. Проверить **Groq API статус** (https://status.groq.com)
2. Попробовать **позже**
3. **Упростить название** (убрать диакритику, спецсимволы)

---

## 🔗 Связанные файлы

### Backend
- **Service:** `internal/modules/admin/service/service.go`
  - `ClassifyIngredient()` - AI классификация (строки 890-1038)
  - `CreateIngredientWithAI()` - создание ингредиента (строки 1042-1079)
  - `CheckIngredientExists()` - проверка дубликатов
- **Handler:** `internal/modules/admin/transport/http/handlers.go`
  - `CreateIngredient()` - HTTP endpoint (строка 1042)
- **Model:** `internal/models/ingredient.go`
  - Структура `Ingredient` с полями name_pl, name_en, name_ru
- **DTO:** `internal/modules/admin/dto/ingredient_ai.go`
  - `CreateIngredientAIRequest` - входной JSON
  - `CreateIngredientAIResponse` - выходной JSON

### Документация
- `docs/AI_INGREDIENT_API_CONTRACT.md` - API документация
- `docs/AI_INGREDIENT_IMPLEMENTATION.md` - детали реализации
- `docs/INGREDIENT_LOCALIZATION_COMPLETE.md` - система локализации

---

## 🎯 Итого

**Кто добавляет переводы?**
- ✅ **AI (Groq llama-3.3-70b)** - автоматически при создании ингредиента

**Что делает админ?**
- ✅ Вводит название на **любом языке**
- ✅ Нажимает "Создать"
- ✅ Получает ингредиент с **3 переводами + классификацией**

**Что делает AI?**
- ✅ Переводит на польский, английский, русский
- ✅ Определяет категорию (fish, meat, vegetable...)
- ✅ Определяет группу питания (protein, carbohydrate, fat...)
- ✅ Выбирает единицу измерения (g, ml, pcs)
- ✅ Нормализует название (salmon, egg, rice...)

**Что делает backend?**
- ✅ Валидирует ответ AI
- ✅ Проверяет дубликаты
- ✅ Сохраняет в PostgreSQL
- ✅ Устанавливает флаг `auto_translated = true`

---

**Status:** ✅ Production ready  
**AI Model:** Groq llama-3.3-70b-versatile  
**Languages:** Polish, English, Russian  
**Created:** 2026-01-19
