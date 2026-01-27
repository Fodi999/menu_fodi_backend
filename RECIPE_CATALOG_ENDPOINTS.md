# 📚 Эндпоинты для отображения рецептов

## 📋 Обзор всех доступных endpoints

Система имеет **3 основных API** для работы с рецептами:

| API | Назначение | Требует auth |
|-----|-----------|-------------|
| **Recipe Catalog API** | Просмотр каталога, фильтрация | ❌ Нет |
| **Recipe Recommendations** | Подбор рецептов по холодильнику | ✅ Да |
| **User Saved Recipes** | Сохраненные рецепты пользователя | ✅ Да |

---

## 🍳 API 1: Recipe Catalog (Основной каталог)

### 1.1 Получить все рецепты с фильтрацией

```
GET /api/recipes?category=soup&maxTime=60&difficulty=easy&limit=20
Authorization: (не требуется)
```

**Query params (все опциональны):**
```
?country=Poland            # Страна происхождения
?category=soup             # Категория (soup, pasta, salad, main, dessert)
?difficulty=easy           # Сложность (easy, medium, hard)
?maxTime=60                # Максимальное время готовки (минуты)
?dietTags=vegetarian       # Диетические ограничения (vegetarian, vegan, gluten_free)
?excludeAllergens=nuts     # Исключить аллергены (nuts, dairy, gluten, etc.)
?limit=20                  # Рецептов на странице (default: 20, max: 100)
?lang=ru                   # Язык (pl, en, ru) - optional
```

**Response (200 OK):**
```json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "title": "Борщ украинский",
    "canonical_name": "borsch",
    "image_url": "https://res.cloudinary.com/...",
    "cook_time": 60,
    "servings": 4,
    "category": "soup",
    "difficulty": "easy",
    "country": "Ukraine",
    "ingredients": [
      {
        "id": "ing-1",
        "canonical_name": "beet",
        "display_name": "Свекла",
        "quantity": 500,
        "unit": "g",
        "category": "vegetable"
      }
    ],
    "steps": [
      "Нарежьте свеклу кубиками",
      "Варите 30 минут"
    ]
  },
  {
    "id": "660e8400-e29b-41d4-a716-446655440001",
    "title": "Паста Карбонара",
    "canonical_name": "pasta_carbonara",
    "image_url": "...",
    "cook_time": 30,
    "servings": 2,
    "category": "pasta",
    "difficulty": "medium",
    "country": "Italy",
    "ingredients": [...],
    "steps": [...]
  }
]
```

**Коды ответов:**
- `200` — Успешно
- `500` — Ошибка сервера

**Примеры фильтрации:**
```bash
# Все супы до 60 минут
GET /api/recipes?category=soup&maxTime=60

# Вегетарианские рецепты
GET /api/recipes?dietTags=vegetarian

# Паста, исключая молочные продукты
GET /api/recipes?category=pasta&excludeAllergens=dairy

# Легкие рецепты из Поль
GET /api/recipes?difficulty=easy&country=Poland
```

---

### 1.2 Статистика по рецептам

```
GET /api/recipes/stats
Authorization: (не требуется)
```

**Response (200 OK):**
```json
{
  "total_recipes": 127,
  "by_category": {
    "soup": 15,
    "pasta": 22,
    "salad": 18,
    "main": 45,
    "dessert": 27
  },
  "by_difficulty": {
    "easy": 50,
    "medium": 55,
    "hard": 22
  },
  "by_cook_time": {
    "quick": 35,      // до 30 минут
    "medium": 60,     // 30-60 минут
    "long": 32        // более 60 минут
  }
}
```

---

### 1.3 Получить один рецепт из каталога

```
GET /api/recipes/{id}?lang=ru
Authorization: (не требуется, но optional для добавления inFridge)
```

**Path params:**
- `id` — UUID рецепта

**Query params:**
- `lang=ru` — Язык (pl, en, ru)

