# 🚀 Frontend Integration Guide - MVP Recipe System

## 📍 Backend Deployment

**Production URL:** `https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app`

**Status:** ✅ Deployed and working
- Recipe matching API: ✅ Working
- Recipe cooking API: ✅ Working with transactions
- testUserID support: ✅ Enabled for MVP testing

---

## 🔑 Authentication (MVP Mode)

**For MVP testing**, backend accepts `testUserID` query parameter:

```typescript
const TEST_USER_ID = "407582be-59d5-4d21-873b-1a72d31b0d42";

// Add to every API call during MVP phase
const url = `${BASE_URL}/api/recipes/match?testUserID=${TEST_USER_ID}&limit=10`;
```

⚠️ **Production:** Replace with JWT token from auth middleware.

---

## 📡 API Endpoints

### 1. GET /api/recipes/match - Recipe Matching

**Purpose:** Find recipes based on user's fridge contents

**Request:**
```typescript
interface RecipeMatchParams {
  testUserID: string;           // ⚠️ MVP only - remove in production
  limit?: number;               // Default: 10
  country?: string;             // "Poland" | "Italy" | "Greece" | ...
  category?: string;            // "main" | "salad" | "soup" | ...
  difficulty?: string;          // "easy" | "medium" | "hard"
  maxTime?: number;             // Max cooking time in minutes
  minScore?: number;            // Min match score 0-100 (default: 0)
  onlyCookable?: boolean;       // Only recipes that can be cooked now
  excludeAllergens?: string[];  // ["gluten", "lactose", ...]
  dietTags?: string[];          // ["vegetarian", "vegan", "keto", ...]
}

// Example call
const response = await fetch(
  `https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/recipes/match?` +
  `testUserID=407582be-59d5-4d21-873b-1a72d31b0d42&` +
  `limit=10&` +
  `onlyCookable=false&` +
  `minScore=0`
);
```

**Response:**
```typescript
interface MatchedRecipe {
  // Basic info
  recipeId: string;              // UUID - use for cooking
  canonicalName: string;         // English name
  localName: string;             // Polish name
  country: string;               // Recipe origin
  category: string;              // "main" | "salad" | ...
  difficulty: string;            // "easy" | "medium" | "hard"
  timeMinutes: number;           // Total cooking time
  servings: number;              // Default servings
  
  // Matching metrics
  score: number;                 // 0-100 (higher = better match)
  coverage: number;              // 0-1 (0.75 = 75% ingredients available)
  canCookNow: boolean;           // true if all ingredients available
  
  // Economy (all in PLN, rounded to 2 decimals)
  usedValue: number;             // Cost of ingredients from fridge (what you'll use)
  costToComplete: number;        // Cost to buy missing ingredients
  totalRecipeCost: number;       // usedValue + costToComplete (full recipe cost)
  savedMoney: number;            // Same as usedValue (for UI: "Zaoszczędzisz")
  wasteRiskSaved: number;        // Value of expiring ingredients used
  
  // Ingredients
  usedIngredients: Array<{
    ingredientId: string;        // ⚠️ ALWAYS use this, NEVER match by name
    name: string;                // Display name (Polish)
    quantity: number;            // How much recipe needs
    unit: string;                // "g" | "ml" | "pcs" | ...
    available: number;           // How much user has in fridge
    isExpiringSoon?: boolean;    // true if expires within 2 days
  }>;
  
  missingIngredients: Array<{
    ingredientId: string;        // ⚠️ ALWAYS use this for shopping list
    name: string;                // Display name (Polish)
    quantity: number;            // How much to buy
    unit: string;                // "g" | "ml" | "pcs" | ...
    estimatedCost?: number;      // Estimated cost in PLN (if available)
  }>;
  
  // Additional info
  hasExpiringItems: boolean;     // true if any ingredient expires soon
  expiringItemsCount: number;    // Count of expiring ingredients
  allergens: string[];           // ["gluten", "lactose", "eggs", ...]
  dietTags: string[];            // ["vegetarian", "gluten-free", ...]
}

interface MatchResponse {
  success: true;
  data: {
    recipes: MatchedRecipe[];
    count: number;
  };
}
```

