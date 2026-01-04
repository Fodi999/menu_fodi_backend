# Recipe Portion Weights

## Overview
Each recipe in the catalog now includes **`portionWeightGrams`** field - the total weight of one serving in grams.

## Database Schema
```sql
ALTER TABLE "Recipe" ADD COLUMN "portionWeightGrams" INTEGER;
```

## Model Field
```go
type RecipeCatalog struct {
    // ...
    Servings            int   `gorm:"not null;default:1" json:"servings"`
    PortionWeightGrams  *int  `gorm:"column:portionWeightGrams" json:"portionWeightGrams,omitempty"`
    // ...
}
```

## Current Recipe Weights (Per Serving)

| Recipe Name                    | Portion Weight | Servings | Notes                           |
|--------------------------------|----------------|----------|---------------------------------|
| **Scrambled Eggs**             | 150g          | 1        | Light breakfast                 |
| **Polish Meat Dumplings**      | 180g          | 1        | 6-8 pieces                      |
| **Pierogi Ruskie**             | 200g          | 1        | 6-8 pieces                      |
| **Polish Potato Pancakes**     | 200g          | 1        | 3-4 pancakes                    |
| **Polish Breaded Pork Chop**   | 220g          | 1        | Single pork chop with coating   |
| **Greek Salad**                | 250g          | 1        | Fresh vegetable salad           |
| **Spaghetti Carbonara**        | 300g          | 1        | Pasta with sauce                |
| **Pizza Margherita**           | 350g          | 1        | Full pizza portion              |
| **Polish Chicken Soup**        | 400g          | 1        | Bowl of soup                    |
| **Polish Hunters Stew (Bigos)**| 450g          | 1        | Hearty stew                     |

## API Response Example
```json
{
  "data": [
    {
      "id": "...",
      "canonicalName": "Pierogi Ruskie",
      "namePl": "Pierogi ruskie",
      "servings": 1,
      "portionWeightGrams": 200,
      "timeMinutes": 45,
      "difficulty": "medium",
      "category": "main",
      "ingredients": [...]
    }
  ],
  "meta": {
    "total": 10,
    "count": 10
  }
}
```

## Usage in Frontend

### Display Portion Information
```typescript
interface Recipe {
  servings: number;           // Always 1 (base portion)
  portionWeightGrams?: number; // e.g., 200g
}

// Display example:
// "Jedna porcia: 200g"
// "One serving: 200g"
```

### Calculate Total Weight for Multiple Servings
```typescript
const servingsMultiplier = 4; // User wants 4 portions
const totalWeight = recipe.portionWeightGrams * servingsMultiplier;
// Example: 200g × 4 = 800g
```

### Recipe Card Display
```tsx
<RecipeCard>
  <h3>{recipe.namePl}</h3>
  <div>
    <span>🍽️ Porcia: {recipe.portionWeightGrams}g</span>
    <span>⏱️ Czas: {recipe.timeMinutes} min</span>
    <span>👨‍🍳 Poziom: {recipe.difficulty}</span>
  </div>
</RecipeCard>
```

## Weight Calculation Logic

### How Weights Were Determined
The portion weights are approximate values based on:
1. **Ingredient quantities** - Sum of all ingredients per serving
2. **Cooking method** - Water loss/gain during cooking
3. **Standard serving sizes** - Industry standards and nutritional guidelines
4. **Cultural context** - Polish portion norms

### Future Automation
To automatically calculate weights from ingredients:
```go
func CalculatePortionWeight(recipeID uuid.UUID) (int, error) {
    var ingredients []CatalogIngredient
    db.Where("recipe_id = ?", recipeID).Find(&ingredients)
    
    totalWeight := 0
    for _, ing := range ingredients {
        totalWeight += int(ing.Quantity) // Assuming grams
    }
    
    // Apply cooking loss factor (e.g., 0.8 for boiling)
    return int(float64(totalWeight) * cookingLossFactor), nil
}
```

## Migration History
- **Migration 061**: Normalized all recipes to `servings = 1` (base portion)
- **Migration 062**: Added `portionWeightGrams` field with initial values

## Endpoints Using This Field
- `GET /api/admin/recipes` - Returns all recipes with portion weights
- `GET /api/recipes/:id` - Returns single recipe with portion weight
- `GET /api/recipes/match` - Includes portion weights in matching results

## Benefits
✅ **Transparency** - Users know exact portion sizes  
✅ **Nutrition Tracking** - Better calorie/macro calculations  
✅ **Meal Planning** - Accurate shopping lists  
✅ **Scaling** - Easy to calculate for multiple servings  
✅ **Standardization** - All recipes use consistent base (1 serving)  

## Notes
- Weights are **approximate** and may vary based on ingredient sizes
- All recipes use `servings = 1` as the base unit
- Frontend should multiply by `servingsMultiplier` for multiple portions
- Consider adding weight ranges (e.g., "180-220g") in future iterations
