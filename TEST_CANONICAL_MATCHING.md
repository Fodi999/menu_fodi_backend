# 🧪 Test Plan: Canonical Ingredient Matching

## Test Case: Vegetable Oil Matching

### Before Fix ❌
```bash
# User has: "Olej roślinny" (1000ml) in fridge
# Recipe needs: "Olej rzepakowy" (30ml)
# Result: missing_ingredients = ["Растительное масло"]
# Match: 67% (2/3 ingredients)
```

### After Fix ✅
```bash
# Same scenario
# Expected: available_ingredients = ["Яйца", "Соль", "Растительное масло"]
# Expected: missing_ingredients = []
# Expected: match_percent = 100%
```

## How to Test

### 1. Wait for Koyeb deployment (~2 minutes)
Check: https://app.koyeb.com/

### 2. Get fresh token
```bash
TOKEN=$(curl -X POST "https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"fodi85@gmail.ru","password":"210185"}' | jq -r '.token')
```

### 3. Test the endpoint
```bash
curl -X GET "https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/recipe-recommendations/zharenye_yaytsa?lang=ru" \
  -H "Authorization: Bearer $TOKEN" | jq '{
    match_percent,
    match_status,
    available: [.available_ingredients[].display_name],
    missing: [.missing_ingredients[].display_name]
  }'
```

### Expected Output
```json
{
  "match_percent": 100,
  "match_status": "ready",
  "available": [
    "Яйца",
    "Соль",
    "Растительное масло"  ← ДОЛЖНО БЫТЬ ТУТ!
  ],
  "missing": []  ← ДОЛЖНО БЫТЬ ПУСТО!
}
```

### 4. Check Koyeb Logs
Look for debug messages:
```
🔍 [FRIDGE CHECK] Starting for userID: 407582be...
📦 [FRIDGE CHECK] Canonical group: vegetable_oil (ingredient_id: 1b7cea8e...)
📊 [FRIDGE CHECK] Total keys in fridgeSet: 14 (includes canonical groups)
🎯 [CANONICAL MATCH] Recipe needs 'Растительное масло', matched via canonical_id='vegetable_oil'
✅ [GET SINGLE RECIPE] DTO built: 3 available, 0 missing, 100.00% match
```

## What Changed

### Database
```sql
-- Added canonical_id column
SELECT id, name_en, canonical_id FROM "Ingredient" 
WHERE canonical_id = 'vegetable_oil';

-- Result:
-- 1b7cea8e... | Vegetable oil | vegetable_oil  (в холодильнике)
-- 9ff773d2... | Vegetable oil | vegetable_oil  (в рецепте)
-- + еще 1 вариант
```

### Matching Logic
```go
// BEFORE:
if fridgeIngredientIDs[ingredient.ID] { ... }  // Only exact ID match

// AFTER:
if fridgeIngredientIDs[ingredient.ID] { ... }  // Exact ID match
OR fridgeIngredientIDs[ingredient.CanonicalID] { ... }  // Canonical group match
```

## Other Test Cases

### Test 2: Salt Matching
```bash
# User has: "Sól" (generic salt)
# Recipe needs: "Sól morska" (sea salt)
# Both have canonical_id='salt'
# Should MATCH ✅
```

### Test 3: Eggs Matching
```bash
# User has: "Jaja" (chicken eggs)
# Recipe needs: "Jaja przepiórcze" (quail eggs)
# Both have canonical_id='eggs'
# Should MATCH ✅
```

## Rollback Plan
If something goes wrong:
```bash
# Revert canonical_id column
ALTER TABLE "Ingredient" DROP COLUMN IF EXISTS canonical_id;

# Redeploy previous version
git revert HEAD
git push
```

## Success Criteria
- ✅ Recipe "zharenye_yaytsa" shows 100% match (was 67%)
- ✅ "Растительное масло" in available_ingredients (was in missing)
- ✅ Debug logs show canonical group matching
- ✅ No errors in Koyeb logs
- ✅ No breaking changes for other recipes
