# Recipe Match API - Правильный Endpoint

## ✅ Правильный Endpoint для UI

```
GET /api/recipes/match
```

**НЕ используйте:**
- `/api/recipes/recommendations` (это для другого сценария)
- `/api/recipes/recommend` (не существует)

---

## Query Parameters

```typescript
interface MatchParams {
  onlyCookable?: boolean;  // true = только рецепты которые можно приготовить СЕЙЧАС
  minScore?: number;       // 100 = только полное покрытие (coverage=100%)
  limit?: number;          // Количество результатов (default: 10)
  country?: string;        // Фильтр по стране
  category?: string;       // main, dessert, soup, etc.
  difficulty?: string;     // easy, medium, hard
  maxTime?: number;        // Максимальное время в минутах
}
```

---

## Happy Path Example

### Request:
```bash
GET /api/recipes/match?onlyCookable=true&minScore=100&limit=5
Authorization: Bearer <token>
```

### Response:
```json
{
  "success": true,
  "data": {
    "recipes": [
      {
        "recipeId": "859d8c56-338e-4da0-8e5c-9ef5412b22ab",
        "localName": "яичница",
        "canonicalName": "яичница",
        "country": "ru",
        "category": "main",
        "difficulty": "easy",
        "timeMinutes": 10,
        "servings": 1,
        
        // Match results
        "score": 100,
        "coverage": 1.0,
        "canCookNow": true,
        
        // Ingredients
        "usedIngredients": [
          {
            "ingredientId": "37bf235a-5023-4e7a-915a-ef31c1cd3cd0",
            "name": "fresh eggs",
            "quantity": 2,
            "unit": "шт",
            "available": 2,
            "isExpiringSoon": false
          },
          {
            "ingredientId": "1b7cea8e-b026-4329-9d2e-c94952e3fa6c",
            "name": "Vegetable oil",
            "quantity": 10,
            "unit": "мл",
            "available": 20
          },
          {
            "ingredientId": "008bd5a9-720d-457c-b46e-2d7871932db4",
            "name": "Rock salt",
            "quantity": 1,
            "unit": "г",
            "available": 5
          }
        ],
        "missingIngredients": [],
        
        // Economy
        "costToComplete": 0,
        "usedValue": 0.1,
        "savedMoney": 0.1,
        "totalRecipeCost": 0.1,
        
        // Priority
        "hasExpiringItems": false,
        "expiringItemsCount": 0
      }
    ],
    "count": 3
  }
}
```

---

## Фильтрация

### Backend делает:

✅ **Фильтрует по `onlyCookable=true`:**
```go
if filters.OnlyCookable && !match.CanMakeNow {
    continue  // Пропускаем рецепты которые нельзя приготовить сейчас
}
```

✅ **Фильтрует по `minScore`:**
```go
if match.MatchScore < filters.MinScore {
    continue  // Пропускаем рецепты с низким score
}
```

✅ **НЕ фильтрует по `source.type`:**
- Рецепты с `type="professional"` и `type="ai"` могут быть в ответе
- Важен только `canMakeNow` и `coverage`

---

## Sorting

Рецепты сортируются по:
1. `canCookNow DESC` - сначала те что можно готовить
2. `score DESC` - потом по score
3. `costToComplete ASC` - дешёвые в дополнении
4. `timeMinutes ASC` - быстрые в приготовлении

---

## Testing

```bash
# Successful test
make test-egg

# Manual test
USER_TOKEN=$(curl -s -X POST "https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"fodi85@gmail.ru","password":"210185"}' | jq -r '.data.token')

curl -s "https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/recipes/match?onlyCookable=true&minScore=100&limit=5" \
  -H "Authorization: Bearer $USER_TOKEN" | jq .
```

---

## TypeScript Interface

```typescript
interface RecipeMatchResponse {
  success: boolean;
  data: {
    recipes: RecipeMatchItem[];
    count: number;
  };
}

interface RecipeMatchItem {
  recipeId: string;
  localName: string;
  canonicalName: string;
  country: string;
  category: string;
  difficulty: string;
  timeMinutes: number;
  servings: number;
  
  // Match
  score: number;         // 0-100
  coverage: number;      // 0-1
  canCookNow: boolean;
  
  // Ingredients
  usedIngredients: IngredientMatch[];
  missingIngredients: IngredientMatch[];
  
  // Economy
  costToComplete: number;
  usedValue: number;
  savedMoney: number;
  totalRecipeCost: number;
  
  // Priority
  hasExpiringItems: boolean;
  expiringItemsCount: number;
}

interface IngredientMatch {
  ingredientId: string;
  name: string;
  quantity: number;
  unit: string;
  optional?: boolean;
  estimatedCost?: number;
  available?: number;
  isExpiringSoon?: boolean;
}
```

---

## UI Integration

### Правильный запрос:
```typescript
const response = await fetch('/api/recipes/match?onlyCookable=true&minScore=100', {
  headers: {
    'Authorization': `Bearer ${token}`
  }
});

const data = await response.json();

if (data.success && data.data.recipes.length > 0) {
  const recipe = data.data.recipes[0];
  console.log('Can cook:', recipe.localName);
  console.log('Missing:', recipe.missingIngredients.length);
}
```

### Неправильные запросы:
```typescript
// ❌ Не используйте
fetch('/api/recipes/recommendations')  // Другой сценарий
fetch('/api/recipes/recommend')        // Не существует
```

---

## Status: ✅ TESTED & WORKING

- Тест: `make test-egg` ✅
- Manual: проверено вручную ✅
- Response: правильная структура ✅
- Filters: работают ✅
