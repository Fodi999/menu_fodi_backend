-- Migration: Create user_saved_recipes table
-- Description: Stores recipes saved by users for later reference
-- Date: 2025-12-21

CREATE TABLE IF NOT EXISTS user_saved_recipes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    user_id TEXT NOT NULL REFERENCES "User"(id) ON DELETE CASCADE,
    recipe_id UUID NOT NULL REFERENCES "Recipe"(id) ON DELETE CASCADE,
    
    servings INT NOT NULL DEFAULT 2 CHECK (servings > 0),
    source TEXT NOT NULL DEFAULT 'fridge',
    
    saved_at TIMESTAMP NOT NULL DEFAULT now(),
    
    UNIQUE (user_id, recipe_id)
);

-- Index for faster lookups by user
CREATE INDEX IF NOT EXISTS idx_user_saved_recipes_user_id ON user_saved_recipes(user_id);

-- Index for sorting by saved_at
CREATE INDEX IF NOT EXISTS idx_user_saved_recipes_saved_at ON user_saved_recipes(user_id, saved_at DESC);

COMMENT ON TABLE user_saved_recipes IS 'User saved recipes for later cooking';
COMMENT ON COLUMN user_saved_recipes.servings IS 'Number of servings user wants to cook';
COMMENT ON COLUMN user_saved_recipes.source IS 'Where recipe was found: fridge, search, etc';
