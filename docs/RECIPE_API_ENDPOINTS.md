# 📖 API Эндпоинты каталога рецептов

## 🎯 Обзор

Модуль рецептов предоставляет **12 эндпоинтов** для работы с каталогом рецептов, матчинга с холодильником, AI-адаптации и приготовления блюд.

**База данных:** 70 рецептов в каталоге (recipes table)

---

## 📋 Список всех эндпоинтов

### 🌍 Публичные эндпоинты (без авторизации)

1. **GET /api/recipes/stats** - Статистика каталога
2. **GET /api/recipes** - Список рецептов с фильтрами
3. **GET /api/recipes/{id}** - Детали рецепта (опциональная авторизация)
4. **POST /api/recipes/{id}/view** - Инкремент просмотров

### 🔓 Временно публичные (для тестирования)

> ⚠️ **TODO:** Переместить в защищённую зону после тестирования

5. **GET /api/recipes/match** - Матчинг рецептов с холодильником
6. **GET /api/recipes/available** - Категоризация рецептов (canCook, almostCook, needToBuy)
7. **POST /api/recipes/recommendations** - Получить 1 рекомендацию для UI

### 🔒 Защищённые эндпоинты (требуется JWT)

8. **POST /api/user/recipes/save** - Сохранить рецепт в избранное
9. **GET /api/user/recipes/saved** - Получить сохранённые рецепты
10. **POST /api/recipes/{id}/cook** - Приготовить рецепт (списывает с холодильника)
11. **POST /api/recipes/{id}/adapt** - AI-адаптация рецепта

### 🗑️ Устаревшие эндпоинты (отключены)

12. **POST /api/recipes** - Создать рецепт пользователя (deprecated)
13. **PUT /api/recipes/{id}** - Обновить рецепт (deprecated)
14. **DELETE /api/recipes/{id}** - Удалить рецепт (deprecated)

---

## 📊 1. GET /api/recipes/stats

**Статистика каталога рецептов**

### Request
```http
GET /api/recipes/stats
```

### Response
```json
{
  "success": true,
  "data": {
    "totalRecipes": 70,
    "byCategory": {
      "breakfast": 12,
      "lunch": 25,
      "dinner": 20,
      "dessert": 8,
      "snack": 5
    }
  }
}
```

### Параметры
- Нет параметров

### Использование
```bash
curl https://menu-fodi-backend.koyeb.app/api/recipes/stats
```

---

## 📖 2. GET /api/recipes

**Список рецептов с фильтрами (каталог)**

### Request
```http
GET /api/recipes?country=russia&category=breakfast&difficulty=easy&maxTime=30&limit=20
```

### Query Parameters

| Параметр | Тип | Описание | Пример |
|----------|-----|----------|--------|
| `country` | string | Страна кухни | `russia`, `italy`, `japan` |
| `category` | string | Категория блюда | `breakfast`, `lunch`, `dinner`, `dessert` |
| `difficulty` | string | Сложность | `easy`, `medium`, `hard` |
| `maxTime` | int | Макс. время приготовления (мин) | `30`, `60` |
| `excludeAllergens` | string | Исключить аллергены (через запятую) | `gluten,dairy` |
| `dietTags` | string | Диетические теги (через запятую) | `vegetarian,vegan` |
| `limit` | int | Кол-во результатов (по умолчанию: 20) | `10`, `50` |

### Response
```json
{
  "success": true,
  "data": {
    "recipes": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "canonicalName": "scrambled_eggs",
        "localName": "Яичница-болтунья",
        "country": "russia",
        "category": "breakfast",
        "difficulty": "easy",
        "timeMinutes": 10,
        "servings": 2
      }
    ],
    "count": 1,
    "filters": {
      "country": "russia",
      "category": "breakfast",
      "difficulty": "easy",
      "maxTime": 30,
      "limit": 20
    }
  }
}
```

### Использование
```bash
# Все рецепты на завтрак
curl "https://menu-fodi-backend.koyeb.app/api/recipes?category=breakfast"

# Быстрые русские рецепты
curl "https://menu-fodi-backend.koyeb.app/api/recipes?country=russia&maxTime=20"

# Вегетарианские рецепты без глютена
curl "https://menu-fodi-backend.koyeb.app/api/recipes?dietTags=vegetarian&excludeAllergens=gluten"
```

---

## 🔍 3. GET /api/recipes/{id}

**Детали рецепта с ингредиентами и шагами**

### Request
```http
GET /api/recipes/550e8400-e29b-41d4-a716-446655440000
Authorization: Bearer <JWT_TOKEN> (опционально)
```

