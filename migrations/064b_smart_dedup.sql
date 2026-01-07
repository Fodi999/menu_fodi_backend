-- Smart deduplication: update references, then remove duplicates
-- Keep oldest ingredient for each normalized_value

BEGIN;

-- Step 1: Create mapping table (duplicate_id -> keep_id)
CREATE TEMP TABLE ingredient_mapping AS
WITH ranked_ingredients AS (
  SELECT 
    id,
    normalized_value,
    ROW_NUMBER() OVER (PARTITION BY normalized_value ORDER BY "createdAt" ASC) as rn
  FROM "Ingredient"
  WHERE normalized_value IS NOT NULL
)
SELECT 
  dup.id as duplicate_id,
  keep.id as keep_id
FROM ranked_ingredients dup
JOIN ranked_ingredients keep 
  ON dup.normalized_value = keep.normalized_value 
  AND keep.rn = 1
WHERE dup.rn > 1;

-- Step 2: Update StockItem references
UPDATE "StockItem"
SET "ingredientId" = m.keep_id
FROM ingredient_mapping m
WHERE "StockItem"."ingredientId" = m.duplicate_id;

-- Step 3: Update RecipeIngredient references
UPDATE "RecipeIngredient"
SET "ingredientId" = m.keep_id
FROM ingredient_mapping m
WHERE "RecipeIngredient"."ingredientId" = m.duplicate_id;

-- Step 4: Delete duplicates
DELETE FROM "Ingredient"
WHERE id IN (SELECT duplicate_id FROM ingredient_mapping);

-- Step 5: Show results
SELECT 
  COUNT(*) as duplicates_removed
FROM ingredient_mapping;

COMMIT;
