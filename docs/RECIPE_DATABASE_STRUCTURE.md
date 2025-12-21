# Recipe Database Structure - Current vs Canonical
**Date:** 2025-12-21  
**Status:** ⚠️ Mostly Aligned, Minor Gaps

---

## 1️⃣ Table: Recipe

### Current Structure ✅
```sql
Recipe
------
id              UUID (PK)           ✅
canonicalName   VARCHAR(255) UNIQUE ✅
localName       VARCHAR(255)        ✅
country         VARCHAR(100)        ✅
region          VARCHAR(100)        ℹ️ Extra field
category        VARCHAR(50)         ✅ (CHECK constraint: 7 values)
difficulty      VARCHAR(20)         ✅ (CHECK constraint: easy/medium/hard)
timeMinutes     INT                 ✅
servings        INT DEFAULT 4       ✅
steps           JSONB               ✅ (array of strings)
nutritionProfile JSONB              ℹ️ Extra field
source          JSONB               ℹ️ Extra field
createdAt       TIMESTAMP           ✅
updatedAt       TIMESTAMP           ℹ️ Extra field
```

### Missing Fields ❌
- `description` TEXT - Recipe description for UI
- `imageUrl` TEXT NULL - Recipe image

### Extra Fields (OK, не мешают)
- `region` - Geographic region detail
- `nutritionProfile` - Calories, protein, etc.
- `source` - Recipe source metadata
- `updatedAt` - Last update timestamp

### Status: ✅ **90% Aligned**

---

## 2️⃣ Table: RecipeIngredient

### Current Structure ✅
```sql
RecipeIngredient
----------------
id              UUID (PK)           ✅
recipeId        UUID (FK)           ✅
ingredientId    TEXT (FK)           ✅ (should be UUID, but TEXT works)
ingredientKey   VARCHAR(255)        ℹ️ Extra field (legacy)
quantity        NUMERIC(10,2)       ✅
unit            VARCHAR(50)         ✅
optional        BOOLEAN DEFAULT false ✅
sortOrder       INT DEFAULT 0       ✅
createdAt       TIMESTAMP           ℹ️ Extra field
```

### Missing Fields ❌
- `groupName` TEXT NULL - For grouping ("baza", "sos", "dodatki")

### Extra Fields (OK)
- `ingredientKey` - Legacy key for compatibility
- `createdAt` - Creation timestamp

### Status: ✅ **85% Aligned** (missing groupName)

---

## 3️⃣ Table: RecipeStep

### Current Structure ❌
**Table does NOT exist**

Currently, steps are stored in `Recipe.steps` as JSONB array:
```json
["1. Rozbij mięso", "2. Obtocz w jajku", ...]
```

### Canonical Structure 📋
```sql
RecipeStep
----------
id          UUID (PK)
recipeId    UUID (FK)
stepNumber  INT
content     TEXT
```

### Benefits of Normalized Structure
1. **Translation** - Easy to add `locale` column for multi-language
2. **Editing** - Update single step without touching entire array
3. **Versioning** - Track step changes over time
4. **Validation** - Enforce step ordering with constraints

### Status: ❌ **Not Implemented** (but current JSONB works for MVP)

---

## 4️⃣ API Response Format

### Current Format (from handler)
```json
{
  "success": true,
  "data": {
    "recipe": {
      "id": "uuid",
      "canonicalName": "Polish Breaded Pork Chop",
      "localName": "Kotlet schabowy",
      "country": "Poland",
      "category": "main",
      "difficulty": "easy",
      "timeMinutes": 25,
      "servings": 4,              // ✅ Matches servingsDefault
      "steps": ["...", "..."]     // ✅ Array of strings
    },
    "match": {
      "canCookNow": true,
      "usedIngredients": [...],
      "missingRequired": [...]
    },
    "economy": {
      "usedFromFridge": 4.95,
      "saved": 4.95
    }
  }
}
```

### Canonical Format (target)
```json
{
  "recipe": {
    "id": "uuid",
    "localName": "Kotlet schabowy",
    "description": "Klasyczny kotlet schabowy panierowany",  // ❌ Missing in DB
    "country": "Poland",
    "category": "main",
    "difficulty": "easy",
    "timeMinutes": 25,
    "servingsDefault": 4,        // ✅ Same as servings
    "imageUrl": null,            // ❌ Missing in DB
    "steps": ["...", "..."]      // ✅ Works
  },
  "ingredients": [               // ✅ Can derive from RecipeIngredient
    {
      "ingredientId": "uuid",
      "name": "Olej roślinny",
      "quantity": 50,
      "unit": "ml",
      "optional": false,
      "group": "baza"            // ❌ Missing: groupName field
    }
  ]
}
```

### Status: ✅ **80% Aligned**
- Missing: `description`, `imageUrl`, `groupName`
- Rename needed: `servings` → `servingsDefault` (just API label)

---

## 5️⃣ Gaps Summary

### Critical (Blocks Features) 🔴
None! Current structure supports all MVP features.

### Important (Improves UX) 🟡
1. **Add `description` to Recipe** - For recipe details page
2. **Add `imageUrl` to Recipe** - For recipe cards
3. **Add `groupName` to RecipeIngredient** - For ingredient grouping in UI

