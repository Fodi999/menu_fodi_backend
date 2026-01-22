-- ============================================================================
-- Migration: Fix unique constraint in user_menu_items table
-- Purpose: Remove 'status' from unique constraint - status is transient!
-- Date: 2026-01-22
-- 
-- PROBLEM: UNIQUE (user_id, recipe_id, planned_for, status) prevents status transitions
--   - When user calls /start (planned → cooking), tries to INSERT with new status
--   - Then /complete (cooking → completed) tries to INSERT again  
--   - Both violate UNIQUE because only status differs!
--
-- SOLUTION: UNIQUE (user_id, recipe_id, planned_for) only
--   - Prevents SAME recipe SAME day (correct)
--   - Allows status transitions (correct)
-- ============================================================================

-- Step 1: Drop old constraint
ALTER TABLE user_menu_items
DROP CONSTRAINT IF EXISTS unique_user_recipe_today;

-- Step 2: Add new constraint (without status)
ALTER TABLE user_menu_items
ADD CONSTRAINT unique_user_recipe_today 
UNIQUE (user_id, recipe_id, planned_for);

-- Step 3: Verify it works
SELECT 
    constraint_name, 
    constraint_type
FROM information_schema.table_constraints
WHERE table_name = 'user_menu_items' AND constraint_name = 'unique_user_recipe_today';

-- Step 4: Show current state
SELECT 
    COUNT(*) as total_menu_items,
    COUNT(*) FILTER (WHERE status = 'planned') as planned,
    COUNT(*) FILTER (WHERE status = 'cooking') as cooking,
    COUNT(*) FILTER (WHERE status = 'completed') as completed,
    COUNT(*) FILTER (WHERE status = 'cancelled') as cancelled
FROM user_menu_items;
