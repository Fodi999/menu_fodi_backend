-- Migration 066: Make canonicalName nullable for user-generated recipes
-- canonicalName is only required for catalog recipes, not user-created ones

-- Step 1: Drop NOT NULL constraint on canonicalName
ALTER TABLE "Recipe" 
ALTER COLUMN "canonicalName" DROP NOT NULL;

-- Step 2: Keep unique constraint (catalog recipes must have unique canonicalName)
-- The existing unique index will remain

-- Step 3: Add comment
COMMENT ON COLUMN "Recipe"."canonicalName" IS 'Canonical English name (required for catalog recipes, NULL for user recipes)';

-- Verify the change
SELECT 
    column_name,
    data_type,
    is_nullable,
    column_default
FROM information_schema.columns
WHERE table_name = 'Recipe' AND column_name IN ('canonicalName', 'localName', 'author_id');
