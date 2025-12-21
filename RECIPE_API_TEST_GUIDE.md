# 🎉 Recipe Match API - Testing Summary

## ✅ What's Done

### 1. Routes Registered ✅
- `GET /api/recipes/match` - Recipe matching endpoint
- `POST /api/recipes/{id}/adapt` - AI adaptation endpoint (requires auth)
- Routes properly initialized in `internal/modules/recipes/module.go`

### 2. Server Running ✅
- Server started on port **8083**
- Dev mode enabled with `testUserID` parameter for testing without auth
- Endpoint responds successfully: `{"success": true}`

### 3. Initial Test Results ✅
```bash
curl "http://localhost:8083/api/recipes/match?testUserID=test-user-123"
# Response:
{
  "success": true,
  "data": {
    "recipes": [],
    "count": 0
  }
}
```

**Why empty?** Test user has no fridge items → Match service returns empty array.

---

## 📋 Next Steps to Test Properly

### Option A: Use Real User with Fridge (Recommended)

1. **Find real user ID** (Dmitrij Fomin):
   ```sql
   SELECT id FROM "User" WHERE name LIKE '%Dmitrij%' OR email LIKE '%dmitrij%' LIMIT 1;
   ```

2. **Check if user has fridge items**:
   ```sql
   SELECT COUNT(*) FROM user_fridge_items WHERE user_id = '<user_id>';
   ```

3. **Test with real user**:
   ```bash
   curl "http://localhost:8083/api/recipes/match?testUserID=<real_user_id>"
   ```

---

### Option B: Add Test Fridge Items

```sql
-- Insert test fridge items for test-user-123
INSERT INTO user_fridge_items (user_id, ingredient_id, quantity, unit, purchase_date, expiry_date)
VALUES 
  ('test-user-123', (SELECT id FROM "Ingredient" WHERE name = 'Makaron (spaghetti)'), 500, 'g', NOW(), NOW() + INTERVAL '30 days'),
  ('test-user-123', (SELECT id FROM "Ingredient" WHERE name = 'Jaja'), 6, 'pcs', NOW(), NOW() + INTERVAL '14 days'),
  ('test-user-123', (SELECT id FROM "Ingredient" WHERE name = 'Parmezan'), 150, 'g', NOW(), NOW() + INTERVAL '60 days'),
  ('test-user-123', (SELECT id FROM "Ingredient" WHERE name = 'Boczek'), 200, 'g', NOW(), NOW() + INTERVAL '21 days');
```

Then test:
```bash
curl "http://localhost:8083/api/recipes/match?testUserID=test-user-123"
```

**Expected**: Spaghetti Carbonara with ~80-100% match score.

---

### Option C: Modify Service (Show All Recipes)

Change `match_service.go` to return all recipes even with empty fridge (score=0):

```go
// In MatchRecipesWithFridge(), replace:
if len(fridgeItems) == 0 {
    return []RecipeMatch{}, nil
}

// With:
if len(fridgeItems) == 0 {
    // Return all recipes with score=0
    recipes, err := s.loadRecipesWithFilters(filters)
    if err != nil {
        return nil, fmt.Errorf("failed to load recipes: %w", err)
    }
    
    matches := make([]RecipeMatch, 0, len(recipes))
    for _, recipe := range recipes {
        match := RecipeMatch{
            RecipeID:       recipe.ID,
            CanonicalName:  recipe.CanonicalName,
            LocalName:      recipe.LocalName,
            Country:        recipe.Country,
            Difficulty:     recipe.Difficulty,
            TimeMinutes:    recipe.TimeMinutes,
            MatchScore:     0,
            Coverage:       0,
            CanCookNow:     false,
            // ... rest with empty arrays
        }
        matches = append(matches, match)
    }
    
    if filters.Limit > 0 && len(matches) > filters.Limit {
        matches = matches[:filters.Limit]
    }
    
    return matches, nil
}
```

---

## 🧪 Test Commands (When Data Ready)

### Test 1: No Filters (All Recipes)
```bash
curl -s "http://localhost:8083/api/recipes/match?testUserID=<user_id>" | jq .
```

**Expected**: 6 recipes with match scores

---

### Test 2: Filter by Country
```bash
curl -s "http://localhost:8083/api/recipes/match?country=Poland&testUserID=<user_id>" | jq .
```

**Expected**: 3 Polish recipes (Pierogi, Bigos, Jajecznica)

---

### Test 3: Filter by Time
```bash
curl -s "http://localhost:8083/api/recipes/match?maxTime=30&testUserID=<user_id>" | jq .
```

**Expected**: 3 recipes (Jajecznica 10min, Greek Salad 15min, Carbonara 25min)

---

### Test 4: Multiple Filters
```bash
curl -s "http://localhost:8083/api/recipes/match?country=Poland&maxTime=30&difficulty=easy&testUserID=<user_id>" | jq .
```

**Expected**: 1 recipe (Jajecznica: Poland, 10min, easy)

---

### Test 5: Check Response Structure
```bash
curl -s "http://localhost:8083/api/recipes/match?limit=1&testUserID=<user_id>" | jq '.data.recipes[0]'
```

**Expected Fields**:
```json
{
  "recipeId": "uuid",
  "canonicalName": "Spaghetti Carbonara",
  "localName": "Spaghetti alla Carbonara",
  "country": "Italy",
  "difficulty": "easy",
  "timeMinutes": 25,
  "servings": 4,
  "matchScore": 85.5,
  "coverage": 0.75,
  "canCookNow": false,
  "usedIngredients": [...],
  "missingIngredients": [...],
  "costToComplete": 12.50,
  "hasExpiringItems": true,
  "allergens": [...],
  "dietTags": [...]
}
```

---

## 🚨 Current Limitations

1. **No Auth**: Using `testUserID` parameter (DEV ONLY)
   - ⚠️ Remove before production!
   - Real endpoint should use JWT from `Authorization: Bearer $TOKEN`

2. **Empty Fridge**: Test user has no fridge items
   - Add test data OR use real user

3. **AI Adapter Not Implemented**: Groq client not integrated yet
   - `/api/recipes/:id/adapt` will fail until Groq client added

---

## 🎯 Recommended Flow

**Quick Test** (5 minutes):
1. Use Option B (Add test fridge items)
2. Run Test 1-4 above
3. Verify response structure

**Full Test** (20 minutes):
1. Find real user with fridge
2. Test all filters
3. Test with different fridge contents
4. Measure performance (<150ms target)
5. Test edge cases (empty fridge, no matches)

---

## 📊 Expected Performance

- **Fridge Load**: 10-20ms
- **Recipe Load**: 20-30ms
- **Matching**: 10-20ms
- **Total**: 50-100ms

---

## 🔗 Server Info

- **URL**: http://localhost:8083
- **Endpoint**: `/api/recipes/match`
- **Method**: GET
- **Auth**: `?testUserID=<user_id>` (dev mode)
- **Logs**: `/tmp/server_8083.log`

---

## ⏭️ After Testing

1. **Remove dev mode** (`testUserID` parameter)
2. **Add real auth** (JWT middleware)
3. **Implement Groq client** for `/adapt` endpoint
4. **Deploy** to production
5. **Frontend integration** with SWR hooks

---

🚀 **Ready to test with real data!**
