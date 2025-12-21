# Recipe Recalculation API 🔄

## Endpoint

```
POST /api/ai/recipe/recalculate
```

**Auth:** Required (Bearer token)

## Purpose

Пересчитать экономику рецепта после обновления холодильника:
- ✅ Новые цены на продукты
- ✅ Добавлены missing ingredients
- ✅ Удалены expired items
- ✅ Изменилось количество products

**НЕ меняет:** steps, title, description, chefTips
**Обновляет:** economy, ingredientsMissing, expiryPriority, usedProducts

---

## Request

### Headers
```http
Authorization: Bearer <jwt_token>
Content-Type: application/json
```

### Body
```json
{
  "recipe": {
    "name": "Smażone ogórki w sosie mlecznym",
    "description": "Kremowy sos z ogórkami",
    "ingredientsUsed": [
      {"name": "Ogórek", "quantity": 500, "unit": "g"},
      {"name": "Mleko 3.2%", "quantity": 200, "unit": "ml"},
      {"name": "Cebula", "quantity": 100, "unit": "g"}
    ],
    "ingredientsMissing": [
      {"name": "Sól", "quantity": 5, "unit": "g"},
      {"name": "Masło", "quantity": 50, "unit": "g"}
    ],
    "steps": [
      "Krok 1: Pokrój ogórki...",
      "Krok 2: Rozgrzej mleko..."
    ],
    "cookingTime": 30,
    "chefTips": ["Tip 1", "Tip 2"],
    "expiryPriority": "warning",
    "economy": {
      "usedFromFridge": true,
      "usedValue": 8.50,
      "estimatedExtraCost": 2.00,
      "savedMoney": 6.50,
      "currency": "PLN"
    }
  },
  "language": "pl"
}
```

### Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `recipe` | RestaurantRecipe | ✅ Yes | Текущий рецепт (весь объект) |
| `recipe.name` | string | ✅ Yes | Название рецепта |
| `recipe.ingredientsUsed` | array | ✅ Yes | Продукты из холодильника |
| `recipe.ingredientsMissing` | array | ✅ Yes | Недостающие продукты |
| `recipe.steps` | array | ✅ Yes | Шаги приготовления (не меняются) |
| `language` | string | ❌ No | Язык (для будущих AI запросов) |

---

## Response

### Success (200 OK)

```json
{
  "success": true,
  "data": {
    "recipe": {
      "name": "Smażone ogórki w sosie mlecznym",
      "description": "Kremowy sos z ogórkami",
      "ingredientsUsed": [
        {"name": "Ogórek", "quantity": 500, "unit": "g"},
        {"name": "Mleko 3.2%", "quantity": 200, "unit": "ml"},
        {"name": "Cebula", "quantity": 100, "unit": "g"},
        {"name": "Masło", "quantity": 50, "unit": "g"}
      ],
      "ingredientsMissing": [
        {"name": "Sól", "quantity": 5, "unit": "g"}
      ],
      "steps": [
        "Krok 1: Pokrój ogórki...",
        "Krok 2: Rozgrzej mleko..."
      ],
      "cookingTime": 30,
      "chefTips": ["Tip 1", "Tip 2"],
      "expiryPriority": "critical",
      "economy": {
        "usedFromFridge": true,
        "usedValue": 12.43,
        "estimatedExtraCost": 0.50,
        "savedMoney": 11.93,
        "currency": "PLN"
      }
    },
    "usedProducts": [
      {
        "name": "Ogórek",
        "quantityUsed": 500,
        "unit": "g",
        "pricePerUnit": 0.007,
        "usedCost": 3.50,
        "currency": "PLN",
        "daysLeft": 1
      },
      {
        "name": "Mleko 3.2%",
        "quantityUsed": 200,
        "unit": "ml",
        "pricePerUnit": 0.00324,
        "usedCost": 0.65,
        "currency": "PLN",
        "daysLeft": 3
      },
      {
        "name": "Cebula",
        "quantityUsed": 100,
        "unit": "g",
        "pricePerUnit": 0.00345,
        "usedCost": 0.35,
        "currency": "PLN",
        "daysLeft": 7
      },
      {
        "name": "Masło",
        "quantityUsed": 50,
        "unit": "g",
        "pricePerUnit": 0.16,
        "usedCost": 8.00,
        "currency": "PLN",
        "daysLeft": 10
      }
    ]
  }
}
```

### Что изменилось:

