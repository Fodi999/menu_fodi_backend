# 🍽️ Recipe Servings Selection Feature

**Date**: 2026-01-03  
**Status**: ✅ IMPLEMENTED

---

## 📋 Overview

Backend now supports **flexible servings selection** when cooking recipes. Users can specify either:
1. **`servingsMultiplier`** (coefficient) - e.g., `2.0` = double the recipe
2. **`targetServings`** (absolute portions) - e.g., `2` = cook for 2 people

The system automatically calculates the correct multiplier based on the recipe's base servings.

---

## 🎯 Implementation Summary

### ✅ What Was Added

1. **New DTO Field**: `targetServings` in `CookRecipeRequest`
2. **Smart Calculator**: `GetMultiplier()` method that converts `targetServings` → `servingsMultiplier`
3. **DB Access**: Added `db *gorm.DB` to `RecipeHandler` for loading recipe metadata
4. **Validation**: Pre-loads recipe to get base servings before calculation

### ✅ What Was Already There

- ✅ `servingsMultiplier` column in `RecipeCookLog` table
- ✅ Ingredient quantity calculation: `requiredQty = baseQty * multiplier`
- ✅ Fridge deduction with correct amounts
- ✅ Cost calculation: `totalCost = usedAmount * pricePerUnit`
- ✅ Full cooking service logic

---

## 📡 API Contract

### Endpoint
```http
POST /api/recipes/{id}/cook
Authorization: Bearer <JWT_TOKEN>
Content-Type: application/json
```

### Request Body (Option 1: Coefficient)
```json
{
  "servingsMultiplier": 2.0,
  "idempotencyKey": "optional-uuid-v4",
  "force": false
}
```

**Result**: Cooks **2x** the recipe (doubles all ingredients)

---

### Request Body (Option 2: Absolute Portions)
```json
{
  "targetServings": 2,
  "idempotencyKey": "optional-uuid-v4",
  "force": false
}
```

**Result**: 
- Recipe has 4 base servings
- User wants 2 servings
- Backend calculates: `multiplier = 2 / 4 = 0.5`
- Cooks **50%** of the recipe (halves all ingredients)

---

### Request Body (Option 3: Default)
```json
{
  "idempotencyKey": "optional-uuid-v4",
  "force": false
}
```

**Result**: Cooks **1x** the recipe (default behavior)

---

## 🧮 Calculation Logic

```go
func (r *CookRecipeRequest) GetMultiplier(recipeServings int) float64 {
    // Priority 1: Explicit multiplier
    if r.ServingsMultiplier > 0 {
        return r.ServingsMultiplier
    }

    // Priority 2: Calculate from target servings
    if r.TargetServings > 0 && recipeServings > 0 {
        return float64(r.TargetServings) / float64(recipeServings)
    }

    // Priority 3: Default
    return 1.0
}
```

### Examples

| Recipe Base Servings | User Input | Multiplier | Result |
|---------------------|------------|------------|--------|
| 4 | `targetServings: 2` | 0.5 | Halves recipe |
| 4 | `targetServings: 8` | 2.0 | Doubles recipe |
| 1 | `targetServings: 3` | 3.0 | Triples recipe |
| 2 | `servingsMultiplier: 1.5` | 1.5 | 150% of recipe |
| 4 | *(empty)* | 1.0 | Original recipe |

---

## 🔧 Technical Implementation

### 1️⃣ DTO Changes

**File**: `internal/modules/recipes/dto/recipe_cook.go`

```go
type CookRecipeRequest struct {
    RecipeID           string  `json:"recipeId" binding:"required"`
    ServingsMultiplier float64 `json:"servingsMultiplier"` // Optional: coefficient (2.0 = double)
    TargetServings     int     `json:"targetServings"`     // Optional: absolute portions (2 = for 2 people)
    IdempotencyKey     string  `json:"idempotencyKey"`     
    Force              bool    `json:"force"`              
}

func (r *CookRecipeRequest) GetMultiplier(recipeServings int) float64 {
    // Smart calculation with priority: explicit > calculated > default
}
```

---

### 2️⃣ Handler Changes

**File**: `internal/modules/recipes/transport/http/handler.go`

**Added**:
- Import `github.com/google/uuid` and `gorm.io/gorm`
- Field `db *gorm.DB` to `RecipeHandler` struct
- Pre-load recipe to get base servings:

```go
// Load recipe to get base servings (needed for targetServings → multiplier conversion)
recipeUUID, err := uuid.Parse(recipeID)
// ... error handling ...

var recipe models.RecipeCatalog
if err := h.db.Where("id = ?", recipeUUID).First(&recipe).Error; err != nil {
    // ... error handling ...
}

// Calculate servings multiplier (supports both servingsMultiplier and targetServings)
servingsMultiplier := req.GetMultiplier(recipe.Servings)
```

**Updated**:
- `CookRecipe()` call uses calculated `servingsMultiplier` instead of `req.ServingsMultiplier`

---

### 3️⃣ Module Changes

**File**: `internal/modules/recipes/module.go`

```go
catalogHandler := httphandlers.NewRecipeHandler(
    db,  // ← Added DB parameter
    matchService,
    adapterService,
    cookService,
    sessionRepository,
    savedRecipeRepo,
    logger.Log,
)
```

---

## 📊 Database Impact

### RecipeCookLog Table
```sql
servingsMultiplier NUMERIC(10,2) NOT NULL DEFAULT 1.0
```

**Examples**:
```sql
-- User cooked recipe for 2 people (recipe base: 4 servings)
servingsMultiplier = 0.5

-- User doubled the recipe
servingsMultiplier = 2.0

-- User cooked default recipe
servingsMultiplier = 1.0
```

