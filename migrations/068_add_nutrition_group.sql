-- Migration 068: Add nutrition_group column to Ingredient table
-- Separates culinary categories (UI) from nutritional grouping (AI, analytics)

-- Add nutrition_group column with default value
ALTER TABLE "Ingredient" 
ADD COLUMN IF NOT EXISTS nutrition_group VARCHAR(50);

-- Add check constraint for valid nutrition groups
ALTER TABLE "Ingredient"
ADD CONSTRAINT check_nutrition_group 
CHECK (nutrition_group IN (
    'protein',      -- Белки: мясо, рыба, яйца, бобовые
    'carbohydrate', -- Углеводы: крупы, макароны, хлеб, картофель
    'fat',          -- Жиры: масла, орехи, семена
    'vegetable',    -- Овощи (некрахмалистые)
    'fruit',        -- Фрукты и ягоды
    'dairy',        -- Молочные продукты
    'condiment',    -- Специи, соусы, приправы
    'other'         -- Прочее
));

-- Migrate existing data: map current category to nutrition_group
-- This is a smart mapping based on typical nutritional profiles
UPDATE "Ingredient" SET nutrition_group = CASE
    -- Protein sources
    WHEN category IN ('protein', 'fish', 'meat', 'egg', 'legume') THEN 'protein'
    
    -- Carbohydrate sources
    WHEN category IN ('grain', 'cereal', 'pasta', 'bread', 'starch') THEN 'carbohydrate'
    
    -- Vegetables (keep as vegetable)
    WHEN category = 'vegetable' THEN 'vegetable'
    
    -- Fruits (keep as fruit)
    WHEN category = 'fruit' THEN 'fruit'
    
    -- Dairy (keep as dairy)
    WHEN category = 'dairy' THEN 'dairy'
    
    -- Condiments (keep as condiment)
    WHEN category = 'condiment' THEN 'condiment'
    
    -- Everything else
    ELSE 'other'
END
WHERE nutrition_group IS NULL;

-- Make nutrition_group NOT NULL after migration
ALTER TABLE "Ingredient" 
ALTER COLUMN nutrition_group SET NOT NULL;

-- Add index for nutrition_group queries
CREATE INDEX IF NOT EXISTS idx_ingredient_nutrition_group 
ON "Ingredient"(nutrition_group);

-- Add comment explaining the separation
COMMENT ON COLUMN "Ingredient".category IS 'Culinary category for UI display (fish, meat, vegetable, etc.)';
COMMENT ON COLUMN "Ingredient".nutrition_group IS 'Nutritional grouping for AI and analytics (protein, carbohydrate, fat, etc.)';
