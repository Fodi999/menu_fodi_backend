-- Migration 006: Add missing recipe nutrition metrics and tokens_earned
-- Adds: protein, fats, carbs (nutrition data) and tokens_earned (ChefTokens system)

-- Add missing nutrition columns
ALTER TABLE "Recipe" ADD COLUMN IF NOT EXISTS protein NUMERIC(10,2);
ALTER TABLE "Recipe" ADD COLUMN IF NOT EXISTS fats NUMERIC(10,2);
ALTER TABLE "Recipe" ADD COLUMN IF NOT EXISTS carbs NUMERIC(10,2);
ALTER TABLE "Recipe" ADD COLUMN IF NOT EXISTS tokens_earned INTEGER DEFAULT 0;

-- Add comments
COMMENT ON COLUMN "Recipe".protein IS 'Protein content in grams';
COMMENT ON COLUMN "Recipe".fats IS 'Fats content in grams';
COMMENT ON COLUMN "Recipe".carbs IS 'Carbohydrates content in grams';
COMMENT ON COLUMN "Recipe".tokens_earned IS 'Total ChefTokens earned from views';

-- Update existing recipes with default nutrition values
UPDATE "Recipe"
SET 
  protein = 0.0,
  fats = 0.0,
  carbs = 0.0,
  tokens_earned = 0
WHERE protein IS NULL OR fats IS NULL OR carbs IS NULL OR tokens_earned IS NULL;

-- Create index on calories for filtering recipes by nutrition
CREATE INDEX IF NOT EXISTS idx_recipe_calories ON "Recipe"(calories);

-- Create index on tokens_earned for leaderboard
CREATE INDEX IF NOT EXISTS idx_recipe_tokens_earned ON "Recipe"(tokens_earned DESC);
