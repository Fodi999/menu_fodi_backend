# 🔧 PostgreSQL Prepared Statements Cache Issue

## Problem
```
ERROR: cached plan must not change result type (SQLSTATE 0A000)
```

## Root Cause
1. PostgreSQL uses **prepared statements** для оптимизации повторяющихся запросов
2. Мы добавили поле `canonical_id` в модель `Ingredient`
3. PostgreSQL закэшировал старый план запроса (без `canonical_id`)
4. При попытке выполнить запрос с новой структурой - ошибка

## What Happened
### Before
```go
type Ingredient struct {
    ID   string
    Name string
    // ... без canonical_id
}
```
PostgreSQL cached plan:
```sql
SELECT id, name, name_pl, ... FROM "Ingredient" WHERE ...
-- Returns 15 columns
```

### After Migration
```go
type Ingredient struct {
    ID          string
    Name        string
    CanonicalID *string  // NEW FIELD
    // ...
}
```
PostgreSQL tries to use cached plan:
```sql
SELECT id, name, name_pl, ... FROM "Ingredient" WHERE ...
-- But now Ingredient has 16 columns!
-- Cached plan expects 15 columns
-- ERROR: result type mismatch
```

## Solution: Force Redeploy
Koyeb will restart the application and PostgreSQL will clear prepared statements cache.

### Trigger Redeploy
```bash
# Create dummy file
echo "# Force redeploy" >> .koyeb_deploy
git add .koyeb_deploy
git commit -m "chore: force redeploy to clear PostgreSQL cache"
git push
```

### Alternative Solutions (if redeploy doesn't work)
```sql
-- Option 1: Clear all prepared statements on database
DEALLOCATE ALL;

-- Option 2: Disable prepared statements in GORM (NOT RECOMMENDED for production)
db.Config.PrepareStmt = false
```

## Logs Before Fix
```
✅ [FRIDGE CHECK] Found 13 ingredients
📦 [FRIDGE CHECK] Canonical group: vegetable_oil (ingredient_id: 1b7cea8e...)
📊 [FRIDGE CHECK] Total keys in fridgeSet: 16

❌ ERROR: cached plan must not change result type (SQLSTATE 0A000)
   SELECT * FROM "Ingredient" WHERE "Ingredient"."id" IN (...)
```

## Expected Logs After Fix
```
✅ [FRIDGE CHECK] Found 13 ingredients
📦 [FRIDGE CHECK] Canonical group: vegetable_oil (ingredient_id: 1b7cea8e...)
📊 [FRIDGE CHECK] Total keys in fridgeSet: 16

✅ [GET SINGLE RECIPE] Recipe found: zharenye_yaytsa (3 ingredients)
🎯 [CANONICAL MATCH] Recipe needs 'Растительное масло', matched via canonical_id='vegetable_oil'
✅ [GET SINGLE RECIPE] DTO built: 3 available, 0 missing, 100.00% match
```

## Prevention
When adding/removing columns from existing tables:
1. ✅ Run migration first (ALTER TABLE)
2. ✅ Update Go models
3. ✅ **RESTART APPLICATION** to clear prepared statements cache
4. ❌ Don't skip step 3!

## Related Issues
- PostgreSQL Prepared Statements: https://www.postgresql.org/docs/current/sql-prepare.html
- GORM PrepareStmt: https://gorm.io/docs/performance.html#Caches-Prepared-Statement

## Status
- Migration applied: ✅
- Model updated: ✅
- Redeploy triggered: ✅
- Waiting for Koyeb: ⏳ (~2 minutes)