**Response (200 OK):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "title": "Борщ украинский",
  "canonical_name": "borsch",
  "image_url": "https://res.cloudinary.com/...",
  "cook_time": 60,
  "servings": 4,
  "category": "soup",
  "difficulty": "easy",
  "rating": 4.5,
  "nutrition": {
    "calories": 150,
    "protein": 8,
    "fat": 5,
    "carbs": 20
  },
  "ingredients": [
    {
      "id": "ing-1",
      "name": "Свекла",
      "quantity": 500,
      "unit": "g",
      "category": "vegetable",
      "in_fridge": false  // ⭐ Только если пользователь авторизован
    },
    {
      "id": "ing-2",
      "name": "Масло растительное",
      "quantity": 100,
      "unit": "ml",
      "category": "oil",
      "in_fridge": true   // ⭐ Из холодильника пользователя
    }
  ],
  "steps": [
    "Нарежьте свеклу кубиками",
    "Разогрейте масло на сковороде",
    "Добавьте свеклу и варите 30 минут"
  ]
}
```

**Коды ответов:**
- `200` — Успешно
- `404` — Рецепт не найден

---

## 🎯 API 2: Recipe Recommendations (Подбор по холодильнику)

### 2.1 Получить рекомендации рецептов

```
GET /api/recipe-recommendations?lang=ru&limit=10
Authorization: Bearer <JWT>
```

**Query params:**
- `lang=ru` — Язык (pl, en, ru) [default: pl]
- `limit=10` — Количество рецептов [default: 10]

**Response (200 OK):**
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
          "display_name": "Масло растительное",
          "quantity": 100,
          "unit": "ml"
        },
        {
          "id": "id-salt-1",
          "display_name": "Соль",
          "quantity": 5,
          "unit": "g"
        }
      ],
      "missing_ingredients": [
        {
          "id": "id-beet-1",
          "display_name": "Свекла",
          "quantity": 500,
          "unit": "g"
        }
      ],
      "steps": [
        "Разогрейте масло в кастрюле",
        "Нарежьте свеклу кубиками"
      ]
    }
  ]
}
```

**Коды ответов:**
- `200` — Успешно
- `401` — Не авторизован
- `400` — Холодильник пуст

---

### 2.2 Получить один рецепт с проверкой холодильника

```
GET /api/recipe-recommendations/{id}?lang=ru
Authorization: Bearer <JWT>
```

**Response:** Один RecipeDTO (структура как выше)

---

## 👤 API 3: User Saved Recipes (Сохраненные рецепты)

### 3.1 Сохранить рецепт

```
POST /api/user/recipes/save
Authorization: Bearer <JWT>
Content-Type: application/json
```

**Request body:**
```json
{
  "recipe_id": "550e8400-e29b-41d4-a716-446655440000",
  "category": "favorites"  // или "to_try", "favorite", etc.
}
```

**Response (201 Created):**
```json
{
  "id": "saved-1",
  "recipe_id": "550e8400-e29b-41d4-a716-446655440000",
  "user_id": "user-123",
  "saved_at": "2025-01-27T10:30:00Z",
  "category": "favorites"
}
```

---

### 3.2 Получить сохраненные рецепты пользователя

```
GET /api/user/recipes/saved?category=favorites&limit=10
Authorization: Bearer <JWT>
```

**Query params:**
- `category=favorites` — Фильтр по категории (optional)
- `limit=10` — Количество (default: 20)
- `page=1` — Пагинация (default: 1)
- `lang=ru` — Язык (pl, en, ru)

**Response (200 OK):**
```json
{
  "recipes": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "title": "Борщ украинский",
      "canonical_name": "borsch",
      "image_url": "...",
      "cook_time": 60,
      "servings": 4,
      "saved_at": "2025-01-20T14:30:00Z",
      "category": "favorites"
    }
  ],
  "total": 5,
  "page": 1,
  "limit": 10
}
```

---

## 🔄 API 4: Recipe Matching & Cooking

### 4.1 Получить доступные рецепты

```
GET /api/recipes/available?lang=ru&limit=20
Authorization: Bearer <JWT>
```

**Response (200 OK):**
```json
{
  "ready": [
    {
      "id": "550e8400-...",
      "title": "Паста Карбонара",
      "match_percent": 100,
      "cook_time": 30
    }
  ],
  "almost_ready": [
    {
      "id": "660e8400-...",
      "title": "Борщ",
      "match_percent": 75,
      "missing_count": 2
    }
  ],
  "summary": "Готово: 1, Почти готово: 3, Нужно купить: 12"
}
```

---

### 4.2 Готовить рецепт (уменьшить кол-во в холодильнике)

```
POST /api/recipes/{id}/cook
Authorization: Bearer <JWT>
Content-Type: application/json
```

**Request body:**
```json
{
  "servings": 2         // Количество порций (по умолчанию как в рецепте)
}
```

**Response (200 OK):**
```json
{
  "message": "Recipe cooked successfully",
  "fridge_updated": {
    "beef": -300,       // Уменьшено на 300g
    "salt": -5,         // Уменьшено на 5g
    "oil": -50          // Уменьшено на 50ml
  }
}
```

---

### 4.3 Адаптировать рецепт (AI)

```
POST /api/recipes/{id}/adapt
Authorization: Bearer <JWT>
Content-Type: application/json
```

**Request body:**
```json
{
  "available_ingredients": ["масло", "соль", "перец"],
  "exclude_ingredients": ["говядина"],
  "dietary_restrictions": ["vegetarian"]
}
```

