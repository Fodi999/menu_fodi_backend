-- Migration: Add cooked_at timestamp to user_saved_recipes
-- Purpose: Track when a saved recipe was actually cooked
-- Author: System
-- Date: 2025-12-22

-- Add cooked_at column (nullable - not all saved recipes have been cooked yet)
ALTER TABLE user_saved_recipes
ADD COLUMN cooked_at TIMESTAMPTZ;

-- Create index for queries filtering cooked recipes
CREATE INDEX idx_user_saved_recipes_cooked_at ON user_saved_recipes(cooked_at) WHERE cooked_at IS NOT NULL;

-- Create composite index for user queries (e.g., "show me recipes I saved but haven't cooked")
CREATE INDEX idx_user_saved_recipes_user_cooked ON user_saved_recipes(user_id, cooked_at);

COMMENT ON COLUMN user_saved_recipes.cooked_at IS 'Timestamp when the user actually cooked this recipe (NULL if not cooked yet)';
