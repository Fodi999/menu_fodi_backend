-- ============================================================================
-- Migration: Create user_menu_items table (Kitchen Pipeline)
-- Purpose: Track recipes user wants to cook TODAY
-- Date: 2026-01-22
-- 
-- Key Principle: Backend is the single source of truth
-- ============================================================================

-- Step 1: Create user_menu_items table
CREATE TABLE IF NOT EXISTS user_menu_items (
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
    notes TEXT,  -- User notes (e.g., "Extra spicy")
    
    -- Constraints
    CONSTRAINT unique_user_recipe_today 
        UNIQUE (user_id, recipe_id, planned_for, status)
);

-- Step 2: Indexes for fast queries
CREATE INDEX idx_user_menu_user_date ON user_menu_items(user_id, planned_for);
CREATE INDEX idx_user_menu_status ON user_menu_items(status);
CREATE INDEX idx_user_menu_recipe ON user_menu_items(recipe_id);

-- Step 3: Create ENUM type for status (optional, better type safety)
DO $$ BEGIN
    CREATE TYPE menu_item_status AS ENUM ('planned', 'cooking', 'completed', 'cancelled');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Step 4: Add helpful comment
COMMENT ON TABLE user_menu_items IS 'Kitchen pipeline: recipes user wants to cook today';
COMMENT ON COLUMN user_menu_items.status IS 'planned=added to menu, cooking=in progress, completed=done';
COMMENT ON COLUMN user_menu_items.planned_for IS 'Date user plans to cook (default: today)';

-- Step 5: Sample data for testing
DO $$
DECLARE
    test_user_id UUID;
    test_recipe_id UUID;
BEGIN
    -- Get test user (fodi85@gmail.ru)
    SELECT id INTO test_user_id FROM users WHERE email = 'fodi85@gmail.ru' LIMIT 1;
    
    -- Get test recipe (zharenye_yaytsa)
    SELECT id INTO test_recipe_id FROM "Recipe" WHERE "canonicalName" = 'zharenye_yaytsa' LIMIT 1;
    
    IF test_user_id IS NOT NULL AND test_recipe_id IS NOT NULL THEN
        -- Add sample menu item
        INSERT INTO user_menu_items (user_id, recipe_id, servings, status, planned_for)
        VALUES (test_user_id, test_recipe_id, 1, 'planned', CURRENT_DATE)
        ON CONFLICT (user_id, recipe_id, planned_for, status) DO NOTHING;
        
        RAISE NOTICE '✅ Sample menu item created for testing';
    END IF;
END $$;

-- Step 6: Verify migration
SELECT 
    COUNT(*) as total_menu_items,
    COUNT(*) FILTER (WHERE status = 'planned') as planned,
    COUNT(*) FILTER (WHERE status = 'cooking') as cooking,
    COUNT(*) FILTER (WHERE status = 'completed') as completed
FROM user_menu_items;

-- Step 7: Show sample data
SELECT 
    umi.id,
    u.email,
    r."canonicalName" as recipe,
    umi.servings,
    umi.status,
    umi.planned_for,
    umi.created_at
FROM user_menu_items umi
JOIN users u ON u.id = umi.user_id
JOIN "Recipe" r ON r.id = umi.recipe_id
LIMIT 5;
