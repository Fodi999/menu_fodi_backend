-- Migration 067: Add status field to Recipe table
-- Purpose: Add status (draft/published/archived) for recipe lifecycle management
-- Date: 2026-01-05

-- Add status column with default value 'draft'
ALTER TABLE "Recipe" 
ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'draft';

-- Add check constraint for valid status values
ALTER TABLE "Recipe" 
ADD CONSTRAINT check_recipe_status 
CHECK (status IN ('draft', 'published', 'archived'));

-- Create index for status (for filtering published recipes)
CREATE INDEX IF NOT EXISTS idx_recipe_status ON "Recipe"(status);

-- Update existing recipes to 'published' if they have author_id NULL (catalog recipes)
UPDATE "Recipe" 
SET status = 'published' 
WHERE author_id IS NULL;

-- Update existing user recipes to 'published' (assuming they were created before this feature)
UPDATE "Recipe" 
SET status = 'published' 
WHERE author_id IS NOT NULL;

COMMENT ON COLUMN "Recipe".status IS 'Recipe lifecycle status: draft (editing), published (live), archived (hidden)';

-- Verify changes
SELECT 
    status, 
    COUNT(*) as count,
    COUNT(*) FILTER (WHERE author_id IS NULL) as catalog_count,
    COUNT(*) FILTER (WHERE author_id IS NOT NULL) as user_count
FROM "Recipe"
GROUP BY status
ORDER BY status;
