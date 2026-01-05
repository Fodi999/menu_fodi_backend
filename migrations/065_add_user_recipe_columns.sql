-- Migration 065: Add user-generated recipe columns to Recipe table
-- Adds columns needed for user-created recipes (POST /api/recipes)

-- Step 1: Add author_id column (for user-generated recipes)
ALTER TABLE "Recipe" 
ADD COLUMN author_id VARCHAR(255);

-- Step 2: Add nutrition metrics columns (all optional for catalog recipes)
ALTER TABLE "Recipe" 
ADD COLUMN gross_weight INTEGER,
ADD COLUMN net_weight INTEGER,
ADD COLUMN calories INTEGER,
ADD COLUMN protein DECIMAL(10,2),
ADD COLUMN fats DECIMAL(10,2),
ADD COLUMN carbs DECIMAL(10,2),
ADD COLUMN yield INTEGER,
ADD COLUMN cost DECIMAL(10,2);

-- Step 3: Add ChefTokens system columns
ALTER TABLE "Recipe" 
ADD COLUMN tokens_reward INTEGER DEFAULT 10,
ADD COLUMN views_count INTEGER DEFAULT 0,
ADD COLUMN tokens_earned INTEGER DEFAULT 0;

-- Step 4: Add foreign key constraint for author_id
ALTER TABLE "Recipe"
ADD CONSTRAINT fk_recipe_author
FOREIGN KEY (author_id) REFERENCES "User"(id) ON DELETE SET NULL;

-- Step 5: Add index on author_id
CREATE INDEX idx_recipe_author ON "Recipe"(author_id);

-- Step 6: Add comments
COMMENT ON COLUMN "Recipe".author_id IS 'User who created this recipe (NULL for catalog recipes)';
COMMENT ON COLUMN "Recipe".tokens_reward IS 'ChefTokens reward for creating recipe';
COMMENT ON COLUMN "Recipe".views_count IS 'Number of times recipe was viewed';
COMMENT ON COLUMN "Recipe".tokens_earned IS 'Total tokens earned from this recipe';

-- Verify the changes
SELECT 
    column_name,
    data_type,
    is_nullable,
    column_default
FROM information_schema.columns
WHERE table_name = 'Recipe' 
AND column_name IN ('author_id', 'gross_weight', 'tokens_reward', 'views_count')
ORDER BY column_name;