**Real Example Response:**
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
        "score": 77,
        "coverage": 0.75,
        "canCookNow": false,
        "usedValue": 4.95,
        "costToComplete": 0.9,
        "totalRecipeCost": 5.85,
        "savedMoney": 4.95,
        "wasteRiskSaved": 1.4,
        "hasExpiringItems": true,
        "expiringItemsCount": 1,
        "usedIngredients": [
          {
            "ingredientId": "fc57dbf2-39bb-4f30-a8e2-cf6585074587",
            "name": "Pomidor",
            "quantity": 400,
            "unit": "g",
            "available": 600
          },
          {
            "ingredientId": "59bf118a-9dae-4ca3-a262-776e18b58338",
            "name": "Ogórek",
            "quantity": 200,
            "unit": "g",
            "available": 800,
            "isExpiringSoon": true
          }
        ],
        "missingIngredients": [
          {
            "ingredientId": "1bf2b50b-1ebd-44c3-bf45-edbe2aca22c6",
            "name": "Oliwa z oliwek",
            "quantity": 30,
            "unit": "ml",
            "estimatedCost": 0.9
          }
        ],
        "allergens": ["lactose"],
        "dietTags": ["vegetarian", "gluten-free", "low-carb"]
      }
    ],
    "count": 1
  }
}
```

---

### 2. POST /api/recipes/{id}/cook - Cook Recipe

**Purpose:** Deduct ingredients from fridge, track economy, save cooking history

**Request:**
```typescript
interface CookRecipeRequest {
  servingsMultiplier?: number;  // Default: 1 (2 = double recipe)
  idempotencyKey?: string;      // UUID to prevent double-cooking
}

// Example call
const idempotencyKey = crypto.randomUUID(); // Generate on frontend!

const response = await fetch(
  `https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/recipes/${recipeId}/cook?` +
  `testUserID=${TEST_USER_ID}`,
  {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      servingsMultiplier: 1,
      idempotencyKey: idempotencyKey
    })
  }
);
```

**⚠️ CRITICAL: Idempotency**

Generate `idempotencyKey` on frontend **before** sending request:
```typescript
const [cookingKey, setCookingKey] = useState<string | null>(null);

const handleCookClick = () => {
  const key = crypto.randomUUID();
  setCookingKey(key); // Store until response
  
  cookRecipe(recipeId, { idempotencyKey: key })
    .then(() => setCookingKey(null))
    .catch(() => setCookingKey(null));
};

// If user clicks again while cooking - use SAME key
const effectiveKey = cookingKey || crypto.randomUUID();
```

**Response:**
```typescript
interface CookResponse {
  success: true;
  data: {
    cookLogId: string;           // UUID of cook event
    cookedAt: string;            // ISO timestamp
    servingsMultiplier: number;  // What was cooked
    
    // Economy snapshot
    usedValue: number;           // PLN spent from fridge
    wasteRiskSaved: number;      // PLN saved from expiring items
    totalRecipeCost: number;     // Total recipe cost
    
    // Ingredient deductions
    ingredientsUsed: Array<{
      ingredientId: string;      // ⚠️ ALWAYS present, use for fridge update
      name: string;              // Display name
      quantityUsed: number;      // How much was deducted
      unit: string;              // "g" | "ml" | "pcs"
      pricePerUnit: number;      // Price at cooking time
      totalCost: number;         // quantityUsed × pricePerUnit
      wasExpiringSoon: boolean;  // true if was expiring
      remainingInFridge: number; // ⚠️ New quantity (0 = removed from fridge)
    }>;
    
    remainingFridgeItems: number; // Total items left in fridge
  };
}
```

**Real Example Response:**
```json
{
  "success": true,
  "data": {
    "cookLogId": "af9474b5-bc61-4f5d-b596-4611d54b4e83",
    "cookedAt": "2025-12-21T19:02:35.428Z",
    "servingsMultiplier": 1,
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
        "remainingInFridge": 0
      },
      {
        "ingredientId": "59bf118a-9dae-4ca3-a262-776e18b58338",
        "name": "Ogórek",
        "quantityUsed": 200,
        "unit": "g",
        "pricePerUnit": 0.007,
        "totalCost": 1.4,
        "wasExpiringSoon": true,
        "remainingInFridge": 600
      }
    ],
    "remainingFridgeItems": 9
  }
}
```

**Error Responses:**
```typescript
// 400 Bad Request - Missing ingredients
{
  "error": "missing ingredients: [Pomidor, Oliwa z oliwek]"
}

// 400 Bad Request - Invalid servingsMultiplier
{
  "error": "servingsMultiplier must be positive"
}

