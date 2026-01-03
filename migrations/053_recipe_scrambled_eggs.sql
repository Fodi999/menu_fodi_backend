-- ========================================
-- RECIPE: Scrambled Eggs (Яичница)
-- ========================================
-- Date: 2026-01-03
-- Purpose: Add simple baseline recipe for testing recipe matching system
-- Why this recipe: Minimal ingredients, easy difficulty, perfect for testing

-- ========================================
-- RECIPE DETAILS
-- ========================================
-- Canonical Name: Scrambled Eggs
-- Local Name: Яичница (Russian)
-- Category: main
-- Difficulty: easy
-- Time: 10 minutes
-- Servings: 1 ⚠️ IMPORTANT (base serving for individual cooking)
-- Country: International

-- ========================================
-- STEP 1: Update ingredient translations
-- ========================================

-- Egg (яйцо)
UPDATE "Ingredient" 
SET 
  name_ru = 'яйцо', 
  name_en = 'egg', 
  normalized_value = 'jaja'
WHERE name = 'Jaja';

-- Butter (масло)
UPDATE "Ingredient" 
SET 
  name_ru = 'масло', 
  name_en = 'butter', 
  normalized_value = 'maslo'
WHERE name = 'Masło';

-- Salt (соль)
UPDATE "Ingredient" 
SET 
  name_ru = 'соль', 
  name_en = 'salt', 
  normalized_value = 'sol'
WHERE name = 'Sól';

-- ========================================
-- STEP 2: Update recipe (already exists)
-- ========================================

UPDATE "Recipe"
SET 
  "localName" = 'Яичница',
  servings = 1,  -- ⚠️ Changed from 2 to 1 (baseline individual serving)
  description = 'Классическая яичница — простой и быстрый завтрак. Идеально для одного человека.',
  steps = '[
    {"step": 1, "instruction": "Разбить яйца в миску"},
    {"step": 2, "instruction": "Разогреть сковороду с маслом"},
    {"step": 3, "instruction": "Вылить яйца на сковороду"},
    {"step": 4, "instruction": "Жарить 3-5 минут, помешивая"},
    {"step": 5, "instruction": "Посолить по вкусу"}
  ]'::jsonb,
  "nutritionProfile" = '{"type": "balanced", "calories": 180, "protein": 12, "fat": 13, "carbs": 1}'::jsonb
WHERE "canonicalName" = 'Scrambled Eggs';

-- ========================================
-- STEP 3: Update ingredient quantities (for 1 serving)
-- ========================================

-- Egg: 2 pcs (was 4 for 2 servings)
UPDATE "RecipeIngredient"
SET quantity = 2.00
WHERE "recipeId" = (SELECT id FROM "Recipe" WHERE "canonicalName" = 'Scrambled Eggs')
  AND "ingredientId" = (SELECT id FROM "Ingredient" WHERE name = 'Jaja');

-- Butter: 10g (was 20g for 2 servings) - OPTIONAL
UPDATE "RecipeIngredient"
SET quantity = 10.00
WHERE "recipeId" = (SELECT id FROM "Recipe" WHERE "canonicalName" = 'Scrambled Eggs')
  AND "ingredientId" = (SELECT id FROM "Ingredient" WHERE name = 'Masło');

-- Remove milk (not needed for classic scrambled eggs)
DELETE FROM "RecipeIngredient"
WHERE "recipeId" = (SELECT id FROM "Recipe" WHERE "canonicalName" = 'Scrambled Eggs')
  AND "ingredientId" IN (SELECT id FROM "Ingredient" WHERE name LIKE 'Mleko%');

-- ========================================
-- STEP 4: Add salt (optional)
-- ========================================

INSERT INTO "RecipeIngredient" (
  "recipeId",
  "ingredientId",
  "ingredientKey",
  quantity,
  unit,
  optional,
  "sortOrder"
)
SELECT 
  r.id,
  i.id,
  'salt',
  2.00,
  'g',
  true,  -- OPTIONAL ingredient
  3
FROM "Recipe" r
CROSS JOIN "Ingredient" i
WHERE r."canonicalName" = 'Scrambled Eggs'
  AND i.name = 'Sól'
ON CONFLICT DO NOTHING;  -- Skip if already exists

-- ========================================
-- VERIFICATION QUERIES
-- ========================================

-- 1. Check recipe details
SELECT 
  "canonicalName",
  "localName",
  category,
  difficulty,
  "timeMinutes",
  servings,
  description
FROM "Recipe"
WHERE "canonicalName" = 'Scrambled Eggs';

-- Expected result:
-- canonicalName: Scrambled Eggs
-- localName: Яичница
-- category: main
-- difficulty: easy
-- timeMinutes: 10
-- servings: 1  ✅
-- description: Классическая яичница...

