-- =====================================================
-- ШАГ 2: УСТАНОВКА defaultShelfLifeDays и defaultPricePerUnit
-- =====================================================
-- Выполнить ПОСЛЕ шага 1 (удаления дубликатов)
-- =====================================================

-- Показываем текущее состояние
SELECT '=== BEFORE UPDATES ===' as status;
SELECT 
    "category",
    COUNT(*) as total,
    COUNT(CASE WHEN "defaultShelfLifeDays" IS NULL THEN 1 END) as missing_shelf_life,
    COUNT(CASE WHEN "defaultPricePerUnit" IS NULL THEN 1 END) as missing_price
FROM "Ingredient"
GROUP BY "category"
ORDER BY "category";

-- =====================================================
-- УСТАНОВКА defaultShelfLifeDays ПО КАТЕГОРИЯМ
-- =====================================================

-- Protein (мясо, рыба) - 3 дня
UPDATE "Ingredient" 
SET "defaultShelfLifeDays" = 3
WHERE "defaultShelfLifeDays" IS NULL 
  AND "category" = 'protein';

-- Vegetable (овощи, фрукты) - 7 дней
UPDATE "Ingredient" 
SET "defaultShelfLifeDays" = 7
WHERE "defaultShelfLifeDays" IS NULL 
  AND "category" = 'vegetable';

-- Dairy (молочные) - 14 дней
UPDATE "Ingredient" 
SET "defaultShelfLifeDays" = 14
WHERE "defaultShelfLifeDays" IS NULL 
  AND "category" = 'dairy';

-- Grain (крупы, макароны) - 365 дней (1 год)
UPDATE "Ingredient" 
SET "defaultShelfLifeDays" = 365
WHERE "defaultShelfLifeDays" IS NULL 
  AND "category" = 'grain';

-- Condiment (специи, соусы) - 365 дней
UPDATE "Ingredient" 
SET "defaultShelfLifeDays" = 365
WHERE "defaultShelfLifeDays" IS NULL 
  AND "category" = 'condiment';

-- Other (прочее) - 180 дней (полгода)
UPDATE "Ingredient" 
SET "defaultShelfLifeDays" = 180
WHERE "defaultShelfLifeDays" IS NULL 
  AND "category" = 'other';

-- =====================================================
-- УСТАНОВКА defaultPricePerUnit ПО КАТЕГОРИЯМ
-- =====================================================

-- Protein - 0.02 PLN за грамм (примерно 20 PLN/кг)
UPDATE "Ingredient" 
SET "defaultPricePerUnit" = 0.02
WHERE "defaultPricePerUnit" IS NULL 
  AND "category" = 'protein';

-- Vegetable, Grain, Dairy - 0.01 PLN за грамм/мл (10 PLN/кг)
UPDATE "Ingredient" 
SET "defaultPricePerUnit" = 0.01
WHERE "defaultPricePerUnit" IS NULL 
  AND "category" IN ('vegetable', 'grain', 'dairy');

-- Condiment - 0.03 PLN за грамм (специи дороже)
UPDATE "Ingredient" 
SET "defaultPricePerUnit" = 0.03
WHERE "defaultPricePerUnit" IS NULL 
  AND "category" = 'condiment';

-- Other - 0.01 PLN за грамм
UPDATE "Ingredient" 
SET "defaultPricePerUnit" = 0.01
WHERE "defaultPricePerUnit" IS NULL 
  AND "category" = 'other';

-- =====================================================
-- ПРОВЕРКА РЕЗУЛЬТАТОВ
-- =====================================================

SELECT '=== AFTER UPDATES ===' as status;
SELECT 
    "category",
    COUNT(*) as total,
    COUNT(CASE WHEN "defaultShelfLifeDays" IS NULL THEN 1 END) as missing_shelf_life,
    COUNT(CASE WHEN "defaultPricePerUnit" IS NULL THEN 1 END) as missing_price,
    AVG("defaultShelfLifeDays") as avg_shelf_life,
    AVG("defaultPricePerUnit") as avg_price
FROM "Ingredient"
GROUP BY "category"
ORDER BY "category";

-- Итоговая статистика
SELECT '=== FINAL STATS ===' as status;
SELECT 
    COUNT(*) as total_ingredients,
    COUNT(CASE WHEN "defaultShelfLifeDays" IS NOT NULL THEN 1 END) as with_shelf_life,
    COUNT(CASE WHEN "defaultPricePerUnit" IS NOT NULL THEN 1 END) as with_price,
    COUNT(CASE WHEN "defaultShelfLifeDays" IS NULL THEN 1 END) as missing_shelf_life,
    COUNT(CASE WHEN "defaultPricePerUnit" IS NULL THEN 1 END) as missing_price
FROM "Ingredient";
