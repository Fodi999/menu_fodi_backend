# 🎯 Визуальное объяснение механизма рекомендаций

## Диаграмма потока данных

```
┌─────────────────────────────────────────────────────────────────┐
│ Frontend GET /api/recipe-recommendations?lang=ru&limit=10      │
└────────────────────┬────────────────────────────────────────────┘
                     │ (JWT token в Authorization)
                     ↓
        ┌────────────────────────┐
        │ RecommendationHandler  │ HTTP layer (тонко)
        │ .GetRecommendations()  │
        └────────────┬───────────┘
                     │
                     ↓
    ┌───────────────────────────────────────┐
    │   RecommendationService               │ Business logic
    │   .GetRecommendations(                │
    │     userID, language, limit)          │
    └──────────┬──────────────────────────┬─┘
               │                          │
        [Step 1]                    [Step 2]
               │                          │
        ┌──────↓──────┐          ┌───────↓────┐
        │ Get Fridge  │          │Get Recipes │
        │Ingredients  │          │from Catalog│
        └──────┬──────┘          └───────┬────┘
               │                         │
        ┌──────↓─────────────────────────↓──────┐
        │   Build Map[id]bool                   │
        │   (direct + canonical)                │
        │                                       │
        │   {                                   │
        │     "id-100": true,   // direct       │
        │     "fruit_apple": true, // canonical│
        │   }                                   │
        └──────┬──────────────────────┬────────┘
               │                      │
               └──────────┬───────────┘
                          │
                    [Step 3-4]
                          │
        ┌─────────────────↓──────────────────┐
        │ For each recipe in catalog:         │
        │                                    │
        │ 1. For each ingredient in recipe:  │
        │    - Check direct match            │
        │    - Check canonical match         │
        │    - Build available/missing lists │
        │                                    │
        │ 2. Calculate metrics:              │
        │    - match_percent = available/    │
        │      total * 100                   │
        │    - match_status = ready/almost/  │
        │      not_ready                     │
        │                                    │
        │ 3. Extract localized steps         │
        │                                    │
        │ 4. Build RecipeDTO                 │
        └─────────────────┬──────────────────┘
                          │
                    [Step 5]
                          │
        ┌─────────────────↓──────────────────┐
        │ Sort by match_percent (DESC)       │
        │ Apply limit (top 10)               │
        │ Analyze overall decision           │
        └─────────────────┬──────────────────┘
                          │
                          ↓
        ┌────────────────────────────────────┐
        │ RecipeRecommendationResponse       │
        │ {                                  │
        │   decision: "almost_ready",        │
        │   summary: "Почти готово!",       │
        │   total_matches: 15,               │
        │   recipes: [...]                   │
        │ }                                  │
        └─────────────────┬──────────────────┘
                          │
                          ↓
                    JSON response
```

---

## Diagram: Matching Logic (3-шаговая проверка)

```
Для каждого ингредиента рецепта:

┌─────────────────────────────────────────────────────┐
│ Ингредиент: "Масло растительное"                   │
│ ID: "id-oil-1"                                      │
│ Canonical: "vegetable_oil"                          │
└────────────┬────────────────────────────────────────┘
             │
             ↓
    ┌────────────────────────┐
    │ Check 1: DIRECT MATCH  │
    │                        │
    │ if fridgeSet["id-oil-1"]
    └────────┬───────────────┘
             │
         YES │ NO
             │  │
        ✅   │  ↓
        AVAIL┤  ┌─────────────────────────────────┐
             │  │ Check 2: CANONICAL MATCH       │
             │  │                                │
             │  │ if canonical != nil &&        │
             │  │ fridgeSet["vegetable_oil"]    │
             │  └────────┬───────────────────────┘
             │           │
             │       YES │ NO
             │           │  │
             │      ✅   │  ↓
             │      AVAIL│  ┌──────────────────┐
             │      (via  │  │ Check 3: ...    │
             │      canon)│  │                  │
             │           │  │ MISSING ❌       │
             │           │  └──────────────────┘
             │           │
             └───────┬───┴─────────┬─────────┘
                     │             │
                     ↓             ↓
                   AVAILABLE    MISSING
                     🟢           🔴
```