// 200 OK - Already cooked (idempotency)
// Returns same response as original cook event
```

---

## 💻 TypeScript API Client

Create `lib/api/recipes.ts`:

```typescript
const BASE_URL = "https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app";
const TEST_USER_ID = "407582be-59d5-4d21-873b-1a72d31b0d42"; // ⚠️ MVP only

// ==================== TYPES ====================

export interface MatchedRecipe {
  recipeId: string;
  canonicalName: string;
  localName: string;
  country: string;
  category: string;
  difficulty: string;
  timeMinutes: number;
  servings: number;
  score: number;
  coverage: number;
  canCookNow: boolean;
  usedValue: number;
  costToComplete: number;
  totalRecipeCost: number;
  savedMoney: number;
  wasteRiskSaved: number;
  usedIngredients: Array<{
    ingredientId: string;
    name: string;
    quantity: number;
    unit: string;
    available: number;
    isExpiringSoon?: boolean;
  }>;
  missingIngredients: Array<{
    ingredientId: string;
    name: string;
    quantity: number;
    unit: string;
    estimatedCost?: number;
  }>;
  hasExpiringItems: boolean;
  expiringItemsCount: number;
  allergens: string[];
  dietTags: string[];
}

export interface CookedIngredient {
  ingredientId: string;
  name: string;
  quantityUsed: number;
  unit: string;
  pricePerUnit: number;
  totalCost: number;
  wasExpiringSoon: boolean;
  remainingInFridge: number; // ⚠️ 0 = removed from fridge
}

export interface CookResult {
  cookLogId: string;
  cookedAt: string;
  servingsMultiplier: number;
  usedValue: number;
  wasteRiskSaved: number;
  totalRecipeCost: number;
  ingredientsUsed: CookedIngredient[];
  remainingFridgeItems: number;
}

export interface RecipeMatchParams {
  limit?: number;
  country?: string;
  category?: string;
  difficulty?: string;
  maxTime?: number;
  minScore?: number;
  onlyCookable?: boolean;
  excludeAllergens?: string[];
  dietTags?: string[];
}

export interface CookRecipeParams {
  servingsMultiplier?: number;
  idempotencyKey?: string;
}

// ==================== API FUNCTIONS ====================

export async function getRecipeMatches(
  params: RecipeMatchParams = {}
): Promise<MatchedRecipe[]> {
  const queryParams = new URLSearchParams({
    testUserID: TEST_USER_ID, // ⚠️ MVP only
    limit: String(params.limit || 10),
  });

  if (params.country) queryParams.set("country", params.country);
  if (params.category) queryParams.set("category", params.category);
  if (params.difficulty) queryParams.set("difficulty", params.difficulty);
  if (params.maxTime) queryParams.set("maxTime", String(params.maxTime));
  if (params.minScore !== undefined) queryParams.set("minScore", String(params.minScore));
  if (params.onlyCookable) queryParams.set("onlyCookable", "true");
  if (params.excludeAllergens?.length) {
    params.excludeAllergens.forEach(a => queryParams.append("excludeAllergens", a));
  }
  if (params.dietTags?.length) {
    params.dietTags.forEach(t => queryParams.append("dietTags", t));
  }

  const response = await fetch(
    `${BASE_URL}/api/recipes/match?${queryParams.toString()}`
  );

  if (!response.ok) {
    throw new Error(`Failed to fetch recipes: ${response.status}`);
  }

  const data = await response.json();
  return data.data.recipes;
}

export async function cookRecipe(
  recipeId: string,
  params: CookRecipeParams = {}
): Promise<CookResult> {
  const queryParams = new URLSearchParams({
    testUserID: TEST_USER_ID, // ⚠️ MVP only
  });

  const response = await fetch(
    `${BASE_URL}/api/recipes/${recipeId}/cook?${queryParams.toString()}`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        servingsMultiplier: params.servingsMultiplier || 1,
        idempotencyKey: params.idempotencyKey || crypto.randomUUID(),
      }),
    }
  );

  if (!response.ok) {
    const error = await response.json().catch(() => ({ error: "Unknown error" }));
    throw new Error(error.error || `Failed to cook recipe: ${response.status}`);
  }

  const data = await response.json();
  return data.data;
}
```

---

## 🎨 React Component Example

Create `components/RecipeCard.tsx`:

```typescript
"use client";

import { useState } from "react";
import { MatchedRecipe, cookRecipe, CookResult } from "@/lib/api/recipes";
import { toast } from "sonner"; // or your toast library