**Response (200 OK):**
```json
{
  "adapted_recipe": {
    "title": "Борщ вегетарианский",
    "ingredients": [
      {
        "name": "Свекла",
        "quantity": 500,
        "unit": "g"
      }
    ],
    "steps": [
      "Нарежьте свеклу",
      "Варите в воде 30 минут"
    ],
    "ai_note": "Заменили говядину на чечевицу для белка"
  }
}
```

---

## 🏛️ Legacy API (для совместимости)

### GET /api/ai-recipe/recommendation

Старый эндпоинт, оставлен для совместимости. Рекомендуется использовать `/api/recipe-recommendations` вместо этого.

---

## 📝 Полная таблица всех эндпоинтов

| Method | Endpoint | Auth | Описание | Return |
|--------|----------|------|---------|--------|
| GET | `/api/recipes` | ❌ | Все рецепты с фильтрацией | 200 или 500 |
| GET | `/api/recipes/stats` | ❌ | Статистика по рецептам | 200 |
| GET | `/api/recipes/{id}` | 🔶 | Один рецепт (опц. с in_fridge) | 200 или 404 |
| GET | `/api/recipe-recommendations` | ✅ | Рецепты по холодильнику | 200 или 401 |
| GET | `/api/recipe-recommendations/{id}` | ✅ | Один рецепт с проверкой холодильника | 200 или 401 |
| POST | `/api/user/recipes/save` | ✅ | Сохранить рецепт | 201 или 401 |
| GET | `/api/user/recipes/saved` | ✅ | Сохраненные рецепты | 200 или 401 |
| GET | `/api/recipes/available` | ✅ | Доступные рецепты по статусам | 200 или 401 |
| POST | `/api/recipes/{id}/cook` | ✅ | Готовить рецепт | 200 или 400 |
| POST | `/api/recipes/{id}/adapt` | ✅ | Адаптировать рецепт (AI) | 200 или 400 |

**Legend:**
- ❌ = Не требуется авторизация (публичный)
- 🔶 = Опционально (лучше с авторизацией)
- ✅ = Требуется авторизация (JWT)

---

## 🔍 Поиск по названию

⚠️ **Поиск по названию рецепта** пока **не реализован** в GET /api/recipes.

Есть два способа найти рецепт по названию:

### Способ 1: По canonical_name (SEO URL)

```
GET /api/recipes/{canonical_name}
```

Примеры:
- GET /api/recipes/borsch
- GET /api/recipes/pasta_carbonara
- GET /api/recipes/chicken_soup

Это работает благодаря тому, что handler `GetRecipeByID` сначала пробует UUID, а потом ищет по `canonical_name`.

### Способ 2: Через фильтры

Текущие фильтры позволяют найти рецепты по:
- **category** (soup, pasta, salad, main, dessert)
- **difficulty** (easy, medium, hard)
- **maxTime** (максимум минут готовки)
- **dietTags** (vegetarian, vegan, gluten_free)
- **excludeAllergens** (nuts, dairy, gluten, etc.)

Пример: "Найти все вегетарианские супы до 45 минут":
```
GET /api/recipes?category=soup&maxTime=45&dietTags=vegetarian
```

---

## 🔐 Авторизация

Все защищенные эндпоинты требуют:

```http
Authorization: Bearer <JWT_TOKEN>
```

JWT должен содержать:
```json
{
  "sub": "user-id",
  "email": "user@example.com",
  "role": "home_chef",
  "exp": 1704067200
}
```

---

## 🌍 Поддерживаемые языки

Все эндпоинты поддерживают параметр `lang`:

```
lang=pl  # Польский (default)
lang=en  # Английский
lang=ru  # Русский
```

---

## 💡 Примеры использования

### Пример 1: Просмотр каталога с фильтрацией

```bash
# Все рецепты
curl "https://api.example.com/api/recipes"

# Только супы (category=soup)
curl "https://api.example.com/api/recipes?category=soup"

# Супы до 60 минут
curl "https://api.example.com/api/recipes?category=soup&maxTime=60"

# Вегетарианские супы
curl "https://api.example.com/api/recipes?category=soup&dietTags=vegetarian"

# Быстрые рецепты (до 30 мин), без молочных продуктов
curl "https://api.example.com/api/recipes?maxTime=30&excludeAllergens=dairy&lang=ru"

# Response:
# [
#   {
#     "id": "550e8400-...",
#     "title": "Борщ",
#     "cook_time": 60,
#     "difficulty": "easy",
#     "ingredients": [...],
#     "steps": [...]
#   }
# ]
```

---

### Пример 2: Получить рекомендации

```bash
curl -H "Authorization: Bearer TOKEN" \
  "https://api.example.com/api/recipe-recommendations?lang=ru&limit=5"

# Response:
# {
#   "decision": "ready",
#   "summary": "Отлично! Вы можете приготовить несколько рецептов",
#   "recipes": [...]
# }
```

---

### Пример 3: Сохранить любимый рецепт

