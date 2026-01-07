-- Migration: Add NOT NULL constraint and UNIQUE INDEX to normalized_value
-- Purpose: Prevent duplicate ingredients across languages
-- Date: 2026-01-07

-- Step 1: Set NULL values to lowercase name_en (or name_pl as fallback)
UPDATE "Ingredient"
SET normalized_value = LOWER(COALESCE(name_en, name_pl, name))
WHERE normalized_value IS NULL;

-- Step 2: Add NOT NULL constraint
ALTER TABLE "Ingredient"
ALTER COLUMN normalized_value SET NOT NULL;

-- Step 3: Create UNIQUE index for duplicate prevention
CREATE UNIQUE INDEX IF NOT EXISTS uniq_ingredient_normalized
ON "Ingredient"(normalized_value);

-- Verification query (optional, comment out in production)
-- SELECT normalized_value, COUNT(*) 
-- FROM "Ingredient" 
-- GROUP BY normalized_value 
-- HAVING COUNT(*) > 1;
