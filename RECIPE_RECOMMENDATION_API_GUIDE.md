# 📘 Recipe Recommendation API Guide (Production)

## Endpoint

```
GET /api/recipe-recommendations?lang={lang}&limit={limit}
```

**Auth**: Bearer JWT token (required)

---

## Query Parameters

| Parameter | Type   | Required | Default | Description                      |
|-----------|--------|----------|---------|----------------------------------|
| `lang`    | string | No       | `pl`    | Language: `pl`, `en`, `ru`       |
| `limit`   | int    | No       | `10`    | Max recipes to return (top-N)    |

---

## Response Structure

```typescript
interface RecipeRecommendationResponse {
  decision: 'ready' | 'almost_ready' | 'need_more';  // Overall verdict
  summary: string;                                     // Localized summary
  total_matches: number;                               // Total recipes found
  recipes: RecipeMatchResult[];                        // Sorted by match_percent DESC
}

interface RecipeMatchResult {
  // Identification
  id: string;
  canonical_name: string;  // Stable key (language-independent)
  title: string;           // Localized name
  
  // Matching Metrics (Rules Engine)
  match_percent: number;         // 0-100
  match_status: 'ready' | 'almost_ready' | 'not_ready';
  missing_count: number;         // How many ingredients missing
  available_count: number;       // How many ingredients available
  total_required: number;        // Total ingredients needed
  
  // Ingredients (Detailed Objects)
  missing_ingredients: IngredientInfo[];    // ❌ What's missing
  available_ingredients: IngredientInfo[];  // ✅ What's available
  
  // Recipe Metadata
  cook_time: number;   // minutes
  portions: number;    // servings
  image_url: string;   // Cloudinary URL
}

interface IngredientInfo {
  id: string;              // UUID
  canonical_name: string;  // e.g., "растительное_масло"
  display_name: string;    // Localized name: "Растительное масло"
  quantity: number;        // 30
  unit: string;            // "ml", "g", "pcs"
  category: string;        // "condiment", "vegetable", "protein"
}
```

---

## Decision Logic (Rules Engine)

```javascript
if (missing_count === 0) {
  decision = "ready";          // 🟢 Can cook now
} else if (missing_count <= 2) {
  decision = "almost_ready";   // 🟡 Missing 1-2 ingredients
} else {
  decision = "not_ready";      // 🔴 Need 3+ ingredients
}
```

---

## Example Response

```json
{
  "decision": "almost_ready",
  "summary": "Почти готово! Не хватает всего нескольких ингредиентов.",
  "total_matches": 1,
  "recipes": [
    {
      "id": "605c8419-2d42-4ef0-a9d2-839582e98727",
      "canonical_name": "zharenye_yaytsa",
      "title": "Жареные яйца",
      "match_percent": 66.67,
      "match_status": "almost_ready",
      "missing_count": 1,
      "available_count": 2,
      "total_required": 3,
      "missing_ingredients": [
        {
          "id": "9ff773d2-a3ee-4f4b-bc45-4cfe0d7f680b",
          "canonical_name": "растительное_масло",
          "display_name": "Растительное масло",
          "quantity": 30,
          "unit": "ml",
          "category": "condiment"
        }
      ],
      "available_ingredients": [
        {
          "id": "3260aadf-52de-4038-9568-ee536495224a",
          "canonical_name": "яйца",
          "display_name": "Яйца",
          "quantity": 3,
          "unit": "pcs",
          "category": "egg"
        },
        {
          "id": "c4d477f8-9123-4175-b515-5201ee1ff61b",
          "canonical_name": "соль",
          "display_name": "Соль",
          "quantity": 2,
          "unit": "g",
          "category": "condiment"
        }
      ],
      "cook_time": 7,
      "portions": 1,
      "image_url": "https://res.cloudinary.com/.../recipe_605c8419.webp"
    }
  ]
}
```

---

## Frontend Usage Examples

### 1. Display Missing Ingredients

```typescript
const { recipes } = await getRecommendations({ lang: 'ru', limit: 5 });

recipes.forEach(recipe => {
  if (recipe.match_status === 'almost_ready') {
    console.log(`Рецепт: ${recipe.title}`);
    console.log('Купить:');
    
    recipe.missing_ingredients.forEach(ing => {
      console.log(`- ${ing.display_name}: ${ing.quantity} ${ing.unit}`);
    });
  }
});
```

