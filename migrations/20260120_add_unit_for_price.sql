-- Add unit_for_price column to user_fridge_price_history table
-- Migration: 2026-01-20 Add unit_for_price to price history

ALTER TABLE user_fridge_price_history
ADD COLUMN IF NOT EXISTS unit_for_price TEXT;

-- Add comment
COMMENT ON COLUMN user_fridge_price_history.unit_for_price IS 'Unit of measurement for the price (kg, l, pcs, g, ml)';

-- Update existing records to have a default unit (based on typical usage)
-- For now, set NULL for existing records - they will be updated on next price update
UPDATE user_fridge_price_history
SET unit_for_price = 'kg'
WHERE unit_for_price IS NULL
  AND price_per_unit < 1; -- Likely normalized to g

UPDATE user_fridge_price_history  
SET unit_for_price = 'pcs'
WHERE unit_for_price IS NULL
  AND price_per_unit >= 1; -- Likely per piece

