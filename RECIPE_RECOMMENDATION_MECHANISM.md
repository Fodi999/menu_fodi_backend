# 🍳 Механизм рекомендации рецептов по содержимому холодильника

## 📋 Обзор

Система анализирует ингредиенты в холодильнике пользователя и **подбирает рецепты из каталога**, сортируя их по проценту совпадения с доступными ингредиентами.

**Архитектура: Rules Engine** (не AI) — чистая логика, быстро, предсказуемо.

---

## 🔄 Общий поток (High-Level)

```
Frontend request
    ↓
GET /api/recipe-recommendations?lang=ru&limit=10
    ↓
RecommendationHandler (HTTP layer)
    ↓
RecommendationService.GetRecommendations() (Business Logic)
    ↓
[Step 1] Получить ингредиенты из холодильника пользователя
    ↓
[Step 2] Получить ВСЕ рецепты из каталога
    ↓
[Step 3] Для каждого рецепта: matching + расчет match_percent
    ↓
[Step 4] Сортировка по match_percent (DESC)
    ↓
[Step 5] Применить limit + локализацию
    ↓
JSON response с RecipeRecommendationResponse
```

---

## 🏗️ Архитектура модулей

```
internal/modules/ai_recipe_recommendation/
├── service/
│   ├── recommendation_service.go       # 🎯 Основная логика
│   ├── dto.go                          # 📦 Структуры данных
│   └── (других методов в service.go)
├── repository/
│   └── recipe_repository.go            # 📚 Получение рецептов
├── transport/http/
│   └── recommendation_handler.go       # 🌐 HTTP endpoints
└── module.go                           # 🔌 Регистрация маршрутов
```

---

## 📊 Step 1: Получение ингредиентов холодильника

### Метод: `getUserFridgeIngredientIDs()`

```go
func (s *RecommendationService) getUserFridgeIngredientIDs(
    ctx context.Context,
    userID string,
) (map[string]bool, error)
```

**SQL Query:**
```sql
SELECT ufi.ingredient_id, i.canonical_id
FROM user_fridge_items AS ufi
LEFT JOIN "Ingredient" AS i ON i.id = ufi.ingredient_id
WHERE ufi.user_id = ? AND ufi.quantity > 0
```

**Что происходит:**

1. Загружаем только **активные** ингредиенты (quantity > 0)
2. Для каждого ингредиента загружаем TWO ID:
   - `ingredient_id` — точный ID (например, "Olej rzepakowy" = id-123)
   - `canonical_id` — группа (например, "vegetable_oil")

3. Создаем **HashMap** (O(1) lookup):
   ```go
   fridgeSet := map[string]bool{
       "id-123":      true,  // Direct match
       "vegetable_oil": true, // Canonical group
   }
   ```

**Результат:** Можем проверять наличие за O(1):
```go
fridgeSet["id-123"]       // ✅ true (прямое совпадение)
fridgeSet["vegetable_oil"] // ✅ true (группа масел)
```

---

## 🍽️ Step 2: Получение рецептов из каталога

### Метод: `RecipeRepository.GetAllRecipes()`

```go
func (r *RecipeRepository) GetAllRecipes(
    ctx context.Context,
) ([]models.RecipeCatalog, error)
```

**SQL Query (с GORM preloading):**
```sql
SELECT * FROM "RecipeCatalog";
SELECT * FROM "RecipeIngredient" WHERE recipe_id IN (...);
SELECT * FROM "Ingredient" WHERE id IN (...);
```

**Что загружается:**

```go
type RecipeCatalog struct {
    ID             uuid.UUID
    CanonicalName  string              // "borsch", "pasta_carbonara"
    NamePl/En/Ru   *string             // Локализованные названия
    ImageUrl       string
    TimeMinutes    int
    Servings       int
    StepsPl/En/Ru  []byte              // JSON: [{"text": "...", "order": 1}]
    
    Ingredients    []RecipeIngredient   // ⭐ Массив ингредиентов
}

type RecipeIngredient struct {
    ID            uuid.UUID
    RecipeID      uuid.UUID
    IngredientID  string
    IngredientKey string             // Canonical name ("tomato", "olive_oil")
    Quantity      float64            // 300, 2.5, 1
    Unit          string             // "g", "ml", "pcs"
    
    Ingredient    *Ingredient        // ⭐ RELATION (preloaded)
}

type Ingredient struct {
    ID          string
    CanonicalID *string             // ⭐ Canonical group
    NamePl/En/Ru *string            // Локализованные названия
    Category    string              // "vegetable", "protein"
}
```