### 2. Calculate Shopping List Total

```typescript
const shoppingList = recipe.missing_ingredients.map(ing => ({
  name: ing.display_name,
  quantity: ing.quantity,
  unit: ing.unit,
  price: calculatePrice(ing.id, ing.quantity) // Use DB prices
}));

const totalCost = shoppingList.reduce((sum, item) => sum + item.price, 0);
```

### 3. Group by Match Status

```typescript
const ready = recipes.filter(r => r.match_status === 'ready');
const almostReady = recipes.filter(r => r.match_status === 'almost_ready');
const needMore = recipes.filter(r => r.match_status === 'not_ready');

// Display in separate sections with badges
<Badge color="green">{ready.length} готовых</Badge>
<Badge color="yellow">{almostReady.length} почти готовых</Badge>
```

### 4. Scale Portions

```typescript
function scaleRecipe(recipe: RecipeMatchResult, newPortions: number) {
  const multiplier = newPortions / recipe.portions;
  
  return {
    ...recipe,
    missing_ingredients: recipe.missing_ingredients.map(ing => ({
      ...ing,
      quantity: ing.quantity * multiplier
    })),
    available_ingredients: recipe.available_ingredients.map(ing => ({
      ...ing,
      quantity: ing.quantity * multiplier
    }))
  };
}
```

---

## Why Objects (not strings)?

✅ **Scalability**: Can calculate costs, scale portions, show units  
✅ **Shopping Cart**: Direct integration with e-commerce  
✅ **AI Features**: Pass structured data to substitution engine  
✅ **Economy**: Track ingredient prices and suggest cheaper alternatives  
✅ **Future-proof**: Can add fields without breaking clients  

---

## Migration from Legacy API

If migrating from `/api/ai-recipe/recommendation`:

**Legacy format**:
```json
{
  "ingredients": ["яйца", "соль"],
  "missingIngredients": ["растительное масло"]
}
```

**New format** (more powerful):
```json
{
  "available_ingredients": [
    {"display_name": "Яйца", "quantity": 3, "unit": "pcs"},
    {"display_name": "Соль", "quantity": 2, "unit": "g"}
  ],
  "missing_ingredients": [
    {"display_name": "Растительное масло", "quantity": 30, "unit": "ml"}
  ]
}
```

**Simple adapter** (if needed):
```typescript
// Extract just names for legacy UI
const ingredientNames = recipe.available_ingredients.map(i => i.display_name);
const missingNames = recipe.missing_ingredients.map(i => i.display_name);
```

---

## Performance

- **Response time**: < 300ms (typical)
- **Caching**: Planned (5 min TTL)
- **Batch processing**: Planned (precompute for active users)

---

## Testing

```bash
# Test production endpoint
curl -s "https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/recipe-recommendations?lang=ru&limit=5" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" | jq '.'

# Test all languages
for lang in pl en ru; do
  echo "Testing $lang..."
  curl -s "https://.../?lang=$lang&limit=3" -H "Authorization: Bearer ..." | jq '.summary'
done
```

---

## Error Handling

```typescript
try {
  const response = await fetch('/api/recipe-recommendations?lang=ru');
  const data = await response.json();
  
  if (data.error) {
    // Handle backend error
    console.error(data.message);
  }
} catch (error) {
  // Handle network error
  console.error('Failed to fetch recommendations');
}
```

Common errors:
- `401 Unauthorized` - Invalid/expired JWT token
- `"fridge is empty"` - User has no ingredients
- `"no recipes available"` - Catalog is empty (should not happen in production)

---

## Next Steps (Phase 2)

- [ ] AI Explanations (`ai_explanation` field)
- [ ] Substitution suggestions
- [ ] Redis caching (5 min TTL)
- [ ] Batch processing for active users
- [ ] WebSocket real-time updates when fridge changes

---

**Status**: ✅ Production Ready (v1.0)  
**Last Updated**: 2026-01-22  
**Endpoint**: `/api/recipe-recommendations`
