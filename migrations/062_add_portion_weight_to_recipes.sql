-- Migration 062: Add portion weight field to Recipe table
-- Adds portionWeightGrams - total weight of one serving in grams

-- Add the new column
ALTER TABLE "Recipe" 
ADD COLUMN "portionWeightGrams" INTEGER;

-- Add comment to explain the field
COMMENT ON COLUMN "Recipe"."portionWeightGrams" IS 'Total weight of one serving in grams (calculated from ingredients)';

-- Create index for potential filtering/sorting
CREATE INDEX idx_recipe_portion_weight ON "Recipe" ("portionWeightGrams");

-- Update existing recipes with calculated weights based on ingredients
-- This will be populated via application logic or manual updates
-- For now, set default approximate values for existing recipes

UPDATE "Recipe" SET "portionWeightGrams" = 150 WHERE "canonicalName" = 'Scrambled Eggs';
UPDATE "Recipe" SET "portionWeightGrams" = 350 WHERE "canonicalName" = 'Pizza Margherita';
UPDATE "Recipe" SET "portionWeightGrams" = 250 WHERE "canonicalName" = 'Greek Salad';
UPDATE "Recipe" SET "portionWeightGrams" = 200 WHERE "canonicalName" = 'Pierogi Ruskie';
UPDATE "Recipe" SET "portionWeightGrams" = 400 WHERE "canonicalName" = 'Polish Chicken Soup';
UPDATE "Recipe" SET "portionWeightGrams" = 450 WHERE "canonicalName" = 'Polish Hunters Stew';
UPDATE "Recipe" SET "portionWeightGrams" = 180 WHERE "canonicalName" = 'Polish Meat Dumplings';
UPDATE "Recipe" SET "portionWeightGrams" = 220 WHERE "canonicalName" = 'Polish Breaded Pork Chop';
UPDATE "Recipe" SET "portionWeightGrams" = 200 WHERE "canonicalName" = 'Polish Potato Pancakes';
UPDATE "Recipe" SET "portionWeightGrams" = 300 WHERE "canonicalName" = 'Spaghetti Carbonara';

-- Verify the changes
SELECT 
    "canonicalName",
    servings,
    "portionWeightGrams",
    CASE 
        WHEN "portionWeightGrams" IS NULL THEN '❌ Missing'
        ELSE '✅ Set'
    END as status
FROM "Recipe"
ORDER BY "canonicalName";
