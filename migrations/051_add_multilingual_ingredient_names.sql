-- +goose Up
-- Migration: Add multilingual support for ingredients (RU/EN/PL)
-- Date: 2025-12-30
-- Description: Adds name_pl, name_en, name_ru, and normalized_value columns for search

-- Step 1: Add multilingual name columns
ALTER TABLE "Ingredient"
ADD COLUMN IF NOT EXISTS name_pl TEXT,
ADD COLUMN IF NOT EXISTS name_en TEXT,
ADD COLUMN IF NOT EXISTS name_ru TEXT;

-- Step 2: Add normalized search column
ALTER TABLE "Ingredient"
ADD COLUMN IF NOT EXISTS normalized_value TEXT;

-- Step 3: Migrate existing data (name → name_pl)
UPDATE "Ingredient"
SET name_pl = name
WHERE name_pl IS NULL;

-- Step 4: Generate normalized values (lowercase, no diacritics for search)
UPDATE "Ingredient"
SET normalized_value = LOWER(
    TRANSLATE(
        name_pl,
        'ąćęłńóśźżĄĆĘŁŃÓŚŹŻ',
        'acelnoszz ACELNOSZZ'
    )
)
WHERE normalized_value IS NULL;

-- Step 5: Create search indexes
-- Index for PL names (case-insensitive)
CREATE INDEX IF NOT EXISTS idx_ingredient_name_pl_lower
ON "Ingredient"(LOWER(name_pl));

-- Index for EN names (case-insensitive)
CREATE INDEX IF NOT EXISTS idx_ingredient_name_en_lower
ON "Ingredient"(LOWER(name_en));

-- Index for RU names (case-insensitive)
CREATE INDEX IF NOT EXISTS idx_ingredient_name_ru_lower
ON "Ingredient"(LOWER(name_ru));

-- Index for normalized search (fast autocomplete without language filter)
CREATE INDEX IF NOT EXISTS idx_ingredient_normalized_value
ON "Ingredient"(normalized_value);

-- Step 6: Add comments
COMMENT ON COLUMN "Ingredient".name_pl IS 'Polish name (primary language)';
COMMENT ON COLUMN "Ingredient".name_en IS 'English name (optional)';
COMMENT ON COLUMN "Ingredient".name_ru IS 'Russian name (optional)';
COMMENT ON COLUMN "Ingredient".normalized_value IS 'Normalized search value (lowercase, no diacritics) for fast multilingual search';

-- +goose Down
DROP INDEX IF EXISTS idx_ingredient_normalized_value;
DROP INDEX IF EXISTS idx_ingredient_name_ru_lower;
DROP INDEX IF EXISTS idx_ingredient_name_en_lower;
DROP INDEX IF EXISTS idx_ingredient_name_pl_lower;

ALTER TABLE "Ingredient"
DROP COLUMN IF EXISTS normalized_value,
DROP COLUMN IF EXISTS name_ru,
DROP COLUMN IF EXISTS name_en,
DROP COLUMN IF EXISTS name_pl;