1. **ingredientsMissing:** Масло удалено (теперь в холодильнике)
2. **ingredientsUsed:** Масло добавлено (было missing, теперь в fridge)
3. **economy.usedValue:** 8.50 → 12.43 PLN (добавилась стоимость масла)
4. **economy.savedMoney:** 6.50 → 11.93 PLN
5. **expiryPriority:** warning → critical (огурец expires in 1 day)
6. **usedProducts:** Добавлен расчёт для масла

### Error Responses

#### 400 Bad Request
```json
{
  "error": "recipe name is required"
}
```

#### 401 Unauthorized
```json
{
  "error": "unauthorized"
}
```

#### 500 Internal Server Error
```json
{
  "error": "failed to recalculate recipe"
}
```

---

## Logic Flow

### 1. Match Ingredients with Fridge

```go
fridgeMap := make(map[string]FridgeItem)
for _, item := range fridgeItems {
    fridgeMap[strings.ToLower(item.Name)] = item
}

for _, ingredient := range recipe.IngredientsUsed {
    if fridgeItem, exists := fridgeMap[ingredient.Name]; exists {
        // Calculate cost
        if fridgeItem.PricePerUnit != nil {
            cost = ingredient.Quantity × fridgeItem.PricePerUnit
            totalUsedCost += cost
        }
    }
}
```

### 2. Update ingredientsMissing

```go
updatedMissing := []Ingredient{}
for _, missing := range recipe.IngredientsMissing {
    // If now in fridge, skip
    if _, inFridge := fridgeMap[missing.Name]; inFridge {
        continue // Remove from missing
    }
    
    // Still missing
    updatedMissing = append(updatedMissing, missing)
}
```

### 3. Recalculate Economy

```go
usedValue = Σ(quantity × pricePerUnit)  // For all ingredients in fridge
savedMoney = usedValue - estimatedExtraCost
```

### 4. Update expiryPriority

```go
if any ingredient has status="critical" {
    expiryPriority = "critical"
} else if any ingredient has status="warning" {
    expiryPriority = "warning"
} else {
    expiryPriority = "ok"
}
```

---

## Use Cases

### Use Case 1: User Adds Missing Ingredients

**Before:**
```json
{
  "ingredientsMissing": [
    {"name": "Sól", "quantity": 5, "unit": "g"},
    {"name": "Masło", "quantity": 50, "unit": "g"}
  ],
  "economy": {
    "usedValue": 8.50,
    "estimatedExtraCost": 3.00,
    "savedMoney": 5.50
  }
}
```

**User Action:** Adds "Masło" to fridge with price 0.16 PLN/g

**After Recalculation:**
```json
{
  "ingredientsUsed": [
    ...existing...,
    {"name": "Masło", "quantity": 50, "unit": "g"}
  ],
  "ingredientsMissing": [
    {"name": "Sól", "quantity": 5, "unit": "g"}
  ],
  "economy": {
    "usedValue": 16.50,    // ↑ 8.00 PLN (50g × 0.16)
    "estimatedExtraCost": 0.50,
    "savedMoney": 16.00    // ↑ 10.50 PLN
  }
}
```

### Use Case 2: User Adds Prices

**Before:**
```json
{
  "economy": {
    "usedValue": 0.00,  // No prices in fridge
    "savedMoney": 0.00
  },
  "usedProducts": [
    {
      "name": "Ogórek",
      "usedCost": 0.00,
      "pricePerUnit": 0.00
    }
  ]
}
```

**User Action:** Adds price 0.007 PLN/g to "Ogórek"

**After Recalculation:**
```json
{
  "economy": {
    "usedValue": 3.50,    // 500g × 0.007
    "savedMoney": 3.50
  },
  "usedProducts": [
    {
      "name": "Ogórek",
      "usedCost": 3.50,
      "pricePerUnit": 0.007
    }
  ]
}
```

### Use Case 3: Expired Item Removed

**Before:**
```json
{
  "ingredientsUsed": [
    {"name": "Ogórek", "quantity": 500, "unit": "g"}
  ],
  "expiryPriority": "critical"
}
```

**User Action:** Removes expired "Ogórek" from fridge

**After Recalculation:**
```json
{
  "ingredientsUsed": [
    {"name": "Ogórek", "quantity": 500, "unit": "g"}
  ],
  "ingredientsMissing": [
    ...existing...,
    {"name": "Ogórek", "quantity": 500, "unit": "g"}
  ],
  "expiryPriority": "ok",
  "economy": {
    "usedValue": 0.00  // No cost (not in fridge)
  }
}
```

---

## Frontend Integration

