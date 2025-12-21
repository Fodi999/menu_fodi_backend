# Recipe Match API Contract

## GET /api/recipes/match

### Query Parameters
```
testUserID=<uuid>           // Dev mode only (remove in production)
country=Poland              // Filter by country
difficulty=easy             // Filter: easy, medium, hard
maxTime=30                  // Max cooking time in minutes
dietTags=vegetarian,vegan   // Comma-separated diet tags
excludeAllergens=gluten     // Comma-separated allergens to exclude
minScore=0                  // Minimum match score (0-100), default: 0
onlyCookable=true           // Only recipes that can be cooked now
limit=10                    // Max results, default: 10
```

### Response
```json
{
  "success": true,
  "data": {
    "recipes": [
      {
        "recipeId": "92691aae-c3af-427d-aaed-1408319f0a3c",
        "canonicalName": "Greek Salad",
        "localName": "Sałatka grecka",
        "country": "Greece",
        "category": "salad",
        "difficulty": "easy",
        "timeMinutes": 15,
        "servings": 4,
        
        "score": 52.67,
        "coverage": 0.75,
        "canCookNow": false,
        
        "usedValue": 2.95,
        "savedMoney": 2.95,
        "costToComplete": 4.0,
        "totalRecipeCost": 6.95,
        "wasteRiskSaved": 1.4,
        
        "hasExpiringItems": true,
        "expiringItemsCount": 1,
        
        "usedIngredients": [
          {
            "ingredientId": "59bf118a-9dae-4ca3-a262-776e18b58338",
            "name": "Ogórek",
            "quantity": 200,
            "unit": "g",
            "available": 3560,
            "isExpiringSoon": true
          }
        ],
        
        "missingIngredients": [
          {
            "ingredientId": "fc57dbf2-39bb-4f30-a8e2-cf6585074587",
            "name": "Pomidor",
            "quantity": 400,
            "unit": "g",
            "optional": false,
            "estimatedCost": 4.0
          }
        ],
        
        "allergens": [],
        "dietTags": ["vegetarian", "healthy"]
      }
    ],
    "count": 1
  }
}
```

## ✅ Guaranteed Contract Fields

### Recipe Identity (always present):
- ✅ **recipeId** (UUID) - Use for cooking
- ✅ **canonicalName** (string) - English name
- ✅ **localName** (string) - Polish name
- ✅ **country** (string) - Recipe origin
- ✅ **difficulty** (string) - easy/medium/hard
- ✅ **timeMinutes** (number) - Cooking time

### Match Results (always present):
- ✅ **score** (0-100) - Higher = better match
- ✅ **coverage** (0-1) - Fraction of ingredients you have
- ✅ **canCookNow** (boolean) - All required ingredients available

### Economy (always present, PLN):
- ✅ **usedValue** - Cost of ingredients from fridge
- ✅ **savedMoney** - Same as usedValue (for UI: "Oszczędności")
- ✅ **costToComplete** - Cost to buy missing ingredients
- ✅ **totalRecipeCost** - Full recipe cost (usedValue + costToComplete)
- ✅ **wasteRiskSaved** - Value of expiring items used

### Ingredients (arrays, may be empty):
- ✅ **usedIngredients[]** - What you have
  - ✅ **ingredientId** (UUID) - ALWAYS present
  - ✅ **name** (string)
  - ✅ **quantity** (number) - Required amount
  - ✅ **unit** (string)
  - ✅ **available** (number) - How much you have
  - ✅ **isExpiringSoon** (boolean)

- ✅ **missingIngredients[]** - What to buy
  - ✅ **ingredientId** (UUID) - ALWAYS present
  - ✅ **name** (string)
  - ✅ **quantity** (number) - Required amount
  - ✅ **unit** (string)
  - ✅ **optional** (boolean) - Can skip this
  - ✅ **estimatedCost** (number) - Cost in PLN

## Frontend Integration Example