---

## 🎯 Step 3: Matching & Building RecipeDTO

### Метод: `buildRecipeDTO()`

```go
func (s *RecommendationService) buildRecipeDTO(
    recipe models.RecipeCatalog,
    fridgeIngredientIDs map[string]bool,
    lang string,
) RecipeDTO
```

### Алгоритм matching для каждого ингредиента:

```
For each ingredient in recipe.Ingredients:
    
    1️⃣ Check DIRECT MATCH
        if fridgeSet[ingredient.ID] {
            available ✅
            continue
        }
    
    2️⃣ Check CANONICAL MATCH (если direct не сработал)
        if ingredient.CanonicalID != nil && fridgeSet[*canonical_id] {
            available ✅ (через canonical group)
            continue
        }
    
    3️⃣ Если ни прямого, ни канонического совпадения
        missing ❌
```

### Пример matching:

**Холодильник пользователя:**
```
user_fridge_items:
  - ingredient_id: "id-apple-1" (canonical_id: "fruit_apple")
  - ingredient_id: "id-oil-1"   (canonical_id: "vegetable_oil")
  - ingredient_id: "id-salt-1"  (canonical_id: "salt")

fridgeSet = {
  "id-apple-1": true,
  "fruit_apple": true,
  "id-oil-1": true,
  "vegetable_oil": true,
  "id-salt-1": true,
  "salt": true,
}
```

**Рецепт "Борщ" требует:**
```
recipe_ingredients:
  1. Свекла (id: "id-beet-1", canonical: "root_vegetable")
     - fridgeSet["id-beet-1"]? ❌
     - fridgeSet["root_vegetable"]? ❌
     → MISSING 🔴

  2. Масло растительное (id: "id-oil-2", canonical: "vegetable_oil")
     - fridgeSet["id-oil-2"]? ❌
     - fridgeSet["vegetable_oil"]? ✅ (CANONICAL MATCH!)
     → AVAILABLE 🟢

  3. Соль (id: "id-salt-2", canonical: "salt")
     - fridgeSet["id-salt-2"]? ❌
     - fridgeSet["salt"]? ✅ (CANONICAL MATCH!)
     → AVAILABLE 🟢
```

**Результат для Борща:**
```
available: [масло, соль]         // 2
missing:   [свекла]              // 1
match_percent = 2 / 3 * 100 = 66.7%
match_status = "almost_ready"    // 1-2 missing
```

---

## 📈 Step 4: Расчет метрик

### match_percent (0-100)
```go
matchPercent = float64(availableCount) / float64(totalRequired) * 100
```

**Примеры:**
- 3 из 3 ингредиентов → 100% (ready 🟢)
- 2 из 3 ингредиентов → 66.7% (almost_ready 🟡)
- 1 из 3 ингредиентов → 33.3% (not_ready 🔴)

### match_status (enum)

```go
func classifyMatchStatus(missingCount int) string {
    switch {
    case missingCount == 0:
        return "ready"        // 🟢 Все есть
    case missingCount <= 2:
        return "almost_ready" // 🟡 Не хватает 1-2
    default:
        return "not_ready"    // 🔴 Не хватает 3+
    }
}
```

### Экстракция локализованных шагов

```go
func (s *RecommendationService) extractLocalizedSteps(
    recipe models.RecipeCatalog,
    lang string,
) []string {
    // Выбираем правильное поле JSON
    var stepsJSON []byte
    switch lang {
    case "ru": stepsJSON = recipe.StepsRu
    case "en": stepsJSON = recipe.StepsEn
    case "pl": stepsJSON = recipe.StepsPl
    }
    
    // Parse JSON: [{"text":"Разогреть сковороду","order":1}, ...]
    // Возвращаем []string с шагами
}
```

