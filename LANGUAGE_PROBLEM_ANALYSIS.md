# 🚨 КРИТИЧЕСКАЯ ПРОБЛЕМА: AI генерирует без языкового контекста

**Дата проверки:** 11 января 2026  
**Статус:** ❌ ПРОБЛЕМА ПОДТВЕРЖДЕНА

---

## 📋 Что Обнаружено

### ✅ Что РАБОТАЕТ:

1. **`SuggestIngredients` (autocomplete)** - ✅ ПРАВИЛЬНО
   - Файл: `internal/modules/admin/transport/http/handlers.go:888`
   - Код:
     ```go
     acceptLang := r.Header.Get("Accept-Language")
     lang := normalizeLang(acceptLang)
     suggestions, err := h.service.SuggestIngredients(query, limit, lang)
     ```
   - ✅ Читает `Accept-Language` из заголовка
   - ✅ Передает `lang` в service
   - ✅ База данных возвращает локализованные названия

2. **AI Prompt Template** - ✅ ГОТОВ
   - Файл: `internal/modules/admin/service/recipe_ai.go:211`
   - Prompt содержит:
     ```go
     systemPrompt := fmt.Sprintf(`...
     Return the recipe in the language specified: %s
     Description must be in %s language
     ...`, context.Language, context.Language)
     ```
   - ✅ AI умеет работать с языком
   - ✅ Шаблон готов к использованию

---

### ❌ Что НЕ РАБОТАЕТ:

**Handler `CreateRecipeWithAI` НЕ ЧИТАЕТ язык!**

**Файл:** `internal/modules/admin/transport/http/recipe_ai_handlers.go`

**Проблемный код (строки 40-47):**
```go
// Парсим запрос
var req service.CreateRecipeAIRequest
if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
    fmt.Printf("❌ Invalid request body: %v\n", err)
    utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
    return
}

// ❌ ПРОБЛЕМА: Нет чтения Accept-Language заголовка!
// ❌ req.Language НЕ установлен из заголовка
// ❌ req.Language приходит только из body (если frontend отправит)
```

**Что происходит:**
1. Frontend отправляет `Accept-Language: ru`
2. Backend **НЕ ЧИТАЕТ** этот заголовок
3. `req.Language` остается пустым (если не передан в body)
4. AI получает `context.Language = ""` или default `"en"`
5. ❌ **AI генерирует на английском, даже если пользователь русскоязычный**

---

## 🔍 Детальный Анализ Потока Данных

### Текущий Поток (НЕПРАВИЛЬНЫЙ):

```
Frontend                     Backend Handler               Service                    AI
   |                              |                         |                          |
   |-- POST /recipes/create-ai -->|                         |                          |
   |   Accept-Language: ru        |                         |                          |
   |   {                          |                         |                          |
   |     title: "Рецепт",         |                         |                          |
   |     ingredients: [...],      |                         |                          |
   |     rawCookingText: "..."    |                         |                          |
   |   }                          |                         |                          |
   |                              |                         |                          |
   |                              |-- json.Decode(req) ✅   |                          |
   |                              |                         |                          |
   |                              |   ❌ Accept-Language    |                          |
   |                              |      НЕ ПРОЧИТАН        |                          |
   |                              |                         |                          |
   |                              |-- CreateRecipeWithAI -->|                          |
   |                              |   req.Language = "" ❌  |                          |
   |                              |                         |                          |
   |                              |                         |-- generateRecipeViaAI -->|
   |                              |                         |   context.Language=""❌ |
   |                              |                         |                          |
   |                              |                         |                          |-- AI generates
   |                              |                         |                          |   in English ❌
   |                              |                         |<-- AIResponse (EN) ------|
   |                              |<-- Recipe (EN) ---------|                          |
   |<-- Response (EN) ------------|                         |                          |
   |   ❌ Пользователь получил    |                         |                          |
   |      рецепт на английском,   |                         |                          |
   |      хотя ожидал русский     |                         |                          |
```

---

### Правильный Поток (КАК ДОЛЖНО БЫТЬ):

```
Frontend                     Backend Handler               Service                    AI
   |                              |                         |                          |
   |-- POST /recipes/create-ai -->|                         |                          |
   |   Accept-Language: ru        |                         |                          |
   |   {                          |                         |                          |
   |     title: "Рецепт",         |                         |                          |
   |     ingredients: [...],      |                         |                          |
   |     rawCookingText: "..."    |                         |                          |
   |   }                          |                         |                          |
   |                              |                         |                          |
   |                              |-- json.Decode(req) ✅   |                          |
   |                              |                         |                          |
   |                              |-- r.Header.Get(         |                          |
   |                              |    "Accept-Language") ✅|                          |
   |                              |                         |                          |
   |                              |-- req.Language = "ru" ✅|                          |
   |                              |                         |                          |
   |                              |-- CreateRecipeWithAI -->|                          |
   |                              |   req.Language = "ru" ✅|                          |
   |                              |                         |                          |
   |                              |                         |-- generateRecipeViaAI -->|
   |                              |                         |   context.Language="ru"✅|
   |                              |                         |                          |
   |                              |                         |                          |-- AI generates
   |                              |                         |                          |   in Russian ✅
   |                              |                         |<-- AIResponse (RU) ------|
   |                              |<-- Recipe (RU) ---------|                          |
   |<-- Response (RU) ------------|                         |                          |
   |   ✅ Пользователь получил    |                         |                          |
   |      рецепт на русском       |                         |                          |
```

