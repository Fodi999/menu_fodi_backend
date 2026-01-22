-- ============================================================================
-- HOTFIX: Drop and recreate user_menu_items with CORRECT constraint
-- THIS MUST BE RUN MANUALLY AGAINST PRODUCTION DATABASE
-- ============================================================================

-- Step 1: Drop the existing table (this will cascade delete foreign keys)
DROP TABLE IF EXISTS user_menu_items CASCADE;

-- Step 2: Create table with CORRECT structure
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
    
    -- ✅ CORRECT CONSTRAINT - only (user_id, recipe_id, planned_for)
    CONSTRAINT unique_user_recipe_today 
        UNIQUE (user_id, recipe_id, planned_for)
);

-- Step 3: Create indexes
CREATE INDEX idx_user_menu_user_date ON user_menu_items(user_id, planned_for);
CREATE INDEX idx_user_menu_status ON user_menu_items(status);
CREATE INDEX idx_user_menu_recipe ON user_menu_items(recipe_id);
CREATE INDEX idx_user_menu_user_status ON user_menu_items(user_id, status);

-- Step 4: Verify
SELECT 'Table recreated successfully!' as status;
SELECT constraint_name, constraint_type
FROM information_schema.table_constraints
WHERE table_name = 'user_menu_items'
ORDER BY constraint_name;