---

## 🔀 Step 5: Сортировка и limit

### Сортировка (по убыванию match_percent)

```go
func sortRecipesByMatchPercent(recipes []RecipeDTO) {
    // Bubble sort (простой, понимаемый)
    // Рецепты с match_percent 100% → 0%
}
```

**После сортировки:**
```
Рецепт A: 100% (ready)
Рецепт B: 80%  (almost_ready)
Рецепт C: 50%  (not_ready)
Рецепт D: 30%  (not_ready)
```

### Применение limit

```go
if req.Limit > 0 && req.Limit < len(recipeDTOs) {
    recipeDTOs = recipeDTOs[:req.Limit]  // Top 10 (default)
}
```

---

## 📦 Структуры данных (DTO)

### RecipeRecommendationResponse (JSON)

```json
{
  "decision": "almost_ready",
  "summary": "Почти готово! Не хватает всего нескольких ингредиентов.",
  "total_matches": 15,
  "recipes": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "title": "Борщ украинский",
      "canonical_name": "borsch",
      "image_url": "https://res.cloudinary.com/...",
      "cook_time": 60,
      "servings": 4,
      "match_percent": 66.7,
      "match_status": "almost_ready",
      "available_ingredients": [
        {
          "id": "id-oil-1",
          "canonical_name": "vegetable_oil",
          "display_name": "Масло растительное",
          "quantity": 100,
          "unit": "ml",
          "category": "condiment"
        },
        {
          "id": "id-salt-1",
          "canonical_name": "salt",
          "display_name": "Соль",
          "quantity": 5,
          "unit": "g",
          "category": "seasoning"
        }
      ],
      "missing_ingredients": [
        {
          "id": "id-beet-1",
          "canonical_name": "root_vegetable",
          "display_name": "Свекла",
          "quantity": 500,
          "unit": "g",
          "category": "vegetable"
        }
      ],
      "steps": [
        "Разогрейте масло в кастрюле",
        "Нарежьте свеклу кубиками",
        "Добавьте свеклу и варите 30 минут"
      ]
    }
  ]
}
```

---

## 🌐 HTTP API

### Endpoint 1: Получить рекомендации

```
GET /api/recipe-recommendations?lang=ru&limit=10
Authorization: Bearer <JWT>
```

**Query params:**
- `lang` — язык (pl, en, ru) [default: pl]
- `limit` — количество рецептов [default: 10, max: 100]

**Response:**
```json
{
  "decision": "almost_ready",
  "summary": "...",
  "total_matches": 15,
  "recipes": [...]
}
```

---

### Endpoint 2: Получить один рецепт с проверкой холодильника

```
GET /api/recipe-recommendations/{id}?lang=ru
Authorization: Bearer <JWT>
```

**Path params:**
- `id` — UUID или canonical_name рецепта

**Response:** Один RecipeDTO (структура как выше)

---

## 🎨 Frontend интеграция

### Пример fetch:

```javascript
// 1. Получить список рекомендаций
const response = await fetch(
  '/api/recipe-recommendations?lang=ru&limit=10',
  { headers: { 'Authorization': `Bearer ${token}` } }
);

const data = await response.json();
// {
//   decision: "almost_ready",
//   summary: "Почти готово!...",
//   recipes: [{id, title, match_percent, available_ingredients, missing_ingredients, ...}]
// }

// 2. Отобразить рецепты
data.recipes.forEach(recipe => {
  // Показать recipe.title и recipe.image_url
  // Показать green/yellow badge на основе recipe.match_status
  // Список available_ingredients (зелено)
  // Список missing_ingredients (красно) с количеством
});

// 3. При клике на рецепт
const single = await fetch(
  `/api/recipe-recommendations/${recipe.id}?lang=ru`,
  { headers: { 'Authorization': `Bearer ${token}` } }
);
const recipeDetail = await single.json();
// Показать все детали + steps
```

---

## 🔧 Пример: Пошаговое выполнение

### Входные данные:

