-- Migration 061: Normalize all recipe servings to 1 (base portion)
-- All recipes should have servings = 1 as a base unit
-- Frontend will use servingsMultiplier for scaling

-- Update all recipes to have servings = 1
UPDATE "Recipe" SET servings = 1 WHERE servings != 1;

-- Verify the change
SELECT 
    COUNT(*) as total_recipes,
    COUNT(CASE WHEN servings = 1 THEN 1 END) as normalized_recipes,
    COUNT(CASE WHEN servings != 1 THEN 1 END) as remaining_non_normalized
FROM "Recipe";
