# 🚀 Recipe Catalog Migration Guide

## 📋 Prerequisites

✅ **DATABASE_URL verified**: `ep-soft-mud-agon8wu3-pooler.c-2.eu-central-1.aws.neon.tech/neondb`

✅ **Migrations ready**:
- `migrations/035_create_recipe_catalog.sql` (schema)
- `migrations/036_seed_real_recipes.sql` (6 recipes)

✅ **Verification script**: `verify_recipe_catalog.sql`

---

## 🔥 Step 1: Open Neon SQL Editor

1. Go to: https://console.neon.tech/
2. Select project: **ep-soft-mud-agon8wu3**
3. Click: **SQL Editor** (left sidebar)
4. Ensure you're connected to: `neondb` database

---

## 📦 Step 2: Apply Migration 035 (Schema)

### Copy-Paste into SQL Editor:

**File**: `migrations/035_create_recipe_catalog.sql`

```bash
# On your terminal, copy full migration:
cat migrations/035_create_recipe_catalog.sql | pbcopy
```

### Then in Neon SQL Editor:
1. **Paste** the entire migration
2. **Click "Run"**
3. **Wait** for success message

### Expected Output:
```
CREATE TABLE
CREATE INDEX
CREATE INDEX
CREATE INDEX
CREATE INDEX
(... repeat for all tables ...)
✅ Query executed successfully
```

### What This Creates:
- ✅ `Recipe` table (canonical recipes)
- ✅ `RecipeIngredient` table (ingredient requirements)
- ✅ `Allergen` table (14 EU allergens)
- ✅ `DietTag` table (11 diet classifications)
- ✅ `RecipeAllergen`, `RecipeDietTag` junction tables
- ✅ Indexes on country, category, difficulty, time

---

## 🌱 Step 3: Apply Migration 036 (Seeds)

### Copy-Paste into SQL Editor:

**File**: `migrations/036_seed_real_recipes.sql`

```bash
# Copy seed migration:
cat migrations/036_seed_real_recipes.sql | pbcopy
```

### Then in Neon SQL Editor:
1. **Paste** the entire seed file
2. **Click "Run"**
3. **Wait** (this is longer - 6 recipes with ingredients)

### Expected Output:
```
INSERT 0 14  (allergens)
INSERT 0 11  (diet tags)
INSERT 0 1   (recipe: Pierogi Ruskie)
INSERT 0 8   (ingredients)
INSERT 0 2   (allergens)
INSERT 0 1   (diet tags)
(... repeat for 6 recipes ...)
✅ Query executed successfully
```

### What This Seeds:
- ✅ 14 allergens (gluten, lactose, eggs, etc.)
- ✅ 11 diet tags (vegetarian, vegan, keto, etc.)
- ✅ 6 real recipes:
  1. **Pierogi Ruskie** (Poland, 90min, medium)
  2. **Bigos** (Poland, 180min, medium)
  3. **Spaghetti Carbonara** (Italy, 25min, easy)
  4. **Pizza Margherita** (Italy, 120min, medium)
  5. **Scrambled Eggs** (Poland, 10min, easy)
  6. **Greek Salad** (Greece, 15min, easy)
- ✅ ~40+ ingredient links
- ✅ ~15+ allergen associations
- ✅ ~10+ diet tag associations

---

## ✅ Step 4: Verify Everything

### Copy-Paste into SQL Editor:

**File**: `verify_recipe_catalog.sql`

```bash
# Copy verification script:
cat verify_recipe_catalog.sql | pbcopy
```

### Then in Neon SQL Editor:
1. **Paste** the entire verification script
2. **Click "Run"** (will run all 10 checks)
3. **Review results** below

---

## 🎯 Expected Results

### Check 1: Tables Exist
```
table_name
-----------------
Allergen
DietTag
Recipe
RecipeAllergen
RecipeDietTag
RecipeIngredient
```
✅ Should show **6 tables**

---

### Check 2: Recipe Count
```
recipes_count
-------------
6
```
✅ Should show **6 recipes**

---

### Check 3: Recipe Overview
```
canonicalName           | localName              | country | category | difficulty | timeMinutes
------------------------|------------------------|---------|----------|------------|------------
Greek Salad            | Grecka sałatka         | Greece  | salad    | easy       | 15
Pizza Margherita       | Pizza Margherita       | Italy   | main     | medium     | 120
Spaghetti Carbonara    | Spaghetti Carbonara    | Italy   | main     | easy       | 25
Bigos                  | Bigos                  | Poland  | main     | medium     | 180
Pierogi Ruskie         | Pierogi ruskie         | Poland  | main     | medium     | 90
Scrambled Eggs         | Jajecznica             | Poland  | main     | easy       | 10
```
✅ Should show recipes from **Poland, Italy, Greece**

---

### Check 4: Ingredient Links
```
recipe_ingredients_count
------------------------
42
```
✅ Should show **~40-50 links** (varies by recipe)

---