### When to Call This Endpoint

✅ **After adding missing ingredients:**
```javascript
// User clicks "Dodaj do lodówki"
await addMissingIngredients(missingIngredients)

// Recalculate recipe economy
const updated = await recalculateRecipe(currentRecipe)
setRecipe(updated.data.recipe)
```

✅ **After adding/updating prices:**
```javascript
// User updates price
await updateFridgeItemPrice(itemId, newPrice)

// Recalculate all saved recipes
for (const recipe of savedRecipes) {
    const updated = await recalculateRecipe(recipe)
    // Update recipe in state
}
```

✅ **After removing items:**
```javascript
// User deletes expired item
await deleteFridgeItem(itemId)

// Recalculate recipes using that item
const updated = await recalculateRecipe(currentRecipe)
```

### Example Implementation

```typescript
interface RecalculateRequest {
  recipe: RestaurantRecipe
  language?: string
}

async function recalculateRecipe(recipe: RestaurantRecipe): Promise<RecalculateResponse> {
  const response = await fetch('/api/ai/recipe/recalculate', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      recipe: recipe,
      language: 'pl'
    })
  })
  
  if (!response.ok) {
    throw new Error('Recalculation failed')
  }
  
  return await response.json()
}

// Usage
const updated = await recalculateRecipe(currentRecipe)

console.log('Old economy:', currentRecipe.economy.usedValue)
console.log('New economy:', updated.data.recipe.economy.usedValue)
console.log('Saved:', updated.data.recipe.economy.savedMoney, 'PLN')
```

---

## Performance

### Response Time
- **Typical:** 50-150ms (database queries only)
- **No AI calls** (instant calculation)
- **No external API dependencies**

### Optimization
- ✅ Single DB query for all fridge items
- ✅ In-memory matching (map lookup)
- ✅ O(n) complexity for ingredient matching

---

## Testing

### Test Case 1: Basic Recalculation
```bash
curl -X POST https://api.example.com/api/ai/recipe/recalculate \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "recipe": {
      "name": "Test Recipe",
      "ingredientsUsed": [
        {"name": "Ogórek", "quantity": 500, "unit": "g"}
      ],
      "ingredientsMissing": [],
      "steps": ["Step 1"],
      "cookingTime": 30,
      "chefTips": [],
      "expiryPriority": "ok",
      "economy": {
        "usedFromFridge": true,
        "usedValue": 0,
        "estimatedExtraCost": 0,
        "savedMoney": 0,
        "currency": "PLN"
      }
    }
  }'
```

### Expected Logs
```
[AI][RECALC] Starting recalculation for recipe: Test Recipe
[AI][RECALC] User: xxx, Fridge items: 6
[AI][RECALC] Processing 1 ingredients used in recipe
[AI][RECALC] ✅ Ogórek: 500.00 g × 0.007000 = 3.50 PLN
[AI][RECALC] Economy: UsedValue=3.50, SavedMoney=3.50, ExpiryPriority=ok
[AI][RECALC] ✅ Recalculation complete: 1 used products, 3.50 PLN total value
```

---

## Related Endpoints

- `POST /api/ai/create-recipe-from-fridge` - Generate new recipe
- `POST /api/ai/add-missing-ingredients` - Add missing to fridge
- `GET /api/fridge/items` - List fridge items
- `PUT /api/fridge/items/:id/price` - Update price

---

## Notes

### Why Not Regenerate Recipe?

**Regenerating** (calling AI again):
- ❌ Slow (1-3 seconds)
- ❌ Different recipe each time
- ❌ Costs tokens
- ❌ May fail (AI errors)
- ❌ Loses user's saved steps/tips

**Recalculating** (this endpoint):
- ✅ Fast (50-150ms)
- ✅ Same recipe preserved
- ✅ Free (no AI calls)
- ✅ Always succeeds
- ✅ Keeps user's modifications

### When to Regenerate vs Recalculate

**Regenerate** when:
- User wants completely new recipe
- Fridge contents drastically changed
- Different cuisine/cooking method desired

**Recalculate** when:
- User adds missing ingredients
- User updates prices
- User removes expired items
- Just need updated economy/costs

---

**Related docs:**
- [ECONOMY_CALCULATION.md](./ECONOMY_CALCULATION.md) - Economy formula
- [PRICE_FLOW_DEBUG.md](./PRICE_FLOW_DEBUG.md) - Price debugging
- [INGREDIENT_API_QUICK_REF.md](./INGREDIENT_API_QUICK_REF.md) - Fridge API