```bash
curl -X POST \
  -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"recipe_id": "550e8400-e29b-41d4-a716-446655440000"}' \
  "https://api.example.com/api/user/recipes/save"

# Response (201):
# {
#   "id": "saved-1",
#   "recipe_id": "550e8400-...",
#   "saved_at": "2025-01-27T10:30:00Z"
# }
```

---

### Пример 4: Получить одно блюдо с деталями

```bash
curl "https://api.example.com/api/recipes/550e8400-e29b-41d4-a716-446655440000?lang=ru"

# Response:
# {
#   "id": "550e8400-...",
#   "title": "Борщ украинский",
#   "cook_time": 60,
#   "ingredients": [
#     { "name": "Свекла", "quantity": 500, "unit": "g", "in_fridge": false },
#     { "name": "Масло", "quantity": 100, "unit": "ml", "in_fridge": true }
#   ],
#   "steps": [...]
# }
```

---

## 🚀 Frontend Integration Patterns

### Pattern 1: Просмотр каталога с фильтрами

```javascript
// Загрузить супы до 60 минут
async function loadFilteredRecipes() {
  const response = await fetch(
    '/api/recipes?category=soup&maxTime=60&lang=ru'
  );
  const recipes = await response.json();
  
  // recipes это массив [...]
  recipes.forEach(recipe => {
    console.log(`${recipe.title} (${recipe.cook_time} мин)`);
  });
}

// Или с параметрами
async function loadWithFilters(filters) {
  const params = new URLSearchParams(filters);
  const response = await fetch(`/api/recipes?${params.toString()}`);
  const recipes = await response.json();
  return recipes;
}

// Использование
loadWithFilters({
  category: 'pasta',
  maxTime: 45,
  dietTags: 'vegetarian',
  lang: 'ru'
});
```

---

### Pattern 2: Умная рекомендация (если пользователь авторизован)

```javascript
async function getSmartRecommendations() {
  const token = localStorage.getItem('token');
  
  if (!token) {
    // Не авторизован - показать просто каталог
    return await fetch('/api/recipes?limit=10').then(r => r.json());
  }
  
  // Авторизован - получить рекомендации по холодильнику
  return await fetch('/api/recipe-recommendations?limit=10', {
    headers: { 'Authorization': `Bearer ${token}` }
  }).then(r => r.json());
}
```

---

### Pattern 3: Детальный просмотр с in_fridge флагами

```javascript
async function viewRecipeDetail(recipeId) {
  const token = localStorage.getItem('token');
  const headers = token ? { 'Authorization': `Bearer ${token}` } : {};
  
  const recipe = await fetch(`/api/recipes/${recipeId}?lang=ru`, { headers })
    .then(r => r.json());
  
  // Отобразить ингредиенты с цветной подсветкой
  recipe.ingredients.forEach(ing => {
    const status = ing.in_fridge ? '✅' : '❌';
    console.log(`${status} ${ing.name} (${ing.quantity} ${ing.unit})`);
  });
}
```

---

## ⚠️ Важные замечания

1. **Пагинация:** По умолчанию 20 рецептов на странице, макс 100
2. **Язык:** Если язык не указан, использует язык профиля пользователя (или "pl" по умолчанию)
3. **Auth:** Многие эндпоинты работают без авторизации, но с авторизацией предоставляют больше информации (in_fridge)
4. **Rate limiting:** На публичные эндпоинты может быть лимит на количество запросов
5. **CORS:** API доступен с https://dima-fomin.pl и http://localhost:3000

---

## 🎯 Какой эндпоинт использовать для разных сценариев?

| Сценарий | Эндпоинт | Фильтры |
|----------|----------|--------|
| Показать все рецепты (публичный каталог) | GET /api/recipes | - |
| Фильтр по категориям | GET /api/recipes?category=soup | category, difficulty, maxTime |
| Вегетарианские/веган рецепты | GET /api/recipes?dietTags=vegetarian | dietTags, excludeAllergens |
| Рецепты без аллергенов | GET /api/recipes?excludeAllergens=nuts | excludeAllergens |
| Быстрые рецепты (до 30 мин) | GET /api/recipes?maxTime=30 | maxTime, category |
| Статистика по каталогу | GET /api/recipes/stats | - |
| Детали одного рецепта | GET /api/recipes/{id} | lang (опционально) |
| Рецепты по холодильнику (авторизованные) | GET /api/recipe-recommendations | lang, limit |
| Только готовые рецепты | GET /api/recipes/available | lang, limit |
| Сохраненные рецепты | GET /api/user/recipes/saved | category, lang, limit |
| Сохранить рецепт | POST /api/user/recipes/save | - |
| Готовить рецепт (убрать из холодильника) | POST /api/recipes/{id}/cook | servings |
| Адаптировать рецепт под ингредиенты | POST /api/recipes/{id}/adapt | - |

