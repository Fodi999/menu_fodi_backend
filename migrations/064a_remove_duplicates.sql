-- Remove duplicate ingredients, keep oldest entry for each normalized_value
-- Purpose: Prepare database for UNIQUE constraint on normalized_value

-- Create temp table with ingredients to keep (oldest createdAt for each normalized_value)
CREATE TEMP TABLE ingredients_to_keep AS
SELECT DISTINCT ON (normalized_value) id
FROM "Ingredient"
WHERE normalized_value IS NOT NULL
ORDER BY normalized_value, "createdAt" ASC;

-- Delete duplicates (all except the ones to keep)
DELETE FROM "Ingredient"
WHERE normalized_value IS NOT NULL
  AND id NOT IN (SELECT id FROM ingredients_to_keep);

-- Show what was kept
SELECT normalized_value, id, name_en, "createdAt"
FROM "Ingredient"
WHERE normalized_value IN (
  SELECT normalized_value 
  FROM "Ingredient" 
  GROUP BY normalized_value 
  HAVING COUNT(*) >= 1
)
ORDER BY normalized_value, "createdAt";
