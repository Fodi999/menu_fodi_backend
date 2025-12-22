-- Migration: Normalize all recipes to servings = 1 (base portion)
-- All cooking logic uses servingsMultiplier, so base recipes should be normalized to 1 serving

-- Цель: Привести все recipes.servings к единому стандарту = 1
-- После этого все расчёты делаются только через servingsMultiplier при cook

BEGIN;

-- 1. Обновляем все существующие рецепты, где servings != 1
UPDATE recipes
SET servings = 1
WHERE servings != 1;

-- 2. Меняем default для будущих рецептов
ALTER TABLE recipes 
ALTER COLUMN servings SET DEFAULT 1;

-- 3. Проверяем результат (для логов)
DO $$
DECLARE
    non_standard_count INT;
BEGIN
    SELECT COUNT(*) INTO non_standard_count 
    FROM recipes 
    WHERE servings != 1;
    
    RAISE NOTICE 'Recipes with servings != 1 after migration: %', non_standard_count;
END $$;

COMMIT;

-- Комментарий: 
-- После этой миграции все рецепты имеют servings = 1
-- При готовке используем servingsMultiplier (например, 2.0 = удвоить количество)
-- Формула: actualQuantity = baseQuantity * servingsMultiplier
