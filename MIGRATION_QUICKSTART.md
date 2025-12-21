# 🎯 Quick Start: Apply Recipe Migrations

## ✅ Prerequisites (CONFIRMED)

- **Database**: `ep-soft-mud-agon8wu3-pooler.c-2.eu-central-1.aws.neon.tech/neondb`
- **Server config**: `DATABASE_URL` in `.env` ✅
- **Migrations ready**: `035_create_recipe_catalog.sql`, `036_seed_real_recipes.sql` ✅
- **Helper script**: `./apply_migrations_quick.sh` ✅

---

## 🚀 Option A: Quick Script (Recommended)

```bash
# Run helper script
./apply_migrations_quick.sh

# Then select:
# 1) Migration 035 (Schema) → Run in Neon
# 2) Migration 036 (Seeds) → Run in Neon
# 3) Verification Script → Run in Neon
```

---

## 🚀 Option B: Manual Copy-Paste

### Step 1: Apply Schema (Migration 035)

```bash
# Copy to clipboard
cat migrations/035_create_recipe_catalog.sql | pbcopy
```

Then:
1. Open: https://console.neon.tech/
2. Go to: **SQL Editor**
3. **Paste** and **Run**
4. ✅ Should see: `CREATE TABLE`, `CREATE INDEX` messages

---

### Step 2: Apply Seeds (Migration 036)

```bash
# Copy to clipboard
cat migrations/036_seed_real_recipes.sql | pbcopy
```

Then:
1. In **SQL Editor** (same as Step 1)
2. **Paste** and **Run**
3. ✅ Should see: `INSERT 0 14`, `INSERT 0 11`, `INSERT 0 1` messages

---

### Step 3: Verify Everything

```bash
# Copy verification script
cat verify_recipe_catalog.sql | pbcopy
```

Then:
1. In **SQL Editor**
2. **Paste** and **Run**
3. ✅ Review all 10 checks

---

## 🎯 Expected Results

After Step 3, you should see:

| Check | Expected Result | Status |
|-------|----------------|--------|
| Tables exist | 6 tables (Recipe, RecipeIngredient, Allergen, DietTag, + junctions) | ⏳ |
| Recipe count | 6 recipes | ⏳ |
| Ingredient links | ~40-50 links | ⏳ |
| Allergens | 14 allergens (gluten, lactose, eggs, etc.) | ⏳ |
| Diet tags | 11 tags (vegetarian, vegan, keto, etc.) | ⏳ |
| Countries | Poland, Italy, Greece | ⏳ |
| Carbonara | 5 ingredients (Makaron, Boczek, Jajko, Parmezan, Czosnek) | ⏳ |

---

## 🔥 Verification Commands (Quick Copy)

```sql
-- Quick check: Tables exist?
SELECT table_name FROM information_schema.tables 
WHERE table_schema = 'public' 
  AND table_name IN ('Recipe', 'RecipeIngredient', 'Allergen', 'DietTag')
ORDER BY table_name;

-- Quick check: Recipe count?
SELECT COUNT(*) AS recipes_count FROM "Recipe";

-- Quick check: Ingredient links?
SELECT COUNT(*) AS links_count FROM "RecipeIngredient";

-- Quick check: Carbonara ingredients?
SELECT r."localName", i.name, ri.quantity, ri.unit
FROM "Recipe" r
JOIN "RecipeIngredient" ri ON ri."recipeId" = r.id
JOIN "Ingredient" i ON i.id = ri."ingredientId"
WHERE r."canonicalName" = 'Spaghetti Carbonara';
```

---

## 📋 Manual Checklist

After applying migrations, mark these as done:

- [ ] Migration 035 applied (schema)
- [ ] Migration 036 applied (seeds)
- [ ] Tables exist (6 tables)
- [ ] Recipe count = 6
- [ ] Ingredient links > 0
- [ ] Allergens = 14
- [ ] Diet tags = 11
- [ ] Countries include PL, IT, GR
- [ ] Carbonara has 5 ingredients

---

## 🚨 Common Issues

### ❌ "Table already exists"
**Solution**: Already applied! Skip to verification.

### ❌ "Foreign key violation"
**Cause**: Ingredient missing from catalog.
**Check**: `SELECT name FROM "Ingredient" WHERE name IN ('Makaron', 'Boczek', 'Jajko');`

### ❌ Count = 0
**Cause**: Seeds didn't run.
**Solution**: Re-run migration 036.

---

## 📖 Full Documentation

- **Detailed guide**: `APPLY_RECIPE_MIGRATIONS.md`
- **Verification script**: `verify_recipe_catalog.sql`
- **Helper script**: `./apply_migrations_quick.sh`

---

## ⏭️ Next Steps (After Verification)

1. Register routes in `main.go`
2. Test `GET /api/recipes/match`
3. Implement Groq client for adaptation
4. Frontend integration

---

🚀 **Ready to go! Start with Option A or B above.**
