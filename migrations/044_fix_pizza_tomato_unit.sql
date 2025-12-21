-- Migration 044: Fix Pizza Margherita - Pomidor unit (ml → g)
-- Issue: Tomatoes are solid, should use 'g' not 'ml'

BEGIN;

-- Fix Pizza Margherita: Pomidor should be 'g' not 'ml'
UPDATE "RecipeIngredient" ri
SET unit = 'g'
FROM "Recipe" r, "Ingredient" i
WHERE ri."recipeId" = r.id
  AND ri."ingredientId" = i.id
  AND r."localName" = 'Pizza Margherita'
  AND i.name = 'Pomidor'
  AND ri.unit = 'ml';

-- Verify fix
SELECT 
    r."localName" AS recipe,
    i.name AS ingredient,
    ri.quantity,
    ri.unit AS fixed_unit
FROM "RecipeIngredient" ri
JOIN "Recipe" r ON r.id = ri."recipeId"
JOIN "Ingredient" i ON i.id = ri."ingredientId"
WHERE r."localName" = 'Pizza Margherita'
ORDER BY ri.quantity DESC;

COMMIT;

-- Expected output:
--       recipe      |   ingredient   | quantity | fixed_unit
-- ------------------+----------------+----------+------------
--  Pizza Margherita | Mąka pszenna   |   300.00 | g
--  Pizza Margherita | Mozzarella     |   200.00 | g
--  Pizza Margherita | Pomidor        |   150.00 | g         <-- FIXED
--  Pizza Margherita | Oliwa z oliwek |    20.00 | ml
