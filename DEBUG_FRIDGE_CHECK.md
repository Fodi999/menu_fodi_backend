# Debug Plan: Fridge Check Verification

## 🎯 Goal

Verify that `user_fridge_items` table is queried when calling `/api/recipe-recommendations/{id}`

---

## 🔍 What to Check in Logs

### Expected Logs (if working correctly):

```
🎯 [GET SINGLE RECIPE] Request: userID=407582be-..., recipeID=zharenye_yaytsa, lang=ru
📦 [GET SINGLE RECIPE] Step 1: Getting fridge for user 407582be-...
🔍 [FRIDGE CHECK] Starting for userID: 407582be-...

[SQL QUERY] SELECT DISTINCT ingredient_id FROM user_fridge_items WHERE user_id = '407582be-...' AND quantity > 0

✅ [FRIDGE CHECK] Found 13 ingredients in fridge for user 407582be-...
📦 [FRIDGE CHECK] Ingredient IDs: [3260aadf-..., c4d477f8-..., ...]
✅ [GET SINGLE RECIPE] Fridge loaded: 13 ingredients
🍳 [GET SINGLE RECIPE] Step 2: Getting recipe zharenye_yaytsa

[SQL QUERY] SELECT * FROM "Recipe" WHERE "canonicalName" = 'zharenye_yaytsa'
[SQL QUERY] SELECT * FROM "Ingredient" WHERE id IN (...)

✅ [GET SINGLE RECIPE] Recipe found: zharenye_yaytsa (3 ingredients)
🔨 [GET SINGLE RECIPE] Step 3: Building DTO with fridge check
✅ [GET SINGLE RECIPE] DTO built: 2 available, 1 missing, 66.67% match
```

### If fridge check is NOT working:

```
❌ Missing: 🔍 [FRIDGE CHECK] Starting for userID...
❌ Missing: SELECT DISTINCT ingredient_id FROM user_fridge_items
❌ Result: 0 available, 3 missing (wrong!)
```

---

## 📝 Test Steps

### Step 1: Wait for Koyeb Deployment

Commit `fc36a60` is deploying...

### Step 2: Get Fresh Token

```bash
curl -X POST "https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"fodi85@gmail.ru","password":"210185"}'
```

### Step 3: Test Endpoint

```bash
TOKEN="..." # from Step 2

curl -X GET "https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/recipe-recommendations/zharenye_yaytsa?lang=ru" \
  -H "Authorization: Bearer $TOKEN"
```

### Step 4: Check Koyeb Logs

Look for our debug logs in Koyeb console:
- `🔍 [FRIDGE CHECK]` logs
- `📦 [GET SINGLE RECIPE]` logs
- SQL queries to `user_fridge_items`

---

## 🎯 Expected Results

### Scenario A: Fridge Check Working ✅

**Logs show:**
```
🔍 [FRIDGE CHECK] Starting for userID: 407582be-...
✅ [FRIDGE CHECK] Found 13 ingredients
SELECT DISTINCT ingredient_id FROM user_fridge_items WHERE user_id = '407582be-...'
```

**Response:**
```json
{
  "available_ingredients": [
    {"display_name": "Яйца"},
    {"display_name": "Соль"}
  ],
  "missing_ingredients": [
    {"display_name": "Растительное масло"}
  ],
  "match_percent": 66.67
}
```

### Scenario B: Fridge Check NOT Working ❌

**Logs show:**
```
❌ No 🔍 [FRIDGE CHECK] logs
❌ No SQL to user_fridge_items
```

**Response:**
```json
{
  "available_ingredients": [],
  "missing_ingredients": [
    {"display_name": "Яйца"},
    {"display_name": "Соль"},
    {"display_name": "Растительное масло"}
  ],
  "match_percent": 0
}
```

---

## 🐛 Possible Issues

### Issue 1: Service Not Called

**Symptom**: No logs at all  
**Cause**: Handler not calling service  
**Fix**: Check handler code

### Issue 2: Empty UserID

**Symptom**: Logs show `userID: ""`  
**Cause**: Middleware not setting userID  
**Fix**: Check auth middleware

### Issue 3: Wrong Table Name

**Symptom**: SQL error "table not found"  
**Cause**: PostgreSQL table name mismatch  
**Fix**: Check actual table name in Neon

### Issue 4: Wrong Column Name

**Symptom**: SQL error "column not found"  
**Cause**: Column name case sensitivity  
**Fix**: Quote column names

---

## 📊 User's Fridge (Expected Data)

User `407582be-59d5-4d21-873b-1a72d31b0d42` should have ~13 ingredients:
- Eggs (Яйца) - `3260aadf-52de-4038-9568-ee536495224a`
- Salt (Соль) - `c4d477f8-9123-4175-b515-5201ee1ff61b`
- ... (11 more)

Recipe `zharenye_yaytsa` requires:
1. Eggs ✅ (in fridge)
2. Salt ✅ (in fridge)
3. Vegetable Oil ❌ (NOT in fridge)

**Expected Result**: 2 available, 1 missing, 66.67% match

---

## ⏰ Next Steps

1. ⏳ Wait for Koyeb deployment (~2 min)
2. 🧪 Test endpoint with logging
3. 📋 Check Koyeb logs for debug output
4. ✅ Verify fridge query executes
5. 🎯 Confirm match calculation correct

---

**Status**: ⏳ Waiting for deployment (commit `fc36a60`)
