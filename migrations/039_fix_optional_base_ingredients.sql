-- Migration: Mark base ingredients as optional
-- Purpose: Salt, pepper, oil, butter should not block canCookNow
-- Date: 2025-12-21
-- 
-- Problem: Basic ingredients like salt, oil, butter were marked as required,
-- causing recipes to appear "uncookable" when only these items were missing.
-- These ingredients don't define a dish and are almost always available.
--
-- Solution: Mark them as optional so they don't affect canCookNow and scoring.

UPDATE "RecipeIngredient" ri
SET optional = true
FROM "Ingredient" i
WHERE ri."ingredientId" = i.id
AND i.name IN (
  'Sól',
  'Pieprz cayenne',
  'Pieprz czarny',
  'Olej roślinny',
  'Oliwa z oliwek',
  'Masło'
);

-- Verification query (optional, for manual check):
-- SELECT 
--   i.name,
--   COUNT(ri.id) as usage_count,
--   SUM(CASE WHEN ri.optional = true THEN 1 ELSE 0 END) as optional_count
-- FROM "Ingredient" i
-- JOIN "RecipeIngredient" ri ON ri."ingredientId" = i.id
-- WHERE i.name IN ('Sól', 'Pieprz cayenne', 'Pieprz czarny', 'Olej roślinny', 'Oliwa z oliwek', 'Masło')
-- GROUP BY i.name;
