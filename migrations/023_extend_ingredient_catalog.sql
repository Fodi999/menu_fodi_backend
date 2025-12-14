-- Migration: Extend Ingredient table for professional catalog
-- Author: AI Assistant
-- Date: 2025-12-14
-- Description: Adds category, default shelf life, and default price fields to support catalog functionality

-- Add category column
ALTER TABLE "Ingredient"
ADD COLUMN IF NOT EXISTS category TEXT;

-- Add default shelf life (days)
ALTER TABLE "Ingredient"
ADD COLUMN IF NOT EXISTS "defaultShelfLifeDays" INT;

-- Add default price per unit (PLN)
ALTER TABLE "Ingredient"
ADD COLUMN IF NOT EXISTS "defaultPricePerUnit" DECIMAL(10,2);

-- Create index on category for filtering
CREATE INDEX IF NOT EXISTS idx_ingredient_category
ON "Ingredient"(category);

-- Create index on name for sorting
CREATE INDEX IF NOT EXISTS idx_ingredient_name
ON "Ingredient"(name);

-- Set default category for existing records (if any)
UPDATE "Ingredient"
SET category = 'other'
WHERE category IS NULL;

-- Make category NOT NULL after setting defaults
ALTER TABLE "Ingredient"
ALTER COLUMN category SET NOT NULL;

-- Add check constraint for valid categories
ALTER TABLE "Ingredient"
ADD CONSTRAINT chk_ingredient_category
CHECK (category IN ('protein', 'vegetable', 'dairy', 'grain', 'condiment', 'other'));

-- Comments
COMMENT ON COLUMN "Ingredient".category IS 'Product category: protein, vegetable, dairy, grain, condiment, other';
COMMENT ON COLUMN "Ingredient"."defaultShelfLifeDays" IS 'Default shelf life in days (optional)';
COMMENT ON COLUMN "Ingredient"."defaultPricePerUnit" IS 'Default price per unit in PLN (optional)';