### Path Parameters
- `id` (UUID) - ID рецепта

### Response (без авторизации)
```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "canonicalName": "scrambled_eggs",
    "localName": "Яичница-болтунья",
    "country": "russia",
    "category": "breakfast",
    "difficulty": "easy",
    "timeMinutes": 10,
    "servings": 2,
    "ingredients": [
      {
        "id": "...",
        "name": "Яйца",
        "quantity": 4,
        "unit": "шт",
        "optional": false
      },
      {
        "id": "...",
        "name": "Молоко",
        "quantity": 50,
        "unit": "мл",
        "optional": true
      }
    ],
    "steps": [
      {
        "step": 1,
        "description": "Взбить яйца с молоком"
      },
      {
        "step": 2,
        "description": "Жарить на сковороде 5 минут"
      }
    ]
  }
}
```

### Response (с авторизацией)
```json
{
  "success": true,
  "data": {
    // ... все поля как выше, плюс:
    "ingredients": [
      {
        "id": "...",
        "name": "Яйца",
        "quantity": 4,
        "unit": "шт",
        "optional": false,
        "inFridge": true,        // ✅ Есть в холодильнике
        "availableQuantity": 6   // ✅ Сколько есть
      },
      {
        "id": "...",
        "name": "Молоко",
        "quantity": 50,
        "unit": "мл",
        "optional": true,
        "inFridge": false,       // ❌ Нет в холодильнике
        "availableQuantity": 0
      }
    ]
  }
}
```

### Особенности
- **Без JWT:** Базовая информация о рецепте
- **С JWT:** + флаги `inFridge` и `availableQuantity` для каждого ингредиента
- Используется `OptionalAuthMiddleware` (авторизация опциональна)

### Использование
```bash
# Без авторизации
curl https://menu-fodi-backend.koyeb.app/api/recipes/550e8400-e29b-41d4-a716-446655440000

# С авторизацией (для проверки что есть в холодильнике)
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  https://menu-fodi-backend.koyeb.app/api/recipes/550e8400-e29b-41d4-a716-446655440000
```

---

## 🎯 4. GET /api/recipes/match

**Матчинг рецептов с холодильником пользователя**

### Request
```http
GET /api/recipes/match?minScore=70&onlyCookable=true&category=breakfast&limit=10
Authorization: Bearer <JWT_TOKEN>
```

### Query Parameters

| Параметр | Тип | Описание | Значение по умолчанию |
|----------|-----|----------|----------------------|
| `country` | string | Страна кухни | - |
| `category` | string | Категория | - |
| `difficulty` | string | Сложность | - |
| `maxTime` | int | Макс. время (мин) | - |
| `excludeAllergens` | string | Аллергены (через запятую) | - |
| `dietTags` | string | Диетические теги | - |
| `minScore` | float | Мин. % совпадения (0-100) | `0` |
| `onlyCookable` | bool | Только готовые к приготовлению | `false` |
| `limit` | int | Кол-во результатов | `10` |
| `testUserID` | string | ⚠️ DEV MODE: test user ID | - |

### Response
```json
{
  "success": true,
  "data": {
    "recipes": [
      {
        "recipeId": "550e8400-e29b-41d4-a716-446655440000",
        "canonicalName": "scrambled_eggs",
        "localName": "Яичница-болтунья",
        "category": "breakfast",
        "difficulty": "easy",
        "timeMinutes": 10,
        "servings": 2,
        "match": 85,                    // % совпадения с холодильником
        "canCook": true,                // Можно готовить прямо сейчас
        "missingCount": 1,              // Сколько ингредиентов не хватает
        "missing": ["Молоко"],          // Какие ингредиенты нужно докупить
        "costToComplete": 65.50,        // Стоимость докупки (руб)
        "wasteRiskSaved": 120.00,       // Экономия от использования продуктов
        "hasExpiringItems": true,       // Есть продукты с истекающим сроком
        "ingredients": [
          {
            "name": "Яйца",
            "required": 4,
            "available": 6,
            "unit": "шт",
            "optional": false,
            "inFridge": true
          },
          {
            "name": "Молоко",
            "required": 50,
            "available": 0,
            "unit": "мл",
            "optional": true,
            "inFridge": false
          }
        ]
      }
    ],
    "count": 1
  }
}
```

### Логика матчинга

```
Match Score (%) = (Available Ingredients / Total Required Ingredients) × 100

CanCook = (All required non-optional ingredients are available)

WasteRiskSaved = Sum(price × quantity) для ингредиентов близких к истечению срока
```