```typescript
interface RecipeMatch {
  recipeId: string;
  canonicalName: string;
  localName: string;
  
  score: number;
  canCookNow: boolean;
  
  usedValue: number;
  savedMoney: number;
  costToComplete: number;
  totalRecipeCost: number;
  wasteRiskSaved: number;
  
  usedIngredients: Ingredient[];
  missingIngredients: Ingredient[];
}

interface Ingredient {
  ingredientId: string;  // ✅ Use this for all operations
  name: string;          // Only for display
  quantity: number;
  unit: string;
}

// Display recipe card
function RecipeCard({ recipe }: { recipe: RecipeMatch }) {
  return (
    <div className="recipe-card">
      <h3>{recipe.localName}</h3>
      
      {/* Economy Display */}
      <div className="economy">
        <div>Wartość z lodówki: {recipe.usedValue} PLN</div>
        <div>Do dokupienia: {recipe.costToComplete} PLN</div>
        <div>Całkowity koszt: {recipe.totalRecipeCost} PLN</div>
        
        {recipe.wasteRiskSaved > 0 && (
          <div className="waste-saved">
            ♻️ Zapobiegłeś marnowaniu: {recipe.wasteRiskSaved} PLN
          </div>
        )}
      </div>
      
      {/* Missing Ingredients */}
      {!recipe.canCookNow && (
        <div className="missing">
          <h4>Potrzebujesz:</h4>
          {recipe.missingIngredients.map(ing => (
            <div key={ing.ingredientId}>
              {ing.name}: {ing.quantity} {ing.unit}
              {ing.optional && " (opcjonalnie)"}
              
              {/* ✅ Use ingredientId for "Add to shopping list" */}
              <button onClick={() => addToShoppingList(ing.ingredientId)}>
                + Lista zakupów
              </button>
            </div>
          ))}
        </div>
      )}
      
      {/* Cook Button */}
      <button 
        onClick={() => cookRecipe(recipe.recipeId)}
        disabled={!recipe.canCookNow}
      >
        {recipe.canCookNow ? "Gotuj teraz!" : "Brakuje składników"}
      </button>
    </div>
  );
}

// Fetch recipes
async function fetchRecipes(filters: {
  country?: string;
  difficulty?: string;
  maxTime?: number;
  onlyCookable?: boolean;
}) {
  const params = new URLSearchParams();
  if (filters.country) params.set('country', filters.country);
  if (filters.difficulty) params.set('difficulty', filters.difficulty);
  if (filters.maxTime) params.set('maxTime', filters.maxTime.toString());
  if (filters.onlyCookable) params.set('onlyCookable', 'true');
  
  const response = await fetch(`/api/recipes/match?${params}`);
  const data = await response.json();
  return data.data.recipes;
}
```

## Sorting Logic (Backend)

Recipes are sorted by:
1. **canCookNow DESC** - Ready-to-cook recipes first
2. **score DESC** - Higher match score first
3. **costToComplete ASC** - Cheaper to complete first
4. **timeMinutes ASC** - Faster recipes first

## Economy Semantics (Important!)

- **usedValue** = Cost of ingredients from your fridge
- **savedMoney** = Same value (UI-friendly: "You're saving...")
- **costToComplete** = What you need to buy
- **totalRecipeCost** = usedValue + costToComplete (full recipe cost)
- **wasteRiskSaved** = Value of expiring items you'll use (waste prevention)

## ⚠️ IMPORTANT: Always use ingredientId

**DO NOT match ingredients by name!**

✅ **Correct:**
```typescript
const ingredient = recipe.usedIngredients.find(
  ing => ing.ingredientId === fridgeItem.ingredientId
);
```

❌ **Wrong:**
```typescript
const ingredient = recipe.usedIngredients.find(
  ing => ing.name === fridgeItem.name  // Names vary by language!
);
```

## Filter Combinations

All filters can be combined:
```
/api/recipes/match?country=Poland&difficulty=easy&maxTime=30&onlyCookable=true
```

This returns:
- Polish recipes only
- Easy difficulty only
- Max 30 minutes cooking
- Only recipes you can cook NOW (no missing ingredients)

## Database Verification

All ingredients in match response have valid `ingredientId`:
```sql
-- Check recipe ingredients
SELECT COUNT(*), COUNT("ingredientId") 
FROM "RecipeIngredient";
-- Result: 24 recipes, all have ingredientId ✅

-- Check ingredient catalog
SELECT COUNT(*), COUNT(id) 
FROM "Ingredient";
-- Result: All catalog items have UUID ✅
```

## Testing Examples

```bash
# Get all recipes
curl "http://localhost:8085/api/recipes/match?testUserID=<uuid>&limit=10"

# Only cookable recipes
curl "http://localhost:8085/api/recipes/match?testUserID=<uuid>&onlyCookable=true"

# Polish easy recipes under 30 minutes
curl "http://localhost:8085/api/recipes/match?testUserID=<uuid>&country=Poland&difficulty=easy&maxTime=30"

# Vegetarian recipes
curl "http://localhost:8085/api/recipes/match?testUserID=<uuid>&dietTags=vegetarian"

# Exclude allergens
curl "http://localhost:8085/api/recipes/match?testUserID=<uuid>&excludeAllergens=gluten,lactose"
```
