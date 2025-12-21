-- ======================================================
-- RECIPE CATALOG VERIFICATION SCRIPT
-- ======================================================
-- Purpose: Verify migrations 035/036 applied correctly
-- Run this in Neon SQL Editor AFTER applying migrations
-- ======================================================

-- 1️⃣ CHECK: Tables exist?
-- Expected: 4 tables (Recipe, RecipeIngredient, Allergen, DietTag)
SELECT table_name
FROM information_schema.tables
WHERE table_schema = 'public'
  AND table_name IN ('Recipe', 'RecipeIngredient', 'Allergen', 'DietTag', 'RecipeAllergen', 'RecipeDietTag')
ORDER BY table_name;

-- 2️⃣ CHECK: How many recipes in catalog?
-- Expected: 6 recipes (Pierogi, Bigos, Carbonara, Pizza, Eggs, Salad)
SELECT COUNT(*) AS recipes_count FROM "Recipe";

-- 3️⃣ CHECK: Recipe overview
-- Expected: Poland (PL), Italy (IT), Greece (GR) countries
SELECT 
  id,
  "canonicalName",
  "localName",
  country,
  category,
  difficulty,
  "timeMinutes",
  servings
FROM "Recipe"
ORDER BY country, "canonicalName"
LIMIT 50;

-- 4️⃣ CHECK: Recipe-Ingredient links exist?
-- Expected: >0 links (each recipe has multiple ingredients)
SELECT COUNT(*) AS recipe_ingredients_count FROM "RecipeIngredient";

-- 5️⃣ CHECK: Allergens seeded?
-- Expected: 14 allergens (gluten, lactose, eggs, fish, shellfish, nuts, etc.)
SELECT 
  key,
  "displayName",
  "iconEmoji"
FROM "Allergen"
ORDER BY "displayName";

-- 6️⃣ CHECK: Diet tags seeded?
-- Expected: 11 diet tags (vegetarian, vegan, keto, paleo, gluten-free, etc.)
SELECT 
  key,
  "displayName",
  description
FROM "DietTag"
ORDER BY "displayName";

-- 7️⃣ EXAMPLE: One recipe with ingredients
-- Expected: Recipe name + list of ingredients with quantities
SELECT
  r."localName" AS recipe,
  i.name AS ingredient,
  ri.quantity,
  ri.unit,
  ri.optional
FROM "Recipe" r
JOIN "RecipeIngredient" ri ON ri."recipeId" = r.id
JOIN "Ingredient" i ON i.id = ri."ingredientId"
WHERE r."canonicalName" = 'Spaghetti Carbonara'
ORDER BY ri."sortOrder";

-- 8️⃣ EXAMPLE: All recipes with ingredient counts
-- Expected: Each recipe should have multiple ingredients
SELECT
  r."localName" AS recipe,
  r.country,
  r."timeMinutes" AS time_min,
  r.difficulty,
  COUNT(ri.id) AS ingredients_count
FROM "Recipe" r
LEFT JOIN "RecipeIngredient" ri ON ri."recipeId" = r.id
GROUP BY r.id, r."localName", r.country, r."timeMinutes", r.difficulty
ORDER BY r.country, r."localName";

-- 9️⃣ EXAMPLE: Recipe allergens
-- Expected: Each recipe should have allergen tags
SELECT
  r."localName" AS recipe,
  a."displayName" AS allergen,
  a."iconEmoji"
FROM "Recipe" r
JOIN "RecipeAllergen" ra ON ra."recipeId" = r.id
JOIN "Allergen" a ON a.id = ra."allergenId"
ORDER BY r."localName", a."displayName";

-- 🔟 EXAMPLE: Recipe diet tags
-- Expected: Recipes tagged with diets (vegetarian, vegan, etc.)
SELECT
  r."localName" AS recipe,
  dt."displayName" AS diet_tag
FROM "Recipe" r
JOIN "RecipeDietTag" rdt ON rdt."recipeId" = r.id
JOIN "DietTag" dt ON dt.id = rdt."dietTagId"
ORDER BY r."localName", dt."displayName";

-- ======================================================
-- ✅ SUCCESS CRITERIA:
-- ======================================================
-- ✅ 6 tables exist (Recipe, RecipeIngredient, Allergen, DietTag, + junctions)
-- ✅ recipes_count = 6
-- ✅ recipe_ingredients_count > 30 (6 recipes × ~5-10 ingredients each)
-- ✅ 14 allergens
-- ✅ 11 diet tags
-- ✅ Recipes have countries: PL, IT, GR
-- ✅ Carbonara shows: Makaron, Boczek, Jajko, Parmezan, Czosnek
-- ✅ Each recipe has multiple allergens/diets
-- ======================================================