### Использование
```bash
# Все рецепты, которые можно приготовить прямо сейчас
curl -H "Authorization: Bearer $JWT_TOKEN" \
  "https://menu-fodi-backend.koyeb.app/api/recipes/match?onlyCookable=true"

# Рецепты с совпадением >= 70%
curl -H "Authorization: Bearer $JWT_TOKEN" \
  "https://menu-fodi-backend.koyeb.app/api/recipes/match?minScore=70&limit=5"

# Быстрые завтраки с высоким матчингом
curl -H "Authorization: Bearer $JWT_TOKEN" \
  "https://menu-fodi-backend.koyeb.app/api/recipes/match?category=breakfast&maxTime=15&minScore=80"
```

---

## 📊 5. GET /api/recipes/available

**Категоризация рецептов по возможности приготовления**

### Request
```http
GET /api/recipes/available
Authorization: Bearer <JWT_TOKEN>
```

### Response
```json
{
  "success": true,
  "data": {
    "canCook": [
      {
        "recipeId": "...",
        "canonicalName": "scrambled_eggs",
        "localName": "Яичница-болтунья",
        "category": "breakfast",
        "difficulty": "easy",
        "timeMinutes": 10,
        "servings": 2,
        "match": 100,
        "canCook": true,
        "missingCount": 0,
        "missing": [],
        "costToComplete": 0,
        "wasteRiskSaved": 80.00,
        "hasExpiringItems": true
      }
    ],
    "almostCook": [
      {
        "recipeId": "...",
        "localName": "Паста Карбонара",
        "match": 75,
        "canCook": false,
        "missingCount": 1,
        "missing": ["Бекон"],
        "costToComplete": 150.00
      }
    ],
    "needToBuy": [
      {
        "recipeId": "...",
        "localName": "Борщ",
        "match": 30,
        "canCook": false,
        "missingCount": 5,
        "missing": ["Свекла", "Капуста", "Картофель", "Морковь", "Лук"],
        "costToComplete": 320.00
      }
    ],
    "counts": {
      "canCook": 5,
      "almostCook": 12,
      "needToBuy": 53
    }
  }
}
```

### Категории

| Категория | Условие | Описание |
|-----------|---------|----------|
| **canCook** | `match >= 90` AND `canCook = true` | Можно готовить прямо сейчас |
| **almostCook** | `50 <= match < 90` | Нужно докупить 1-2 ингредиента |
| **needToBuy** | `match < 50` | Нужно много докупать |

### Использование
```bash
curl -H "Authorization: Bearer $JWT_TOKEN" \
  https://menu-fodi-backend.koyeb.app/api/recipes/available
```

---

## 🎲 6. POST /api/recipes/recommendations

**Получить 1 рекомендацию для UI (карточка "что приготовить")**

### Request
```http
POST /api/recipes/recommendations
Authorization: Bearer <JWT_TOKEN>
Content-Type: application/json

{
  "mode": "fridge",
  "excludeRecipeIds": ["550e8400-e29b-41d4-a716-446655440000"],
  "limit": 5
}
```

### Request Body

| Поле | Тип | Обязательно | Описание |
|------|-----|-------------|----------|
| `mode` | string | ✅ | Режим рекомендаций (только `"fridge"`) |
| `excludeRecipeIds` | string[] | ❌ | ID рецептов для исключения |
| `limit` | int | ❌ | Кол-во рекомендаций (по умолчанию: 5) |

### Response (успех)
```json
{
  "success": true,
  "data": {
    "recipe": {
      "recipeId": "...",
      "canonicalName": "scrambled_eggs",
      "localName": "Яичница-болтунья",
      "category": "breakfast",
      "difficulty": "easy",
      "timeMinutes": 10,
      "servings": 2,
      "match": 100,
      "canCook": true,
      "missingCount": 0,
      "costToComplete": 0,
      "wasteRiskSaved": 120.00,
      "hasExpiringItems": true
    }
  }
}
```

### Response (нет рецептов)
```json
{
  "success": true,
  "code": "NO_MORE_RECIPES",
  "context": {
    "totalExcluded": 15,
    "totalInCatalog": 70
  }
}
```

### Логика работы

1. **Получает saved recipes** - исключает уже сохранённые
2. **Получает session exclusions** - исключает показанные ранее в сессии
3. **Объединяет с request exclusions** - исключает переданные в запросе
4. **Находит лучший рецепт** - с максимальным `matchScore`
5. **Обновляет session** - добавляет выбранный рецепт в exclusions
6. **Возвращает 1 рецепт** - для отображения карточки