---

## 🎯 Use Cases

### Use Case 1: Single Person Cooking
```json
POST /api/recipes/e8ab233b-46dc-4a55-8e87-1c0dd3656790/cook
{
  "targetServings": 1
}
```

**Recipe**: Pierogi Ruskie (base: 4 servings)
- Multiplier: `1 / 4 = 0.25`
- Ingredients: 25% of original amounts
- Ziemniak: 500g → 125g ✅

---

### Use Case 2: Party Cooking
```json
POST /api/recipes/e8ab233b-46dc-4a55-8e87-1c0dd3656790/cook
{
  "targetServings": 12
}
```

**Recipe**: Pierogi Ruskie (base: 4 servings)
- Multiplier: `12 / 4 = 3.0`
- Ingredients: 300% of original amounts
- Ziemniak: 500g → 1500g ✅

---

### Use Case 3: Precise Control
```json
POST /api/recipes/e8ab233b-46dc-4a55-8e87-1c0dd3656790/cook
{
  "servingsMultiplier": 1.5
}
```

**Recipe**: Pierogi Ruskie (base: 4 servings)
- Multiplier: `1.5` (explicit)
- Effective servings: 4 * 1.5 = 6
- Ingredients: 150% of original amounts
- Ziemniak: 500g → 750g ✅

---

## ✅ Validation & Error Handling

### Valid Requests
```json
✅ { "targetServings": 2 }
✅ { "servingsMultiplier": 0.5 }
✅ { } // defaults to 1.0
✅ { "targetServings": 2, "servingsMultiplier": 1.5 } // servingsMultiplier takes priority
```

### Invalid Requests
```json
❌ { "targetServings": 0 } → defaults to 1.0
❌ { "targetServings": -1 } → defaults to 1.0
❌ { "servingsMultiplier": 0 } → defaults to 1.0
❌ { "servingsMultiplier": -0.5 } → defaults to 1.0
```

---

## 🧪 Testing Examples

### Test 1: Half Recipe
```bash
curl -X POST https://dima-fomin.pl/api/recipes/e8ab233b-46dc-4a55-8e87-1c0dd3656790/cook \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "targetServings": 2
  }'
```

**Expected Response**:
```json
{
  "success": true,
  "data": {
    "cookLogId": "uuid",
    "servingsMultiplier": 0.5,
    "ingredientsUsed": [
      {
        "name": "Ziemniak",
        "quantityUsed": 250.0,  // ← 500g * 0.5
        "unit": "g"
      }
    ]
  }
}
```

---

### Test 2: Double Recipe
```bash
curl -X POST https://dima-fomin.pl/api/recipes/e8ab233b-46dc-4a55-8e87-1c0dd3656790/cook \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "servingsMultiplier": 2.0
  }'
```

**Expected Response**:
```json
{
  "success": true,
  "data": {
    "cookLogId": "uuid",
    "servingsMultiplier": 2.0,
    "ingredientsUsed": [
      {
        "name": "Ziemniak",
        "quantityUsed": 1000.0,  // ← 500g * 2.0
        "unit": "g"
      }
    ]
  }
}
```

---

## 🔒 Security & Business Logic

### Idempotency
```json
{
  "targetServings": 2,
  "idempotencyKey": "550e8400-e29b-41d4-a716-446655440000"
}
```
- Prevents duplicate cooking if request is retried
- Same `idempotencyKey` → returns existing cook log

### Force Re-cooking
```json
{
  "targetServings": 2,
  "force": true,
  "idempotencyKey": "required-when-force-true"
}
```
- Allows cooking already-cooked recipes
- **Requires** `idempotencyKey` for safety

---

## 📈 Analytics Impact

### RecipeCookLog Insights
```sql
-- Average servings multiplier per user
SELECT 
    userId,
    AVG(servingsMultiplier) as avg_multiplier,
    COUNT(*) as total_cooks
FROM "RecipeCookLog"
GROUP BY userId;

-- Most popular serving sizes
SELECT 
    ROUND(servingsMultiplier, 1) as multiplier_rounded,
    COUNT(*) as cook_count
FROM "RecipeCookLog"
GROUP BY multiplier_rounded
ORDER BY cook_count DESC;
```

### Business Intelligence
- Users cooking for 1 person → suggest smaller recipes
- Users doubling recipes → offer bulk ingredient discounts
- Track portion waste (servingsMultiplier < 1.0 = less waste)

---

## 🚀 Future Enhancements

- [ ] **Smart Suggestions**: "Based on your fridge, we recommend cooking 2 servings"
- [ ] **Leftover Tracking**: Track if multiplier > 1.0 creates leftover portions
- [ ] **Meal Planning**: Schedule multi-day recipes with different serving sizes
- [ ] **Nutritional Scaling**: Auto-calculate calories per actual serving
- [ ] **Shopping List**: Generate based on `targetServings`

---

## 📝 Migration Status

✅ **No database migration required!**

The `servingsMultiplier` column already exists in `RecipeCookLog` table (created in migration 035).

This feature is **100% backward compatible** - existing functionality remains unchanged, just more flexible now.

---

## 🎉 Summary

### Before
```json
POST /api/recipes/{id}/cook
{ } // Always cooked 1x recipe
```

### After
```json
POST /api/recipes/{id}/cook
{ "targetServings": 2 }        // ← User-friendly
{ "servingsMultiplier": 0.5 }  // ← Power user
```

**Result**: Same backend logic, more flexible API! 🚀
