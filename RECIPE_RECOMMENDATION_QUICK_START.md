# 🚀 Recipe Recommendations Quick Start

## TL;DR

```bash
GET /api/recipe-recommendations?lang=ru&limit=5
Authorization: Bearer <JWT>
```

Returns recipes you can cook with your fridge ingredients.

---

## Response Structure (Simplified)

```json
{
  "decision": "almost_ready",           // 🟡 Overall verdict
  "summary": "Почти готово!",           // User-friendly message
  "recipes": [
    {
      "title": "Жареные яйца",
      "match_status": "almost_ready",   // 🟡 This recipe status
      "match_percent": 66.67,           // 2/3 ingredients available
      
      "available_ingredients": [        // ✅ What you have
        {"display_name": "Яйца", "quantity": 3, "unit": "pcs"},
        {"display_name": "Соль", "quantity": 2, "unit": "g"}
      ],
      
      "missing_ingredients": [          // ❌ What you need to buy
        {"display_name": "Растительное масло", "quantity": 30, "unit": "ml"}
      ]
    }
  ]
}
```

---

## Status Logic

| Scenario      | Missing Count | Badge | Action                     |
|---------------|---------------|-------|----------------------------|
| `ready`       | 0             | 🟢    | Cook now!                  |
| `almost_ready`| 1-2           | 🟡    | Buy 1-2 ingredients        |
| `not_ready`   | 3+            | 🔴    | Need to shop more          |

---

## Frontend Examples

### Show Recipe Card
```tsx
<RecipeCard
  title={recipe.title}
  status={recipe.match_status}
  image={recipe.image_url}
  matchPercent={recipe.match_percent}
  missing={recipe.missing_ingredients}
/>
```

### Shopping List Button
```tsx
{recipe.match_status === 'almost_ready' && (
  <Button onClick={() => addToShoppingList(recipe.missing_ingredients)}>
    Купить {recipe.missing_count} ингредиента
  </Button>
)}
```

### Filter by Status
```tsx
const readyRecipes = recipes.filter(r => r.match_status === 'ready');
const almostReadyRecipes = recipes.filter(r => r.match_status === 'almost_ready');
```

---

## TypeScript Types

```typescript
interface RecipeRecommendationResponse {
  decision: 'ready' | 'almost_ready' | 'not_ready';
  summary: string;
  total_matches: number;
  recipes: RecipeMatchResult[];
}

interface RecipeMatchResult {
  id: string;
  title: string;
  match_status: 'ready' | 'almost_ready' | 'not_ready';
  match_percent: number;
  missing_count: number;
  available_ingredients: IngredientInfo[];
  missing_ingredients: IngredientInfo[];
  cook_time: number;
  portions: number;
  image_url: string;
}

interface IngredientInfo {
  id: string;
  display_name: string;  // Localized
  quantity: number;
  unit: string;          // "ml", "g", "pcs"
  category: string;
}
```

---

## Why Objects Instead of Strings?

**Old (Legacy)**:
```json
{
  "ingredients": ["яйца", "соль"],
  "missingIngredients": ["растительное масло"]
}
```

❌ Can't calculate quantities  
❌ Can't show units  
❌ Can't calculate prices  
❌ Can't scale portions  

**New (Production)**:
```json
{
  "available_ingredients": [
    {"display_name": "Яйца", "quantity": 3, "unit": "pcs"}
  ],
  "missing_ingredients": [
    {"display_name": "Растительное масло", "quantity": 30, "unit": "ml"}
  ]
}
```

✅ Full ingredient details  
✅ Shopping cart integration  
✅ Price calculation  
✅ Portion scaling  
✅ Future-proof (extensible)  

---

## Testing

```bash
# Get token
TOKEN=$(curl -s -X POST https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"fodi85@gmail.ru","password":"YOUR_PASSWORD"}' | jq -r '.token')

# Get recommendations
curl -s "https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/recipe-recommendations?lang=ru&limit=5" \
  -H "Authorization: Bearer $TOKEN" | jq '.'
```

---

## Full Documentation

See: [RECIPE_RECOMMENDATION_API_GUIDE.md](./RECIPE_RECOMMENDATION_API_GUIDE.md)

---

**Status**: ✅ Production Ready  
**Endpoint**: `/api/recipe-recommendations`  
**Version**: 1.0  
**Last Updated**: 2026-01-22