### Check 5: Allergens
```
key          | displayName         | iconEmoji
-------------|---------------------|----------
gluten       | Gluten             | 🌾
lactose      | Lactose (mleko)    | 🥛
eggs         | Jajka              | 🥚
fish         | Ryby               | 🐟
shellfish    | Skorupiaki         | 🦐
(... 14 total ...)
```
✅ Should show **14 allergens**

---

### Check 6: Diet Tags
```
key          | displayName        | description
-------------|--------------------|----------------------------------
vegetarian   | Vegetarian         | Bez mięsa i ryb
vegan        | Vegan              | Bez produktów zwierzęcych
gluten-free  | Gluten-free        | Bez glutenu
keto         | Keto               | Niskowęglowodanowa, wysoko tłuszczowa
(... 11 total ...)
```
✅ Should show **11 diet tags**

---

### Check 7: Carbonara Ingredients
```
recipe               | ingredient  | quantity | unit  | optional
---------------------|-------------|----------|-------|----------
Spaghetti Carbonara  | Makaron     | 400      | g     | false
Spaghetti Carbonara  | Boczek      | 200      | g     | false
Spaghetti Carbonara  | Jajko       | 4        | szt   | false
Spaghetti Carbonara  | Parmezan    | 100      | g     | false
Spaghetti Carbonara  | Czosnek     | 2        | ząbki | false
```
✅ Should show **5 ingredients** with quantities

---

### Check 8: Recipe Ingredient Counts
```
recipe              | country | time_min | difficulty | ingredients_count
--------------------|---------|----------|------------|------------------
Grecka sałatka      | Greece  | 15       | easy       | 7
Pizza Margherita    | Italy   | 120      | medium     | 6
Spaghetti Carbonara | Italy   | 25       | easy       | 5
Bigos               | Poland  | 180      | medium     | 10
Pierogi ruskie      | Poland  | 90       | medium     | 8
Jajecznica          | Poland  | 10       | easy       | 6
```
✅ Each recipe should have **5-10 ingredients**

---

### Check 9: Recipe Allergens
```
recipe               | allergen           | iconEmoji
---------------------|-------------------|----------
Spaghetti Carbonara  | Gluten            | 🌾
Spaghetti Carbonara  | Jajka             | 🥚
Spaghetti Carbonara  | Lactose (mleko)   | 🥛
Pizza Margherita     | Gluten            | 🌾
Pizza Margherita     | Lactose (mleko)   | 🥛
(... more ...)
```
✅ Each recipe should have **1-3 allergens**

---

### Check 10: Recipe Diet Tags
```
recipe              | diet_tag
--------------------|------------------
Grecka sałatka      | Gluten-free
Grecka sałatka      | Low-carb
Grecka sałatka      | Vegetarian
Jajecznica          | Gluten-free
Jajecznica          | Vegetarian
Pierogi ruskie      | Vegetarian
Pizza Margherita    | Vegetarian
(... more ...)
```
✅ Recipes should have **1-3 diet tags**

---

## 🚨 Troubleshooting

### ❌ Error: Table already exists
**Solution**: Tables already created. Skip to Step 3 (seeds) or Step 4 (verify).

### ❌ Error: Foreign key violation (ingredientId)
**Cause**: Ingredient not found in `Ingredient` table.
**Solution**: Check ingredient names in migration 036. May need to add missing ingredients first.

### ❌ Error: Duplicate key (canonicalName)
**Cause**: Recipe already seeded.
**Solution**: Already done! Skip to Step 4 (verify).

### ❌ Count is 0 (no recipes)
**Cause**: Seeds didn't run or failed silently.
**Solution**: 
1. Check Neon SQL Editor for errors
2. Re-run migration 036
3. Check ingredient catalog exists: `SELECT COUNT(*) FROM "Ingredient";`

---

## 🎉 Success Checklist

After verification, you should see:

- ✅ 6 tables exist
- ✅ 6 recipes in catalog
- ✅ 40+ recipe-ingredient links
- ✅ 14 allergens
- ✅ 11 diet tags
- ✅ Countries: Poland, Italy, Greece
- ✅ Carbonara has 5 ingredients
- ✅ Each recipe has allergens and diet tags

---

## 📊 Next Steps (After Verification)

1. **Register routes** in `cmd/server/main.go`:
   ```go
   r.Get("/api/recipes/match", recipeHandler.MatchRecipes)
   r.Post("/api/recipes/{id}/adapt", recipeHandler.AdaptRecipe)
   ```

2. **Test match endpoint**:
   ```bash
   curl -H "Authorization: Bearer $TOKEN" \
     "https://api.fodifood.com/api/recipes/match?country=Poland&minScore=50"
   ```

3. **Implement Groq client** for adaptation

4. **Frontend integration** with SWR hooks

---

## 📝 Notes

- **Migration 035**: ~184 lines (schema only)
- **Migration 036**: ~550+ lines (seeds with 6 recipes)
- **Verification**: 10 SQL checks
- **Estimated time**: 5-10 minutes total

**Database**: Neon.tech (PostgreSQL 15+)
**Module**: Recipe Catalog System
**Version**: 1.0 (MVP)

---

🚀 **Ready to apply migrations!**
