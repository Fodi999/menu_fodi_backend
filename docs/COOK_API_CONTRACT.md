# Cook Recipe API Contract

## POST /api/recipes/{id}/cook

### Request
```json
{
  "servingsMultiplier": 1.0,
  "idempotencyKey": "optional-unique-key"
}
```

### Success Response (200)
```json
{
  "success": true,
  "data": {
    "cookLogId": "ab58e23f-2b4a-4737-a497-18cbdd8df9ca",
    "recipeId": "92691aae-c3af-427d-aaed-1408319f0a3c",
    "canonicalName": "Greek Salad",
    "localName": "Sałatka grecka",
    "servingsMultiplier": 1,
    "cookedAt": "2025-12-21T18:55:48+01:00",
    "usedValue": 6.15,
    "wasteRiskSaved": 1.4,
    "totalRecipeCost": 6.15,
    "ingredientsUsed": [
      {
        "ingredientId": "fc57dbf2-39bb-4f30-a8e2-cf6585074587",
        "name": "Pomidor",
        "quantityUsed": 400,
        "unit": "g",
        "pricePerUnit": 0.008,
        "totalCost": 3.2,
        "wasExpiringSoon": false,
        "remainingInFridge": 100
      }
    ],
    "fridgeUpdated": true,
    "remainingItems": 9
  }
}
```

### Error Response (400 - Missing Ingredients)
```json
{
  "success": false,
  "message": "missing ingredients: [Pomidor Oliwa z oliwek]",
  "error": "Failed to cook recipe"
}
```

## ✅ Guaranteed Contract Fields

### ingredientsUsed[] - ALWAYS present:
- ✅ **ingredientId** (UUID string) - ALWAYS present, NEVER empty
- ✅ **name** (string) - Human-readable name
- ✅ **quantityUsed** (number) - Amount deducted
- ✅ **unit** (string) - Measurement unit
- ✅ **remainingInFridge** (number) - Amount left after cooking
- ✅ **wasExpiringSoon** (boolean) - Was close to expiry
- ✅ **pricePerUnit** (number) - Price per unit
- ✅ **totalCost** (number) - Total cost of this ingredient

## Frontend Integration Example

```typescript
interface CookedIngredient {
  ingredientId: string;        // ✅ Use this for fridge updates
  name: string;                // Only for display
  quantityUsed: number;
  unit: string;
  remainingInFridge: number;
  wasExpiringSoon: boolean;
  pricePerUnit: number;
  totalCost: number;
}

async function cookRecipe(recipeId: string) {
  const response = await fetch(`/api/recipes/${recipeId}/cook`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ servingsMultiplier: 1.0 })
  });

  if (!response.ok) {
    const error = await response.json();
    // Show error: "Nie masz wystarczająco: Pomidor"
    showError(error.message);
    return;
  }

  const result = await response.json();
  
  // ✅ Update fridge using ingredientId (NOT name!)
  result.data.ingredientsUsed.forEach((ing: CookedIngredient) => {
    updateFridgeItem(ing.ingredientId, ing.remainingInFridge);
    
    if (ing.remainingInFridge === 0) {
      removeFridgeItem(ing.ingredientId); // Item depleted
    } else if (ing.remainingInFridge < 100) {
      showLowStockWarning(ing.name, ing.remainingInFridge, ing.unit);
    }
  });

  // Show success
  showSuccess(`
    Gotowe! ${result.data.canonicalName}
    Wykorzystano: ${result.data.usedValue} PLN
    Zapobiegłeś marnowaniu: ${result.data.wasteRiskSaved} PLN
  `);

  // Refresh recipe matches (other recipes now closer)
  refreshRecipeMatches();
}
```

## ⚠️ IMPORTANT: Always use ingredientId

**DO NOT match by name!** Names can be:
- In different languages (Polish, English)
- Have typos or variations
- Be similar for different ingredients

**ALWAYS use `ingredientId` for:**
- Updating fridge quantities
- Removing depleted items
- Matching with fridge data
- Any data synchronization

## Database Verification

All 24 recipe ingredients have valid `ingredientId`:
```sql
SELECT COUNT(*) as total, COUNT("ingredientId") as with_id 
FROM "RecipeIngredient";
-- Result: total=24, with_id=24 ✅
```

## Idempotency

Sending the same `idempotencyKey` twice will:
- Return the same cook log
- NOT deduct ingredients again
- Return same `cookedAt` timestamp

This prevents double-cooking on accidental double-click.

## Transaction Safety

If ANY error occurs during cooking:
- All changes are rolled back
- Fridge remains unchanged
- No cook log is created
- 400 error is returned with reason
