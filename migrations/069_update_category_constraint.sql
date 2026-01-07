-- Migration 069: Update category constraint to support new culinary categories
-- Replaces old categories (protein, vegetable) with specific ones (fish, meat, egg, vegetable)

-- Step 1: Drop old constraint first
ALTER TABLE "Ingredient" 
DROP CONSTRAINT IF EXISTS chk_ingredient_category;

-- Step 2: Migrate existing data BEFORE adding new constraint
UPDATE "Ingredient" SET category = CASE
    -- If it's clearly fish (by name or normalized_value)
    WHEN LOWER(name) LIKE '%salmon%' OR LOWER(name) LIKE '%tuna%' OR LOWER(name) LIKE '%fish%' 
         OR LOWER(COALESCE(name_en, '')) LIKE '%fish%' OR LOWER(COALESCE(name_en, '')) LIKE '%salmon%'
         OR LOWER(COALESCE(name_ru, '')) LIKE '%лосось%' OR LOWER(COALESCE(name_ru, '')) LIKE '%рыб%'
         OR LOWER(COALESCE(name_pl, '')) LIKE '%łosoś%' OR LOWER(COALESCE(name_pl, '')) LIKE '%ryb%'
    THEN 'fish'
    
    -- If it's clearly egg
    WHEN LOWER(name) LIKE '%egg%' OR LOWER(COALESCE(name_ru, '')) LIKE '%яйц%' 
         OR LOWER(COALESCE(name_pl, '')) LIKE '%jaj%'
    THEN 'egg'
    
    -- Otherwise protein → meat (chicken, beef, pork)
    WHEN category = 'protein' THEN 'meat'
    
    -- Keep everything else as is
    ELSE category
END
WHERE category NOT IN ('fish', 'meat', 'egg', 'vegetable', 'fruit', 'dairy', 'grain', 'condiment', 'other');

-- Step 3: Set any remaining invalid categories to 'other'
UPDATE "Ingredient" SET category = 'other'
WHERE category NOT IN ('fish', 'meat', 'egg', 'vegetable', 'fruit', 'dairy', 'grain', 'condiment', 'other');

-- Step 4: NOW add new constraint (after data is clean)
ALTER TABLE "Ingredient"
ADD CONSTRAINT chk_ingredient_category 
CHECK (category IN (
    'fish',       -- Рыба и морепродукты
    'meat',       -- Мясо и птица
    'egg',        -- Яйца
    'vegetable',  -- Овощи
    'fruit',      -- Фрукты и ягоды
    'dairy',      -- Молочные продукты
    'grain',      -- Крупы, макароны, хлеб
    'condiment',  -- Специи, соусы, масла
    'other'       -- Прочее
));
