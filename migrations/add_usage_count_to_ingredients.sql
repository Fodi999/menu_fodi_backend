-- Add usage_count column to Ingredient table for tracking recipe usage
-- This enables efficient sorting by popularity without counting joins

ALTER TABLE "Ingredient" 
ADD COLUMN IF NOT EXISTS usage_count INTEGER DEFAULT 0;

-- Create index for fast sorting by usage
CREATE INDEX IF NOT EXISTS idx_ingredients_usage_count ON "Ingredient"(usage_count DESC);

-- Create indexes for other sort modes (if not exist)
CREATE INDEX IF NOT EXISTS idx_ingredients_created_at ON "Ingredient"("createdAt" DESC);
CREATE INDEX IF NOT EXISTS idx_ingredients_name ON "Ingredient"(name);
CREATE INDEX IF NOT EXISTS idx_ingredients_category ON "Ingredient"(category);

-- Initialize usage_count from existing recipe_ingredients relationships
-- Count how many times each ingredient is used in RecipeIngredient
UPDATE "Ingredient" i
SET usage_count = COALESCE((
    SELECT COUNT(DISTINCT "recipeId")
    FROM "RecipeIngredient" ri
    WHERE ri."ingredientId" = i.id
), 0)
WHERE usage_count = 0;

COMMENT ON COLUMN "Ingredient".usage_count IS 'Number of recipes using this ingredient. Updated on recipe create/update/delete.';