---

## Таблица: Примеры matching для разных типов ингредиентов

| Ингредиент | Direct ID | Canonical | В холодильнике | Результат | Логика |
|-----------|----------|-----------|---|---------|---------|
| Масло оливковое | id-oil-1 | vegetable_oil | ❌ | ❌ MISSING | Ни direct, ни canonical |
| Масло растительное | id-oil-2 | vegetable_oil | ✅ (vegetable_oil в холодильнике) | ✅ AVAIL | Canonical match |
| Помидор красный | id-tom-1 | vegetable_tomato | ❌ | ❌ MISSING | Но canonical могут быть есть |
| Помидор черри | id-tom-2 | vegetable_tomato | ✅ (помидор из другого вида) | ✅ AVAIL | Canonical match (оба tomato) |
| Соль морская | id-salt-1 | salt | ✅ (salt в холодильнике) | ✅ AVAIL | Canonical match |
| Перец черный | id-pepper-1 | NULL | ❌ | ❌ MISSING | Нет canonical, нет direct match |

---

## Примеры рецептов с разными процентами

### 🟢 100% Match (Ready) - Паста Карбонара

```
Холодильник:
  ✅ Спагетти (id-1, canonical: "pasta")
  ✅ Бекон (id-2, canonical: "meat_bacon")
  ✅ Яйца (id-3, canonical: "egg")
  ✅ Пармезан (id-4, canonical: "cheese_parmesan")
  ✅ Перец черный (id-5, canonical: "spice_pepper")

Рецепт требует:
  1. Спагетти (id-1) ........................ ✅ MATCH
  2. Бекон (id-2) .......................... ✅ MATCH
  3. Яйца (id-3) ........................... ✅ MATCH
  4. Пармезан (id-4) ....................... ✅ MATCH
  5. Перец (id-5) .......................... ✅ MATCH

Result: 5/5 = 100% ✅ READY 🟢
```

---

### 🟡 66% Match (Almost Ready) - Борщ

```
Холодильник:
  ✅ Масло растительное (id-10, canonical: "vegetable_oil")
  ✅ Соль (id-11, canonical: "salt")
  ❌ Свекла
  ❌ Капуста
  ❌ Говядина

Рецепт требует:
  1. Свекла ........................... ❌ MISSING (нет ни direct, ни canonical)
  2. Капуста ......................... ❌ MISSING
  3. Говядина ........................ ❌ MISSING
  4. Масло (canonical: vegetable_oil) ... ✅ MATCH (canonical)
  5. Соль (canonical: salt) ........... ✅ MATCH (canonical)

Result: 2/5 = 40% 🟡 ALMOST_READY
         (но не в "ready" т.к. много missing)
```

---

### 🔴 20% Match (Not Ready) - Паэлья

```
Холодильник:
  ✅ Масло (id-20, canonical: "vegetable_oil")
  ❌ Рис
  ❌ Морские гады
  ❌ Цыпленок
  ❌ Овощи (перец, помидор)
  ❌ Шафран

Рецепт требует:
  1. Рис ........................... ❌ MISSING
  2. Морские гады ................. ❌ MISSING
  3. Цыпленок ..................... ❌ MISSING
  4. Перец ........................ ❌ MISSING
  5. Помидор ...................... ❌ MISSING
  6. Шафран ....................... ❌ MISSING
  7. Масло (canonical: vegetable_oil) .. ✅ MATCH
  8. Вода ......................... ✅ DEFAULT (всегда есть)

Result: 2/8 = 25% 🔴 NOT_READY
```

---

## Таблица статусов

| Match Percent | Missing Count | Status | UI Badge | Рекомендация |
|--------|--------|--------|---------|---------|
| 100% | 0 | ready 🟢 | GREEN | "Можно готовить прямо сейчас!" |
| 80-99% | 1 | almost_ready 🟡 | YELLOW | "Нужно купить 1 ингредиент" |
| 60-79% | 2 | almost_ready 🟡 | YELLOW | "Нужно купить 2 ингредиента" |
| 40-59% | 3-5 | not_ready 🔴 | RED | "Много нехватает" |
| < 40% | 6+ | not_ready 🔴 | RED | "Очень далеко от готовки" |

