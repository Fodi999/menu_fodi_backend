-- Migration: Add index for case-insensitive ingredient search
-- Author: AI Assistant
-- Date: 2025-12-14
-- Description: Creates index on LOWER(name) for fast autocomplete searches

-- Create functional index for case-insensitive search
CREATE INDEX IF NOT EXISTS idx_ingredient_name_lower
ON "Ingredient"(LOWER(name));

-- Comment
COMMENT ON INDEX idx_ingredient_name_lower IS 'Case-insensitive search index for ingredient autocomplete - makes searches instant even with large datasets';