### Nice to Have (Future) 🟢
1. **Create `RecipeStep` table** - Better for translations/editing
2. **Add translations support** - `RecipeTranslation` table
3. **Add `nutritionProfile` properly** - Standardized format

---

## 6️⃣ Migration Plan

### Phase 1: Add Missing Columns (Quick, non-breaking)
```sql
-- Add description and imageUrl to Recipe
ALTER TABLE "Recipe" 
ADD COLUMN description TEXT,
ADD COLUMN "imageUrl" TEXT;

-- Add groupName to RecipeIngredient
ALTER TABLE "RecipeIngredient"
ADD COLUMN "groupName" VARCHAR(50);
```

**Impact:** None, all nullable, backward compatible

### Phase 2: Populate Data (Manual work)
```sql
-- Add descriptions for popular recipes
UPDATE "Recipe" 
SET description = 'Klasyczny polski kotlet schabowy panierowany w jajku i bułce tartej'
WHERE "localName" = 'Kotlet schabowy';

-- Group ingredients
UPDATE "RecipeIngredient" ri
SET "groupName" = 'baza'
FROM "Recipe" r
WHERE ri."recipeId" = r.id 
AND r."localName" = 'Kotlet schabowy'
AND ri."ingredientId" IN (
  SELECT id FROM "Ingredient" WHERE name IN ('Wieprzowina (schab)', 'Jaja', 'Bułka')
);
```

### Phase 3: Normalize Steps (Optional, long-term)
```sql
-- Create RecipeStep table
CREATE TABLE "RecipeStep" (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  "recipeId" UUID NOT NULL REFERENCES "Recipe"(id) ON DELETE CASCADE,
  "stepNumber" INT NOT NULL,
  content TEXT NOT NULL,
  locale VARCHAR(5) DEFAULT 'pl',
  "createdAt" TIMESTAMP DEFAULT now(),
  CONSTRAINT unique_recipe_step UNIQUE ("recipeId", "stepNumber", locale)
);

-- Migrate existing steps from JSONB to RecipeStep
-- (requires custom migration script)
```

---

## 7️⃣ Current Status: Production Ready ✅

### What Works Now
- ✅ Recipe matching by ingredientId (UUID)
- ✅ Optional ingredients logic
- ✅ Servings stored correctly (base for scaling)
- ✅ Steps display correctly
- ✅ Economy calculations accurate
- ✅ Exclude feature for sequential recommendations

### What's Missing (Non-blocking)
- ⚠️ Recipe descriptions (can use localName for now)
- ⚠️ Recipe images (can use placeholder)
- ⚠️ Ingredient grouping (can show flat list)

### Recommendation
**Ship MVP without these fields**, add in v2:
1. Test with real users
2. Gather feedback on what's actually needed
3. Add description/images based on priority

---

## 8️⃣ Incomplete Recipe Data (URGENT)

### Current Issue
12 recipes have only 1-2 ingredients due to migration 038 using wrong ingredient names:

| Recipe | Current | Should Be |
|--------|---------|-----------|
| Kotlet schabowy | 4 | 4 ✅ (fixed) |
| Kopytka | 1 | 3-4 |
| Naleśniki | 1 | 3-4 |
| Pancakes | 1 | 3-4 |
| Gołąbki | 2 | 5-6 |
| Shakshuka | 2 | 5-6 |
| Others | 2 | 4-6 |

### Root Cause
Migration 038 uses incorrect names:
- `'Wieprzowina'` → should be `'Wieprzowina (schab)'`
- `'Jajko'` → should be `'Jaja'`
- `'Bułka tarta'` → should be `'Bułka'`
- Many more mismatches

### Solution
Create migration 042 to fix all 12 recipes:
1. Find correct ingredient names from Ingredient table
2. Add missing RecipeIngredient rows
3. Verify each recipe has realistic ingredient count (4-8)

**Priority:** 🔴 **HIGH** - Directly impacts recommendation quality

---

## 9️⃣ Summary

### Database Structure: ✅ **85% Canonical**
- Core tables aligned
- Missing nice-to-have fields (description, imageUrl, groupName)
- RecipeStep not normalized (acceptable for MVP)

### Data Quality: ⚠️ **60% Complete**
- 19/31 recipes complete (✅ Kotlet fixed)
- 12/31 recipes incomplete (❌ need fixing)
- All existing data follows correct structure

### API Format: ✅ **90% Canonical**
- Response structure matches target
- Minor naming differences (servings vs servingsDefault)
- Missing fields can default to null/empty

### Action Items
1. 🔴 **URGENT:** Fix 12 incomplete recipes (migration 042)
2. 🟡 **Important:** Add description, imageUrl, groupName columns
3. 🟢 **Nice:** Normalize RecipeStep table (v2 feature)

---

## 🎯 Conclusion

**Current structure is production-ready for MVP.**

The schema is well-designed and 85% aligned with canonical format. Missing fields are nice-to-have, not blockers. The real priority is **fixing incomplete recipe data** (12 recipes with 1-2 ingredients).

Once recipe data is complete, the system will provide excellent recommendations with accurate ingredient matching, proper optional handling, and realistic economy calculations.
