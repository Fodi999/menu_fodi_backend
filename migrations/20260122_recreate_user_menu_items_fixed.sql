-- ============================================================================
-- Migration: Recreate user_menu_items table with correct constraint
-- Purpose: Fix UNIQUE constraint - only on (user_id, recipe_id, planned_for)
-- Date: 2026-01-22
-- 
-- ISSUE: Current constraint has status which prevents status transitions
-- ============================================================================

-- Step 1: Drop existing table
DROP TABLE IF EXISTS user_menu_items CASCADE;

-- Step 2: Recreate table with CORRECT constraint
CREATE TABLE user_menu_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    recipe_id UUID NOT NULL REFERENCES "Recipe"(id) ON DELETE CASCADE,
    
    -- Cooking parameters
    servings INTEGER NOT NULL DEFAULT 1 CHECK (servings > 0),
    planned_for DATE NOT NULL DEFAULT CURRENT_DATE,
    
    -- Status workflow: planned → cooking → completed
    status TEXT NOT NULL DEFAULT 'planned' 
        CHECK (status IN ('planned', 'cooking', 'completed', 'cancelled')),
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    started_cooking_at TIMESTAMP,
    completed_at TIMESTAMP,
    
    -- Metadata
    notes TEXT,
    
    -- Constraints - CORRECT: Only user_id, recipe_id, planned_for
    CONSTRAINT unique_user_recipe_today 
        UNIQUE (user_id, recipe_id, planned_for),
    
    CONSTRAINT check_dates CHECK (
        completed_at IS NULL OR (started_cooking_at IS NOT NULL AND completed_at >= started_cooking_at)
    ),
    
    CONSTRAINT check_status_timestamps CHECK (
        (status = 'cooking' AND started_cooking_at IS NOT NULL) OR
        (status != 'cooking' AND (started_cooking_at IS NULL OR started_cooking_at IS NOT NULL)) OR
        (status = 'planned' AND started_cooking_at IS NULL AND completed_at IS NULL)
    )
);

-- Step 3: Create indexes for fast queries
CREATE INDEX idx_user_menu_user_date ON user_menu_items(user_id, planned_for);
CREATE INDEX idx_user_menu_status ON user_menu_items(status);
CREATE INDEX idx_user_menu_recipe ON user_menu_items(recipe_id);
CREATE INDEX idx_user_menu_user_status ON user_menu_items(user_id, status);

-- Step 4: Add helpful comments
COMMENT ON TABLE user_menu_items IS 'Kitchen pipeline: recipes user wants to cook today';
COMMENT ON COLUMN user_menu_items.status IS 'planned=added to menu, cooking=in progress, completed=done';
COMMENT ON COLUMN user_menu_items.planned_for IS 'Date user plans to cook (default: today)';
COMMENT ON CONSTRAINT unique_user_recipe_today ON user_menu_items IS 'User can only add same recipe once per day';

-- Step 5: Verify table created successfully
SELECT 
    table_name, 
    column_name, 
    data_type,
    is_nullable
FROM information_schema.columns
WHERE table_name = 'user_menu_items'
ORDER BY ordinal_position;

-- Step 6: List all constraints
SELECT
    constraint_name,
    constraint_type,
    table_name
FROM information_schema.table_constraints
WHERE table_name = 'user_menu_items'
ORDER BY constraint_name;

-- Step 7: Show indexes
SELECT
    schemaname,
    tablename,
    indexname,
    indexdef
FROM pg_indexes
WHERE tablename = 'user_menu_items'
ORDER BY indexname;