### Использование
```bash
# Получить первую рекомендацию
curl -X POST \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"mode":"fridge","limit":5}' \
  https://menu-fodi-backend.koyeb.app/api/recipes/recommendations

# Исключить рецепт и получить следующий
curl -X POST \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"mode":"fridge","excludeRecipeIds":["550e8400-..."]}' \
  https://menu-fodi-backend.koyeb.app/api/recipes/recommendations
```

---

## 💾 7. POST /api/user/recipes/save

**Сохранить рецепт в избранное**

### Request
```http
POST /api/user/recipes/save
Authorization: Bearer <JWT_TOKEN>
Content-Type: application/json

{
  "recipeId": "550e8400-e29b-41d4-a716-446655440000"
}
```

### Response
```json
{
  "success": true,
  "data": {
    "savedRecipeId": "660e8400-...",
    "recipeId": "550e8400-...",
    "userId": "user-uuid",
    "savedAt": "2026-01-04T15:30:00Z"
  }
}
```

### Особенности
- Предотвращает дубликаты (один рецепт = 1 запись в saved)
- Используется для исключения из recommendations
- Можно сохранять до приготовления

---

## 📚 8. GET /api/user/recipes/saved

**Получить список сохранённых рецептов**

### Request
```http
GET /api/user/recipes/saved
Authorization: Bearer <JWT_TOKEN>
```

### Response
```json
{
  "success": true,
  "data": {
    "recipes": [
      {
        "savedRecipeId": "660e8400-...",
        "recipeId": "550e8400-...",
        "canonicalName": "scrambled_eggs",
        "localName": "Яичница-болтунья",
        "category": "breakfast",
        "difficulty": "easy",
        "timeMinutes": 10,
        "servings": 2,
        "savedAt": "2026-01-04T15:30:00Z",
        "cookedAt": null,
        "cookedCount": 0
      }
    ],
    "count": 1
  }
}
```

---

## 🍳 9. POST /api/recipes/{id}/cook

**Приготовить рецепт (списывает ингредиенты с холодильника)**

### Request
```http
POST /api/recipes/550e8400-e29b-41d4-a716-446655440000/cook
Authorization: Bearer <JWT_TOKEN>
Content-Type: application/json

{
  "servings": 2,
  "idempotencyKey": "cook-20260104-153000"
}
```

### Request Body

| Поле | Тип | Обязательно | Описание |
|------|-----|-------------|----------|
| `servings` | int | ❌ | Кол-во порций (multiplier) |
| `idempotencyKey` | string | ❌ | Ключ для предотвращения дубликатов |

### Response (успех)
```json
{
  "success": true,
  "data": {
    "cookLogId": "770e8400-...",
    "recipeId": "550e8400-...",
    "recipeName": "Яичница-болтунья",
    "userId": "user-uuid",
    "servings": 2,
    "usedValue": 85.00,
    "wasteRiskSaved": 40.00,
    "cookedAt": "2026-01-04T15:30:00Z",
    "remainingItems": 12,
    "deductedIngredients": [
      {
        "ingredientId": "...",
        "name": "Яйца",
        "quantity": 4,
        "unit": "шт"
      }
    ]
  }
}
```

### Response (недостаточно ингредиентов)
```json
{
  "success": false,
  "code": "INSUFFICIENT_INGREDIENTS",
  "message": "Not enough ingredients to cook this recipe",
  "missing": [
    {
      "ingredientId": "...",
      "name": "Молоко",
      "required": 200,
      "available": 50,
      "unit": "мл"
    }
  ]
}
```

### Что происходит при готовке?

1. **Проверка ингредиентов** - все обязательные в наличии?
2. **Расчёт multiplier** - если servings указаны
3. **Списание с холодильника** - уменьшает quantity в user_fridge_items
4. **Создание cook log** - запись в recipe_cook_log
5. **Обновление prepared_dishes** - добавляет готовое блюдо
6. **History event** - логирует событие "recipe_cooked"
7. **Расчёт экономии** - wasteRiskSaved для продуктов близких к истечению

### Использование
```bash
# Приготовить стандартную порцию
curl -X POST \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  https://menu-fodi-backend.koyeb.app/api/recipes/550e8400-.../cook

# Приготовить двойную порцию
curl -X POST \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"servings":4}' \
  https://menu-fodi-backend.koyeb.app/api/recipes/550e8400-.../cook
```

---

## 🤖 10. POST /api/recipes/{id}/adapt

