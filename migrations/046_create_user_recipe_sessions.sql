-- Migration: Create user_recipe_sessions table
-- Description: Tracks user's recipe browsing session to avoid showing duplicates
-- Date: 2025-12-21

CREATE TABLE IF NOT EXISTS user_recipe_sessions (
    user_id TEXT PRIMARY KEY REFERENCES "User"(id) ON DELETE CASCADE,
    
    last_recipe_id UUID REFERENCES "Recipe"(id) ON DELETE SET NULL,
    excluded_recipe_ids UUID[] NOT NULL DEFAULT '{}',
    
    updated_at TIMESTAMP NOT NULL DEFAULT now()
);

-- Index for efficient lookups
CREATE INDEX IF NOT EXISTS idx_user_recipe_sessions_updated_at ON user_recipe_sessions(updated_at);

COMMENT ON TABLE user_recipe_sessions IS 'Tracks recipe recommendation sessions to prevent duplicate suggestions';
COMMENT ON COLUMN user_recipe_sessions.last_recipe_id IS 'Last recipe shown to user';
COMMENT ON COLUMN user_recipe_sessions.excluded_recipe_ids IS 'Array of recipe IDs already shown in this session';