interface RecipeCardProps {
  recipe: MatchedRecipe;
  onCookSuccess: (result: CookResult) => void;
}

export function RecipeCard({ recipe, onCookSuccess }: RecipeCardProps) {
  const [cooking, setCooking] = useState(false);
  const [cookKey, setCookKey] = useState<string | null>(null);

  const handleCook = async () => {
    if (cooking) return;

    // Generate idempotency key
    const key = cookKey || crypto.randomUUID();
    setCookKey(key);
    setCooking(true);

    try {
      const result = await cookRecipe(recipe.recipeId, {
        servingsMultiplier: 1,
        idempotencyKey: key,
      });

      // Show success toast
      toast.success("Gotowe!", {
        description: `Wykorzystano: ${result.usedValue.toFixed(2)} PLN\n` +
                     `Uratowano przed marnowaniem: ${result.wasteRiskSaved.toFixed(2)} PLN`,
      });

      // Update parent state (refresh matches, update fridge)
      onCookSuccess(result);
      
      setCookKey(null);
    } catch (error: any) {
      toast.error("Błąd podczas gotowania", {
        description: error.message,
      });
    } finally {
      setCooking(false);
    }
  };

  return (
    <div className="border rounded-lg p-4 space-y-3">
      {/* Header */}
      <div className="flex justify-between items-start">
        <div>
          <h3 className="font-bold text-lg">{recipe.localName}</h3>
          <p className="text-sm text-gray-500">
            {recipe.country} • {recipe.difficulty} • {recipe.timeMinutes} min
          </p>
        </div>
        <div className="text-right">
          <div className="text-2xl font-bold text-green-600">
            {recipe.score}%
          </div>
          <div className="text-xs text-gray-500">
            {Math.round(recipe.coverage * 100)}% składników
          </div>
        </div>
      </div>

      {/* Economy */}
      <div className="grid grid-cols-3 gap-2 text-sm">
        <div>
          <div className="text-gray-500">Z lodówki</div>
          <div className="font-semibold">{recipe.usedValue.toFixed(2)} PLN</div>
        </div>
        <div>
          <div className="text-gray-500">Do dokupienia</div>
          <div className="font-semibold">{recipe.costToComplete.toFixed(2)} PLN</div>
        </div>
        <div>
          <div className="text-gray-500">Całość</div>
          <div className="font-semibold">{recipe.totalRecipeCost.toFixed(2)} PLN</div>
        </div>
      </div>

      {/* Waste saving */}
      {recipe.wasteRiskSaved > 0 && (
        <div className="bg-amber-50 border border-amber-200 rounded p-2 text-sm">
          🌱 Uratowano {recipe.wasteRiskSaved.toFixed(2)} PLN przed marnowaniem
        </div>
      )}

      {/* Ingredients */}
      <div className="space-y-2">
        <div>
          <div className="font-medium text-sm mb-1">
            Z lodówki ({recipe.usedIngredients.length}):
          </div>
          <div className="text-xs space-y-1">
            {recipe.usedIngredients.map((ing) => (
              <div key={ing.ingredientId} className="flex justify-between">
                <span>
                  {ing.name} {ing.isExpiringSoon && "⏰"}
                </span>
                <span className="text-gray-500">
                  {ing.quantity} {ing.unit} / {ing.available} {ing.unit}
                </span>
              </div>
            ))}
          </div>
        </div>

        {recipe.missingIngredients.length > 0 && (
          <div>
            <div className="font-medium text-sm mb-1 text-orange-600">
              Do dokupienia ({recipe.missingIngredients.length}):
            </div>
            <div className="text-xs space-y-1">
              {recipe.missingIngredients.map((ing) => (
                <div key={ing.ingredientId} className="flex justify-between">
                  <span>{ing.name}</span>
                  <span className="text-gray-500">
                    {ing.quantity} {ing.unit}
                    {ing.estimatedCost && ` (~${ing.estimatedCost.toFixed(2)} PLN)`}
                  </span>
                </div>
              ))}
            </div>
          </div>
        )}
      </div>

      {/* Action button */}
      <button
        onClick={handleCook}
        disabled={cooking || !recipe.canCookNow}
        className={`w-full py-2 px-4 rounded font-medium ${
          recipe.canCookNow
            ? "bg-green-600 hover:bg-green-700 text-white"
            : "bg-gray-300 text-gray-600 cursor-not-allowed"
        }`}
      >
        {cooking
          ? "Gotuję..."
          : recipe.canCookNow
          ? "Gotuj"
          : "Dodaj brakujące do listy zakupów"}
      </button>
    </div>
  );
}
```

---

## 🔄 Update Fridge After Cooking

```typescript
function handleCookSuccess(result: CookResult) {
  // 1. Update local fridge state
  result.ingredientsUsed.forEach((ing) => {
    if (ing.remainingInFridge === 0) {
      // Remove from fridge UI
      removeFridgeItem(ing.ingredientId);
    } else {
      // Update quantity
      updateFridgeItem(ing.ingredientId, {
        quantity: ing.remainingInFridge,
      });
    }
  });

  // 2. Refresh recipe matches (new scores/coverage)
  refetchRecipeMatches();

  // 3. Show success message
  toast.success("Gotowe!", {
    description: `Wykorzystano: ${result.usedValue.toFixed(2)} PLN`,
  });
}
```

---

## ⚠️ Important Rules

### 1. ALWAYS Use ingredientId

```typescript
// ❌ WRONG - DO NOT match by name
const ingredient = fridge.find(item => item.name === "Pomidor");