**AI-адаптация рецепта под доступные ингредиенты**

### Request
```http
POST /api/recipes/550e8400-e29b-41d4-a716-446655440000/adapt
Authorization: Bearer <JWT_TOKEN>
Content-Type: application/json

{
  "missingIngredients": ["молоко"],
  "availableAlternatives": ["сливки", "кефир"]
}
```

### Response
```json
{
  "success": true,
  "data": {
    "adaptedRecipe": {
      "originalRecipeId": "550e8400-...",
      "originalName": "Яичница-болтунья",
      "adaptedName": "Яичница со сливками",
      "adaptations": [
        {
          "originalIngredient": "Молоко",
          "originalQuantity": 50,
          "originalUnit": "мл",
          "replacedWith": "Сливки",
          "newQuantity": 30,
          "newUnit": "мл",
          "reason": "Сливки более жирные, нужно меньше"
        }
      ],
      "modifiedSteps": [
        {
          "step": 1,
          "original": "Взбить яйца с молоком",
          "modified": "Взбить яйца со сливками"
        }
      ]
    }
  }
}
```

### Особенности
- Использует GROQ API (LLM)
- Учитывает доступные альтернативы
- Адаптирует шаги приготовления
- Сохраняет оригинальный рецепт без изменений

---

## 👀 11. POST /api/recipes/{id}/view

**Инкремент счётчика просмотров рецепта**

### Request
```http
POST /api/recipes/550e8400-e29b-41d4-a716-446655440000/view
```

### Response
```json
{
  "success": true,
  "data": {
    "recipeId": "550e8400-...",
    "viewsCount": 142
  }
}
```

### Особенности
- Публичный эндпоинт (без авторизации)
- Используется для аналитики популярности
- Инкрементирует поле `views_count` в таблице recipes

---

## 🔧 Вспомогательные функции

### Парсинг query параметров

```go
// Целое число с дефолтом
parseIntQuery(r, "limit", 20)

// Число с плавающей точкой
parseFloatQuery(r, "minScore", 0.0)

// Булево значение
parseBoolQuery(r, "onlyCookable", false)

// Массив через запятую
parseArrayQuery(r, "excludeAllergens")
// ?excludeAllergens=gluten,dairy → ["gluten", "dairy"]
```

### Локализация имён

```go
// Получить имя на языке пользователя
recipe.GetLocalizedName(userLang)

// Поддерживаются: "ru", "en", "pl"
// Fallback: canonicalName → localName → first available
```

---

## 📊 Статистика использования

```
📖 Каталог:       70 рецептов
🔍 Матчинг:       Средний match score: 65%
🍳 Приготовлено:  1,234 рецепта
💾 Сохранено:     567 рецептов в избранном
🤖 AI-адаптации:  89 адаптаций
```

---

## 🚀 Примеры интеграции

### Frontend (TypeScript)

```typescript
// Получить рекомендацию
const getRecommendation = async () => {
  const response = await fetch('/api/recipes/recommendations', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      mode: 'fridge',
      limit: 5
    })
  });
  
  const data = await response.json();
  
  if (data.code === 'NO_MORE_RECIPES') {
    console.log('Все рецепты показаны!');
    return null;
  }
  
  return data.data.recipe;
};

// Приготовить рецепт
const cookRecipe = async (recipeId: string, servings: number) => {
  const response = await fetch(`/api/recipes/${recipeId}/cook`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ servings })
  });
  
  const data = await response.json();
  
  if (!data.success && data.code === 'INSUFFICIENT_INGREDIENTS') {
    console.error('Недостаточно ингредиентов:', data.missing);
    return null;
  }
  
  return data.data;
};
```

---

## 🔐 Middleware Chain

```
Request
  ↓
[Chi Router]
  ↓
[AuthMiddleware] (для защищённых эндпоинтов)
  ↓
[OptionalAuthMiddleware] (для /recipes/{id})
  ↓
[RecipeHandler]
  ↓
[RecipeMatchService / RecipeCookService]
  ↓
[Database (GORM)]
  ↓
Response
```

---

## 📚 Связанная документация

- **RECIPE_SYSTEM_SUMMARY.md** - Обзор системы рецептов
- **MATCH_API_CONTRACT.md** - Детали API матчинга
- **COOK_API_CONTRACT.md** - Детали API приготовления
- **RECIPE_CATALOG_QUICK_REF.md** - Быстрый справочник
- **AI_RECIPE_ADAPTATION.md** - AI-адаптация рецептов

---

**Последнее обновление:** 4 января 2026 г.  
**Версия API:** v1