**Холодильник пользователя:**
```
user_id: "407582be-59d5-4d21-873b-1a72d31b0d42"

user_fridge_items:
  ├─ apple (id-100, canonical: "fruit_apple", qty: 5)
  ├─ olive_oil (id-101, canonical: "vegetable_oil", qty: 250)
  └─ salt (id-102, canonical: "salt", qty: 1000)
```

**Каталог рецептов:**
```
Recipe 1: "Apple Pie"
  ├─ apple (id-100)           ✅ available (direct match)
  ├─ flour (id-201)           ❌ missing
  ├─ sugar (id-202)           ❌ missing
  └─ butter (id-203, canonical: "vegetable_oil")  ✅ available (canonical match)
  
  Result: 2/4 = 50%, "not_ready"

Recipe 2: "Simple Salad"
  ├─ apple (id-100)           ✅ available
  ├─ salt (id-102)            ✅ available
  └─ oil (id-104, canonical: "vegetable_oil")  ✅ available (canonical match)
  
  Result: 3/3 = 100%, "ready"
```

### Выход:

```json
{
  "decision": "ready",
  "summary": "Отлично! Вы можете приготовить несколько рецептов прямо сейчас.",
  "total_matches": 2,
  "recipes": [
    {
      "id": "recipe-2",
      "title": "Simple Salad",
      "match_percent": 100,
      "match_status": "ready",
      "available_ingredients": [
        {"display_name": "Apple", "quantity": 5, "unit": "pcs"},
        {"display_name": "Salt", "quantity": 1, "unit": "g"},
        {"display_name": "Oil", "quantity": 250, "unit": "ml"}
      ],
      "missing_ingredients": []
    },
    {
      "id": "recipe-1",
      "title": "Apple Pie",
      "match_percent": 50,
      "match_status": "not_ready",
      "available_ingredients": [
        {"display_name": "Apple", "quantity": 5, "unit": "pcs"},
        {"display_name": "Butter", "quantity": 250, "unit": "ml"}
      ],
      "missing_ingredients": [
        {"display_name": "Flour", "quantity": 400, "unit": "g"},
        {"display_name": "Sugar", "quantity": 200, "unit": "g"}
      ]
    }
  ]
}
```

---

## ⚡ Оптимизация

### Что работает быстро:

1. **HashMap для проверки ингредиентов:** O(1) lookup вместо O(n) поиска
2. **Preloading (GORM):** Загружаем всё за 3 SQL запроса вместо N+1
3. **Сортировка в памяти:** Bubble sort простой и понимаемый
4. **Кеширование:** Рецепты редко меняются, можно кешировать GetAllRecipes()

### Что можно улучшить:

- Использовать Quick Sort вместо Bubble Sort
- Добавить Redis кеш для GetAllRecipes()
- Добавить индекс на `canonical_id` в таблице Ingredient

---

## 🔐 Security

1. **Authorization:** Все endpoints требуют Bearer token
2. **User isolation:** Рекомендации основаны ТОЛЬКО на ингредиентах пользователя
3. **SQL injection protection:** GORM параметризованные запросы

---

## 📝 Code locations

| Компонент | Файл | Метод |
|-----------|------|--------|
| HTTP layer | `recommendation_handler.go` | `GetRecommendations()` |
| Business logic | `recommendation_service.go` | `GetRecommendations()`, `buildRecipeDTO()` |
| Fridge check | `recommendation_service.go` | `getUserFridgeIngredientIDs()` |
| Recipe fetch | `recipe_repository.go` | `GetAllRecipes()` |
| Matching | `recommendation_service.go` | Lines 120-140 |
| DTOs | `dto.go` | RecipeDTO, IngredientInfo |

---

## 🎯 Ключевые моменты

✅ **Canonical matching** позволяет найти похожие ингредиенты  
✅ **Rules Engine** (не AI) быстро и предсказуемо  
✅ **Полная информация** (available + missing) в одном ответе  
✅ **Локализация** на 3 языках (pl, en, ru)  
✅ **Сортировка** по готовности (100% первыми)  
✅ **O(1) lookup** благодаря HashMap'у  