// ✅ CORRECT - Use ingredientId
const ingredient = fridge.find(item => item.ingredientId === "fc57dbf2-39bb-4f30-a8e2-cf6585074587");
```

**Why:** Names can have typos, multi-language variants, or duplicates. IDs are unique.

### 2. Generate idempotencyKey on Frontend

```typescript
// ❌ WRONG - Backend generates (double-click creates duplicates)
cookRecipe(recipeId, {});

// ✅ CORRECT - Frontend generates and stores
const key = crypto.randomUUID();
cookRecipe(recipeId, { idempotencyKey: key });
```

### 3. Handle Idempotent Responses

If user clicks "Gotuj" twice with same key, backend returns **same response** (200 OK).

```typescript
try {
  const result = await cookRecipe(recipeId, { idempotencyKey: key });
  
  // Success - either new cook or idempotent duplicate
  onCookSuccess(result);
} catch (error) {
  // Real error (400 missing ingredients, 500 server error)
  toast.error(error.message);
}
```

### 4. Update Fridge Locally + Refetch

**Best practice:** Optimistic update + background revalidation

```typescript
// Immediate update (optimistic)
result.ingredientsUsed.forEach(ing => {
  updateFridgeItemLocally(ing.ingredientId, ing.remainingInFridge);
});

// Background refetch (revalidate)
setTimeout(() => refetchFridge(), 1000);
```

---

## 🧪 Testing Checklist

### Manual Testing

1. ✅ Open Assistant page
2. ✅ Click "Pokaż przepisy" → see 10 recipes
3. ✅ See correct economy values (usedValue, costToComplete, totalRecipeCost)
4. ✅ Click "Gotuj" on cookable recipe → toast shows success
5. ✅ Fridge quantities update immediately
6. ✅ Recipe list refreshes (new scores)
7. ✅ Double-click "Gotuj" → no duplicate cooking

### Edge Cases

- Missing ingredients → "Dodaj brakujące" button disabled
- All ingredients consumed → removed from fridge UI
- Partial consumption → quantity updated
- Expiring ingredients → wasteRiskSaved > 0

---

## 🚀 Next Steps (Post-MVP)

1. **Replace testUserID with JWT**
   - Get userId from auth middleware
   - Remove `testUserID` query parameter

2. **Move routes to protected**
   - `/api/recipes/match` → requires auth
   - `/api/recipes/{id}/cook` → requires auth

3. **Add missingCount/usedCount** (UI badges)
   ```typescript
   interface MatchedRecipe {
     missingCount: number;  // For "2 missing" badge
     usedCount: number;     // For "3/5 ingredients" badge
   }
   ```

4. **Shopping list integration**
   - "Dodaj brakujące do listy zakupów" button
   - POST /api/shopping-list/add

---

## 📞 Support

**Issues?** Check:
1. Backend URL correct: `yeasty-madelaine-fodi999-671ccdf5.koyeb.app`
2. testUserID included: `407582be-59d5-4d21-873b-1a72d31b0d42`
3. CORS enabled on backend (Koyeb should handle this)
4. Console errors in browser DevTools

**Test backend directly:**
```bash
curl "https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/recipes/match?testUserID=407582be-59d5-4d21-873b-1a72d31b0d42&limit=2"
```

---

✅ **Ready for integration!** Copy `lib/api/recipes.ts` and `RecipeCard.tsx` to your frontend repo.
