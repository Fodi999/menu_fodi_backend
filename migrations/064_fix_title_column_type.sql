-- Migration 064: Fix title column type to match Recipe model
-- Changes title from TEXT to VARCHAR(255) to match user-generated Recipe model

-- Step 1: Change title column type to VARCHAR(255)
ALTER TABLE "Recipe" 
ALTER COLUMN title TYPE VARCHAR(255);

-- Step 2: Verify the change
SELECT 
    column_name,
    data_type,
    character_maximum_length,
    is_nullable
FROM information_schema.columns
WHERE table_name = 'Recipe' AND column_name = 'title';

-- Step 3: Test that existing data is preserved
SELECT 
    "canonicalName",
    title,
    LENGTH(title) as title_length,
    CASE 
        WHEN LENGTH(title) > 255 THEN '⚠️ Too long!'
        ELSE '✅ OK'
    END as status
FROM "Recipe"
ORDER BY "canonicalName"
LIMIT 10;
