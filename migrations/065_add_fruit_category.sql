-- Migration: Add 'fruit' category to ingredient check constraint
-- Purpose: Allow AI to classify fruits separately from vegetables
-- Date: 2026-01-07

-- Drop old constraint
ALTER TABLE "Ingredient"
DROP CONSTRAINT IF EXISTS chk_ingredient_category;

-- Add new constraint with 'fruit' category
ALTER TABLE "Ingredient"
ADD CONSTRAINT chk_ingredient_category
CHECK (category IN ('protein', 'vegetable', 'fruit', 'dairy', 'grain', 'condiment', 'other'));

-- Update comment
COMMENT ON COLUMN "Ingredient".category IS 'Product category: protein, vegetable, fruit, dairy, grain, condiment, other';