-- 2. Check ingredients
SELECT 
  i.name,
  i.name_ru,
  i.name_en,
  ri.quantity,
  ri.unit,
  ri.optional
FROM "RecipeIngredient" ri
JOIN "Ingredient" i ON ri."ingredientId" = i.id
JOIN "Recipe" r ON ri."recipeId" = r.id
WHERE r."canonicalName" = 'Scrambled Eggs'
ORDER BY ri."sortOrder";

-- Expected result:
-- Jaja  | яйцо  | egg    | 2.00  | pcs | false (REQUIRED)
-- Masło | масло | butter | 10.00 | g   | true  (OPTIONAL)
-- Sól   | соль  | salt   | 2.00  | g   | true  (OPTIONAL)

-- 3. Full recipe with ingredients (JSON)
SELECT 
  r."canonicalName",
  r."localName",
  r.servings,
  json_agg(
    json_build_object(
      'name_ru', i.name_ru,
      'quantity', ri.quantity,
      'unit', ri.unit,
      'optional', ri.optional
    ) ORDER BY ri."sortOrder"
  ) as ingredients
FROM "Recipe" r
JOIN "RecipeIngredient" ri ON r.id = ri."recipeId"
JOIN "Ingredient" i ON ri."ingredientId" = i.id
WHERE r."canonicalName" = 'Scrambled Eggs'
GROUP BY r.id;

-- ========================================
-- MATCHING LOGIC TEST CASES
-- ========================================

-- Test Case 1: User has eggs only
-- Expected: Recipe should be suggested (100% match on required ingredients)
-- Missing: масло (optional), соль (optional) → OK to cook!

-- Test Case 2: User has eggs + butter
-- Expected: Recipe should be suggested (100% match + 1 optional)
-- Missing: соль (optional) → OK to cook!

-- Test Case 3: User has all ingredients
-- Expected: Perfect match (100% + all optionals)

-- Test Case 4: User has no eggs
-- Expected: Recipe should NOT be suggested (0% match on required)

-- ========================================
-- SCORING FORMULA
-- ========================================
/*
score = (available_required / total_required) * 100

For Scrambled Eggs:
- Required ingredients: 1 (яйцо)
- Optional ingredients: 2 (масло, соль)

Scoring examples:
- Has яйцо: 1/1 = 100% ✅ CAN COOK
- No яйцо: 0/1 = 0% ❌ CANNOT COOK

Optional ingredients do NOT block cooking!
*/

-- ========================================
-- RECIPE MATCHING QUERY (Example)
-- ========================================
/*
SELECT 
  r.id,
  r."canonicalName",
  r."localName",
  COUNT(DISTINCT ri."ingredientId") as total_ingredients,
  COUNT(DISTINCT CASE WHEN NOT ri.optional THEN ri."ingredientId" END) as required_ingredients,
  COUNT(DISTINCT CASE 
    WHEN NOT ri.optional AND ufi.ingredient_id IS NOT NULL 
    THEN ri."ingredientId" 
  END) as available_required,
  ROUND(
    COUNT(DISTINCT CASE WHEN NOT ri.optional AND ufi.ingredient_id IS NOT NULL THEN ri."ingredientId" END)::numeric 
    / NULLIF(COUNT(DISTINCT CASE WHEN NOT ri.optional THEN ri."ingredientId" END), 0) * 100,
    0
  ) as match_score
FROM "Recipe" r
JOIN "RecipeIngredient" ri ON r.id = ri."recipeId"
LEFT JOIN user_fridge_items ufi ON ri."ingredientId" = ufi.ingredient_id 
  AND ufi.user_id = '<USER_ID>'
WHERE r."canonicalName" = 'Scrambled Eggs'
GROUP BY r.id
HAVING COUNT(DISTINCT CASE WHEN NOT ri.optional THEN ri."ingredientId" END) > 0;

-- If match_score = 100 → User CAN cook this recipe ✅
*/

-- ========================================
-- NOTES
-- ========================================
/*
WHY THIS RECIPE IS PERFECT FOR TESTING:
✅ Minimal ingredients (only 1 required)
✅ Easy difficulty (user-friendly)
✅ Fast cooking (10 minutes)
✅ servings = 1 (baseline for scaling with targetServings)
✅ Clear optional vs required distinction
✅ International recipe (works for any user)
✅ Common ingredients (likely in fridge)

NEXT STEPS:
1. Test recipe matching API: GET /api/recipes/match
2. Test cooking with targetServings: POST /api/recipes/{id}/cook {"targetServings": 2}
3. Verify ingredient deduction from fridge
4. Check RecipeCookLog entry with correct servingsMultiplier
5. Add more recipes with varying complexity
*/