---

## 🛠️ Решение

### Вариант 1: Читать из Accept-Language (РЕКОМЕНДУЕТСЯ)

**Файл:** `internal/modules/admin/transport/http/recipe_ai_handlers.go`

**Добавить после строки 47:**

```go
// Парсим запрос
var req service.CreateRecipeAIRequest
if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
    fmt.Printf("❌ Invalid request body: %v\n", err)
    utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
    return
}

// ✅ ДОБАВИТЬ: Читаем язык из Accept-Language заголовка
if req.Language == "" {
    acceptLang := r.Header.Get("Accept-Language")
    req.Language = normalizeLang(acceptLang) // "pl", "en", "ru"
    fmt.Printf("🌐 Language from Accept-Language: %s → %s\n", acceptLang, req.Language)
}
```

**Преимущества:**
- ✅ Совместимо с существующим `SuggestIngredients`
- ✅ Frontend автоматически отправляет `Accept-Language`
- ✅ Не требует изменений в frontend
- ✅ Стандартный HTTP подход

---

### Вариант 2: Явный язык в body (альтернатива)

Frontend отправляет язык явно:

```typescript
const response = await fetch('/api/admin/recipes/create-ai', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${token}`,
    'Accept-Language': 'ru', // Остается для совместимости
  },
  body: JSON.stringify({
    title: "Рецепт",
    language: "ru", // ✅ Явно в body
    ingredients: [...],
    rawCookingText: "..."
  })
});
```

**Преимущества:**
- ✅ Явный контроль
- ✅ Можно переопределить язык пользователя

**Недостатки:**
- ❌ Требует изменений в frontend
- ❌ Дублирование (заголовок + body)

---

## 📊 Приоритет Исправления

### 🔥 КРИТИЧЕСКИЙ (исправить СЕГОДНЯ):

1. **CreateRecipeWithAI handler** - добавить чтение `Accept-Language`
2. **PreviewRecipeWithAI handler** - добавить чтение `Accept-Language`
3. Тестирование:
   ```bash
   curl -X POST http://localhost:8080/api/admin/recipes/preview-ai \
     -H "Accept-Language: ru" \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer $TOKEN" \
     -d '{
       "title": "Тестовый рецепт",
       "ingredients": [...],
       "rawCookingText": "Обжарить на сковороде"
     }'
   
   # Проверить: response должен быть на русском
   ```

---

## ✅ Что Уже Готово

1. ✅ **AI Prompt шаблон** поддерживает мультиязычность
2. ✅ **База данных** содержит локализованные ингредиенты (pl/en/ru)
3. ✅ **SuggestIngredients** правильно работает с языками
4. ✅ **normalizeLang()** функция уже существует
5. ✅ **Frontend** отправляет `Accept-Language`

**Осталось:** Добавить **2 строки кода** в handler! 🎯

---

## 🎯 Тестовый Сценарий

После исправления:

```bash
# Тест 1: Русский язык
curl -X POST http://localhost:8080/api/admin/recipes/preview-ai \
  -H "Accept-Language: ru" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Паста Карбонара",
    "ingredients": [{"ingredientId": "...", "quantity": 200, "unit": "g"}],
    "rawCookingText": "Отварить пасту, добавить соус"
  }'

# Ожидаемый результат:
# {
#   "title": "Паста Карбонара",
#   "language": "ru",
#   "description": "Классическое итальянское блюдо с кремовым соусом",
#   "steps": [
#     {"order": 1, "text": "Отварить пасту аль денте", "time": 10}
#   ]
# }
```

```bash
# Тест 2: Польский язык
curl -X POST http://localhost:8080/api/admin/recipes/preview-ai \
  -H "Accept-Language: pl" \
  -H "Content-Type: application/json" \
  -d '{ ... }'

# Ожидаемый результат:
# {
#   "language": "pl",
#   "description": "Klasyczne włoskie danie z kremowym sosem",
#   "steps": [{"text": "Ugotować makaron al dente", ...}]
# }
```

---

## 🔗 Связанные Файлы

1. **Handler (ТРЕБУЕТ ИСПРАВЛЕНИЯ):**
   - `internal/modules/admin/transport/http/recipe_ai_handlers.go`
   - Функции: `CreateRecipeWithAI`, `PreviewRecipeWithAI`

2. **Service (УЖЕ ГОТОВ):**
   - `internal/modules/admin/service/recipe_ai.go`
   - Функция: `generateRecipeViaAI` - поддерживает `context.Language`

3. **Эталонный handler (ПРАВИЛЬНЫЙ ПРИМЕР):**
   - `internal/modules/admin/transport/http/handlers.go`
   - Функция: `SuggestIngredients` - показывает, как читать `Accept-Language`

4. **Utility функция:**
   - `internal/modules/admin/transport/http/handlers.go:927`
   - Функция: `normalizeLang(acceptLang string)` - готова к использованию

---

## 📝 Summary

**Проблема:** AI генерирует рецепты на английском для всех пользователей, игнорируя их язык.

**Причина:** Handler не читает `Accept-Language` заголовок.

**Решение:** Добавить 2 строки кода в handler для чтения заголовка.

**Влияние:** 
- ❌ **Сейчас:** Русскоязычные пользователи получают английские рецепты
- ✅ **После fix:** Каждый пользователь получает рецепт на своем языке

**Время на исправление:** 5 минут

**Приоритет:** 🔥 КРИТИЧЕСКИЙ
