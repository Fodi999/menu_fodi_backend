-- Migration 063: Add title column to Recipe table
-- Adds unified title field (can be populated from name_pl for Polish recipes)

-- Step 1: Add title column (nullable initially)
ALTER TABLE "Recipe" 
ADD COLUMN title TEXT;

-- Step 2: Populate title from name_pl (primary language)
-- If name_pl is null, fallback to canonicalName
UPDATE "Recipe" 
SET title = COALESCE(name_pl, "canonicalName");

-- Step 3: Make title NOT NULL after populating data
ALTER TABLE "Recipe" 
ALTER COLUMN title SET NOT NULL;

-- Step 4: Add comment to explain the field
COMMENT ON COLUMN "Recipe".title IS 'Primary title for the recipe (unified across languages, typically Polish)';

-- Verify the migration
SELECT 
    "canonicalName",
    title,
    name_pl,
    CASE 
        WHEN title IS NULL THEN '❌ Missing'
        WHEN title = name_pl THEN '✅ From name_pl'
        WHEN title = "canonicalName" THEN '⚠️ From canonicalName'
        ELSE '✅ Set'
    END as status
FROM "Recipe"
ORDER BY "canonicalName"
LIMIT 10;