---

## Пример JSON для каждого match_status

### 1️⃣ Ready (100%) 

```json
{
  "id": "recipe-123",
  "title": "Паста Карбонара",
  "match_percent": 100,
  "match_status": "ready",
  "available_ingredients": [
    {"display_name": "Спагетти", "quantity": 400, "unit": "g"},
    {"display_name": "Бекон", "quantity": 200, "unit": "g"},
    {"display_name": "Яйца", "quantity": 3, "unit": "pcs"},
    {"display_name": "Пармезан", "quantity": 100, "unit": "g"},
    {"display_name": "Перец черный", "quantity": 5, "unit": "g"}
  ],
  "missing_ingredients": []
}
```

### 2️⃣ Almost Ready (60-79%)

```json
{
  "id": "recipe-456",
  "title": "Борщ",
  "match_percent": 66.7,
  "match_status": "almost_ready",
  "available_ingredients": [
    {"display_name": "Масло растительное", "quantity": 100, "unit": "ml"},
    {"display_name": "Соль", "quantity": 5, "unit": "g"}
  ],
  "missing_ingredients": [
    {"display_name": "Свекла", "quantity": 500, "unit": "g"},
    {"display_name": "Капуста", "quantity": 500, "unit": "g"},
    {"display_name": "Говядина", "quantity": 300, "unit": "g"}
  ]
}
```

### 3️⃣ Not Ready (< 60%)

```json
{
  "id": "recipe-789",
  "title": "Паэлья",
  "match_percent": 25,
  "match_status": "not_ready",
  "available_ingredients": [
    {"display_name": "Масло", "quantity": 100, "unit": "ml"}
  ],
  "missing_ingredients": [
    {"display_name": "Рис", "quantity": 400, "unit": "g"},
    {"display_name": "Морские гады", "quantity": 500, "unit": "g"},
    {"display_name": "Цыпленок", "quantity": 600, "unit": "g"},
    {"display_name": "Перец", "quantity": 2, "unit": "pcs"},
    {"display_name": "Помидор", "quantity": 3, "unit": "pcs"},
    {"display_name": "Шафран", "quantity": 1, "unit": "g"}
  ]
}
```

---

## SQL Query трассировка (для отладки)

### Query 1: Получить ингредиенты холодильника

```sql
-- Что выполняется в getUserFridgeIngredientIDs()
SELECT 
  ufi.ingredient_id,
  i.canonical_id
FROM user_fridge_items AS ufi
LEFT JOIN "Ingredient" AS i ON i.id = ufi.ingredient_id
WHERE ufi.user_id = '407582be-59d5-4d21-873b-1a72d31b0d42'
  AND ufi.quantity > 0;

-- Результат:
-- ingredient_id       | canonical_id
-- id-100 (apple)      | fruit_apple
-- id-101 (oil)        | vegetable_oil
-- id-102 (salt)       | salt
```

### Query 2: Получить все рецепты

```sql
-- Что выполняется в GetAllRecipes()

-- Query 2a: Рецепты
SELECT * FROM "RecipeCatalog" LIMIT 100;

-- Query 2b: Ингредиенты для каждого рецепта
SELECT * FROM "RecipeIngredient" 
WHERE recipe_id IN (uuid1, uuid2, uuid3, ...);

-- Query 2c: Детали ингредиентов
SELECT * FROM "Ingredient" 
WHERE id IN (id-100, id-101, id-102, ...);
```

### Query 3: Результат matching

```
Для рецепта "Apple Pie":

RECIPE INGREDIENT    | ID     | CANONICAL    | IN FRIDGE? | STATUS
─────────────────────┼────────┼──────────────┼────────────┼─────────
apple                | id-100 | fruit_apple  | ✅ YES     | AVAILABLE
flour                | id-201 | NULL         | ❌ NO      | MISSING
sugar                | id-202 | NULL         | ❌ NO      | MISSING
butter               | id-203 | vegetable_oil| ✅ YES*    | AVAILABLE*
                     |        |              |  (canonical)| (via canonical)

Result: 2 available, 2 missing → 50% match
```

