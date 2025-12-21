# Recipe Recalculation Quick Reference 🔄

## Endpoint
```
POST /api/ai/recipe/recalculate
Auth: Bearer token required
```

## Quick Request
```bash
curl -X POST $API_URL/api/ai/recipe/recalculate \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"recipe": {...}, "language": "pl"}'
```

## What It Does

✅ **Recalculates:**
- Economy (usedValue, savedMoney)
- ingredientsMissing (removes items now in fridge)
- expiryPriority (critical/warning/ok)
- usedProducts (with updated costs)

❌ **Does NOT change:**
- steps
- title/name
- description
- chefTips

## When to Use

| Scenario | Action |
|----------|--------|
| User adds missing ingredients | ✅ Recalculate |
| User updates prices | ✅ Recalculate |
| User removes expired items | ✅ Recalculate |
| Want completely different recipe | ❌ Regenerate instead |

## Response Time
- **50-150ms** (no AI calls)
- **Instant** updates
- **Always succeeds** (no LLM failures)

## Minimal Request
```json
{
  "recipe": {
    "name": "Recipe Name",
    "ingredientsUsed": [...],
    "ingredientsMissing": [...],
    "steps": [...],
    "cookingTime": 30,
    "chefTips": [],
    "expiryPriority": "ok",
    "economy": {...}
  }
}
```

## Minimal Response
```json
{
  "success": true,
  "data": {
    "recipe": {
      "name": "Recipe Name",
      "economy": {
        "usedValue": 12.43,
        "savedMoney": 11.93
      },
      "expiryPriority": "critical"
    },
    "usedProducts": [...]
  }
}
```

## Frontend Example
```typescript
// After adding missing ingredients
await addToFridge(missingIngredients)
const updated = await recalculateRecipe(currentRecipe)
setRecipe(updated.data.recipe)

// Show toast
toast.success(`Economy updated: ${updated.data.recipe.economy.usedValue} PLN`)
```

## Logic
```
1. Match recipe.ingredientsUsed with current fridge
2. Calculate: usedValue = Σ(quantity × pricePerUnit)
3. Update ingredientsMissing (remove items now in fridge)
4. Set expiryPriority based on item status
5. Return updated recipe
```

## Comparison

| Feature | Recalculate | Regenerate |
|---------|-------------|-----------|
| Speed | 50-150ms ⚡ | 1-3s 🐌 |
| AI calls | None | Yes |
| Cost | Free | Tokens |
| Recipe | Preserved ✅ | New one ⚠️ |
| Success rate | 100% | ~90% |

## Use Recalculate When:
- ✅ Just need updated economy
- ✅ User modified fridge
- ✅ Prices changed
- ✅ Want instant results

## Use Regenerate When:
- ❌ Want completely new recipe
- ❌ Fridge drastically changed
- ❌ Different cuisine needed

---

**Full docs:** [RECIPE_RECALCULATION_API.md](./RECIPE_RECALCULATION_API.md)
