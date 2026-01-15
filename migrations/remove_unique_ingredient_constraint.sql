-- Migration: Remove UNIQUE constraint to allow multiple batches of same ingredient
-- Date: 2026-01-15
-- Reason: Allow users to have multiple entries of same ingredient with different:
--   - expiry dates (different batches)
--   - prices (bought at different times)
--   - arrival dates (restocking)

-- Remove the UNIQUE constraint
ALTER TABLE user_fridge_items
DROP CONSTRAINT IF EXISTS user_fridge_items_user_id_ingredient_id_key;

-- Add index for performance (replace removed unique index)
CREATE INDEX IF NOT EXISTS idx_user_fridge_items_user_ingredient 
ON user_fridge_items(user_id, ingredient_id);

-- Add comment to table
COMMENT ON TABLE user_fridge_items IS 
'Fridge items - allows multiple batches of same ingredient with different expiry dates and prices';