---

## Frontend реализация (примеры)

### 1. Получить список

```javascript
async function getRecipeRecommendations(lang = 'ru', limit = 10) {
  const response = await fetch(
    `/api/recipe-recommendations?lang=${lang}&limit=${limit}`,
    {
      headers: { 'Authorization': `Bearer ${getToken()}` }
    }
  );
  
  const data = await response.json();
  return data;
}

// Использование
const recommendations = await getRecipeRecommendations('ru', 10);
console.log(`Всего рецептов: ${recommendations.total_matches}`);
console.log(`Статус: ${recommendations.decision}`); // "ready", "almost_ready", "need_more"
```

### 2. Отобразить рецепты

```javascript
function displayRecipes(recipes) {
  recipes.forEach(recipe => {
    // Выбрать цвет на основе match_status
    const color = {
      'ready': 'green',
      'almost_ready': 'yellow',
      'not_ready': 'red'
    }[recipe.match_status];
    
    // Показать recipe.title и recipe.image_url
    // Badge с recipe.match_percent% и recipe.match_status
    // Список available_ingredients (зелено)
    // Список missing_ingredients (красно)
    
    console.log(`
      ${recipe.title} (${recipe.match_percent}%)
      Available: ${recipe.available_ingredients.map(i => i.display_name).join(', ')}
      Missing: ${recipe.missing_ingredients.map(i => i.display_name).join(', ')}
    `);
  });
}
```

### 3. Получить один рецепт

```javascript
async function getRecipeDetail(recipeId, lang = 'ru') {
  const response = await fetch(
    `/api/recipe-recommendations/${recipeId}?lang=${lang}`,
    {
      headers: { 'Authorization': `Bearer ${getToken()}` }
    }
  );
  
  const recipe = await response.json();
  
  // Показать все детали + steps
  console.log(`
    ${recipe.title}
    ${recipe.cook_time} минут | ${recipe.servings} порций
    
    Доступно:
    ${recipe.available_ingredients.map(i => `  ✅ ${i.display_name} (${i.quantity} ${i.unit})`).join('\n')}
    
    Нужно купить:
    ${recipe.missing_ingredients.map(i => `  ❌ ${i.display_name} (${i.quantity} ${i.unit})`).join('\n')}
    
    Шаги:
    ${recipe.steps.map((s, i) => `  ${i+1}. ${s}`).join('\n')}
  `);
}
```

---

## Performance características

| Operation | Time Complexity | Space Complexity | Notes |
|-----------|-----------------|------------------|-------|
| Load fridge ingredients | O(n) | O(n) | n = items in fridge |
| Create fridgeSet | O(n) | O(n) | HashMap creation |
| Check single ingredient | O(1) | - | HashMap lookup |
| Load all recipes | O(r) | O(r) | r = recipes in catalog |
| Match one recipe | O(r*i) | O(i) | i = ingredients in recipe |
| Match all recipes | O(r*i) | O(r*i) | r recipes, i avg ingredients |
| Sort recipes | O(r log r) | O(1) | In-place bubble sort |
| Total | **O(r*i + r log r)** | **O(n + r*i)** | Usually ~100ms |

---

## Возможные улучшения

1. **Caching:** Redis кеш для GetAllRecipes() (обновляется при добавлении рецепта)
2. **Pagination:** Вместо limit, добавить offset для больших списков
3. **Filters:** Добавить фильтры (по времени готовки, калориям, рейтингу)
4. **AI Suggestions:** Phase 2 - добавить AI блок с предложениями замен
5. **Substitution API:** Отдельный endpoint для замены ингредиентов
6. **Search:** Поиск рецептов по названию + фильтр по match_percent
7. **User preferences:** Сохранять любимые рецепты и ранжировать их выше

