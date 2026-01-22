-- ============================================================================
-- Migration: Add canonical_id to Ingredient table
-- Purpose: Group similar ingredients (e.g., different vegetable oils)
-- Date: 2026-01-22
-- ============================================================================

-- Step 1: Add canonical_id column
ALTER TABLE "Ingredient" 
ADD COLUMN IF NOT EXISTS canonical_id VARCHAR(255);

-- Step 2: Create index for faster lookups
CREATE INDEX IF NOT EXISTS idx_ingredient_canonical 
ON "Ingredient"(canonical_id);

-- Step 3: Populate canonical_id for common duplicates

-- Vegetable oils (all types)
UPDATE "Ingredient" 
SET canonical_id = 'vegetable_oil'
WHERE id IN (
    '1b7cea8e-b026-4329-9d2e-c94952e3fa6c',  -- Olej roślinny
    '9ff773d2-a3ee-4f4b-bc45-4cfe0d7f680b'   -- Olej rzepakowy
)
OR name_en ILIKE '%vegetable oil%'
OR name_en ILIKE '%rapeseed oil%'
OR name_en ILIKE '%canola oil%'
OR name_en ILIKE '%sunflower oil%';

-- Salt (all types)
UPDATE "Ingredient"
SET canonical_id = 'salt'
WHERE name_en ILIKE '%salt%'
AND name_en NOT ILIKE '%salted%';  -- Exclude "salted butter"

-- Eggs (all types)
UPDATE "Ingredient"
SET canonical_id = 'eggs'
WHERE name_en ILIKE '%egg%'
AND name_en NOT ILIKE '%eggplant%';  -- Exclude eggplant

-- Milk (all types)
UPDATE "Ingredient"
SET canonical_id = 'milk'
WHERE name_en ILIKE '%milk%'
AND name_en NOT ILIKE '%coconut milk%'  -- Coconut milk is different
AND name_en NOT ILIKE '%almond milk%';  -- Almond milk is different

-- Butter (all types)
UPDATE "Ingredient"
SET canonical_id = 'butter'
WHERE name_en ILIKE '%butter%'
AND name_en NOT ILIKE '%peanut butter%'
AND name_en NOT ILIKE '%almond butter%';

-- Sugar (all types)  
UPDATE "Ingredient"
SET canonical_id = 'sugar'
WHERE name_en ILIKE '%sugar%'
AND name_en NOT ILIKE '%sugar snap%';  -- Sugar snap peas

-- Flour (all types)
UPDATE "Ingredient"
SET canonical_id = 'flour'
WHERE name_en ILIKE '%flour%';

-- Step 4: Verify migration
DO $$
DECLARE
    canonical_count INTEGER;
    oil_count INTEGER;
BEGIN
    SELECT COUNT(DISTINCT canonical_id) INTO canonical_count
    FROM "Ingredient"
    WHERE canonical_id IS NOT NULL;
    
    SELECT COUNT(*) INTO oil_count
    FROM "Ingredient"
    WHERE canonical_id = 'vegetable_oil';
    
    RAISE NOTICE '============================================';
    RAISE NOTICE 'Canonical groups created: %', canonical_count;
    RAISE NOTICE 'Vegetable oil variants: %', oil_count;
    RAISE NOTICE '============================================';
END $$;

-- Step 5: Show sample results
SELECT 
    canonical_id,
    COUNT(*) as variants,
    STRING_AGG(DISTINCT name_en, ', ') as examples
FROM "Ingredient"
WHERE canonical_id IS NOT NULL
GROUP BY canonical_id
ORDER BY variants DESC
LIMIT 10;
