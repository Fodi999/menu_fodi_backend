-- =====================================================
-- БЫСТРАЯ ОЧИСТКА КАТАЛОГА - ТОЛЬКО УДАЛЕНИЕ
-- =====================================================
-- Этот скрипт можно выполнить через Render Dashboard
-- или любой SQL клиент
-- =====================================================

-- Показываем текущее состояние ДО очистки
SELECT '=== BEFORE CLEANUP ===' as status;
SELECT 
    COUNT(*) as total_ingredients,
    COUNT(CASE WHEN "name" LIKE '%Тестов%' OR "name" LIKE '%тест%' THEN 1 END) as test_items,
    COUNT(CASE WHEN "name" LIKE '%Лосось%' OR "name" LIKE '%лосось%' THEN 1 END) as salmon_items,
    COUNT(CASE WHEN "name" LIKE '%Креветки%' THEN 1 END) as shrimp_items,
    COUNT(CASE WHEN "name" LIKE '%Тунец%' OR "name" LIKE '%тунец%' THEN 1 END) as tuna_items
FROM "Ingredient";

-- =====================================================
-- УДАЛЕНИЕ ТЕСТОВЫХ И ДУБЛИКАТОВ
-- =====================================================

-- 1. Удаление ВСЕХ тестовых продуктов
DELETE FROM "Ingredient" 
WHERE "name" LIKE '%Тестов%' 
   OR "name" LIKE '%тест%'
   OR "name" = 'Тестовый лосось через API'
   OR "name" = 'Тестовый угорь';

-- 2. Удаление ВСЕХ русских дубликатов лосося (оставляем только "Łosoś")
DELETE FROM "Ingredient" 
WHERE "name" IN (
    'Лосось свежий',
    'Лосось норвежский', 
    'Лосось Фермерский',
    'Лосось фермерский',
    'Лосось чилийский'
);

-- 3. Удаление ВСЕХ русских дубликатов креветок
DELETE FROM "Ingredient" 
WHERE "name" IN (
    'Креветки Королевские',
    'Креветки тигровые'
);

-- 4. Удаление ВСЕХ русских дубликатов тунца и рыбы
DELETE FROM "Ingredient" 
WHERE "name" IN (
    'Лещь',
    'Тунец',
    'Тунец Желтопёрый',
    'Тунец желтоперый',
    'Тунец свежий'
);

-- 5. Удаление русских базовых продуктов (есть польские версии)
DELETE FROM "Ingredient" 
WHERE "name" IN (
    'Минеральная вода',
    'Мука',
    'Соль',
    'Яица'
);

-- Показываем результат ПОСЛЕ очистки
SELECT '=== AFTER CLEANUP ===' as status;
SELECT 
    COUNT(*) as total_ingredients,
    COUNT(CASE WHEN "name" LIKE '%Тестов%' OR "name" LIKE '%тест%' THEN 1 END) as test_items,
    COUNT(CASE WHEN "name" LIKE '%Лосось%' OR "name" LIKE '%лосось%' THEN 1 END) as salmon_items,
    COUNT(CASE WHEN "name" LIKE '%Креветки%' THEN 1 END) as shrimp_items,
    COUNT(CASE WHEN "name" LIKE '%Тунец%' OR "name" LIKE '%тунец%' THEN 1 END) as tuna_items
FROM "Ingredient";

-- Показываем статистику по категориям
SELECT '=== CATEGORY STATS ===' as status;
SELECT 
    "category",
    COUNT(*) as count,
    COUNT(CASE WHEN "defaultShelfLifeDays" IS NULL THEN 1 END) as missing_shelf_life,
    COUNT(CASE WHEN "defaultPricePerUnit" IS NULL THEN 1 END) as missing_price
FROM "Ingredient"
GROUP BY "category"
ORDER BY "category";
