-- =====================================================
-- ОЧИСТКА И НОРМАЛИЗАЦИЯ КАТАЛОГА ИНГРЕДИЕНТОВ
-- =====================================================
-- Удаляет тестовые, дубликаты, устаревшие записи
-- Приводит к единому языку (PL)
-- Проверяет и исправляет defaultShelfLifeDays и unit
-- =====================================================

-- 1. УДАЛЕНИЕ ТЕСТОВЫХ ПРОДУКТОВ
-- =====================================================
DELETE FROM "Ingredient" 
WHERE "name" LIKE '%Тестов%' 
   OR "name" LIKE '%тест%'
   OR "name" = 'Тестовый лосось через API'
   OR "name" = 'Тестовый угорь';

-- 2. УДАЛЕНИЕ РУССКИХ ДУБЛИКАТОВ (оставляем только польские названия)
-- =====================================================

-- Удаляем русские записи про лосось (оставляем польский "Łosoś")
DELETE FROM "Ingredient" 
WHERE "name" IN (
    'Лосось свежий',
    'Лосось норвежский', 
    'Лосось Фермерский',
    'Лосось фермерский',
    'Лосось чилийский'
);

-- Удаляем русские записи про креветки (оставляем польский "Krewetki")
DELETE FROM "Ingredient" 
WHERE "name" IN (
    'Креветки Королевские',
    'Креветки тигровые'
);

-- Удаляем русские записи про рыбу
DELETE FROM "Ingredient" 
WHERE "name" IN (
    'Лещь',
    'Тунец',
    'Тунец Желтопёрый',
    'Тунец желтоперый',
    'Тунец свежий'
);

-- Удаляем русские базовые продукты (есть польские аналоги)
DELETE FROM "Ingredient" 
WHERE "name" IN (
    'Минеральная вода',
    'Мука',
    'Соль',
    'Яица'
);

-- 3. ИСПРАВЛЕНИЕ КАТЕГОРИЙ
-- =====================================================
-- Меняем категорию "other" на правильные для известных продуктов

-- Исправляем категории для продуктов, которые были с category="other"
-- но должны быть в правильных категориях

UPDATE "Ingredient" 
SET "category" = 'protein'
WHERE "category" = 'other' 
  AND "name" IN (
    'Креветки Королевские',
    'Креветки тигровые',
    'Лещь',
    'Лосось Фермерский',
    'Лосось норвежский',
    'Лосось свежий',
    'Лосось фермерский',
    'Лосось чилийский',
    'Тестовый лосось через API',
    'Тестовый угорь',
    'Тунец',
    'Тунец Желтопёрый',
    'Тунец желтоперый',
    'Тунец свежий'
  );

UPDATE "Ingredient" 
SET "category" = 'grain'
WHERE "category" = 'other' 
  AND "name" = 'Мука';

UPDATE "Ingredient" 
SET "category" = 'condiment'
WHERE "category" = 'other' 
  AND "name" = 'Соль';

UPDATE "Ingredient" 
SET "category" = 'protein'
WHERE "category" = 'other' 
  AND "name" = 'Яица';

-- 4. УСТАНОВКА defaultShelfLifeDays ДЛЯ ЗАПИСЕЙ БЕЗ НЕГО
-- =====================================================

-- Устанавливаем значения по умолчанию для каждой категории
UPDATE "Ingredient" 
SET "defaultShelfLifeDays" = 3
WHERE "defaultShelfLifeDays" IS NULL 
  AND "category" = 'protein';

UPDATE "Ingredient" 
SET "defaultShelfLifeDays" = 7
WHERE "defaultShelfLifeDays" IS NULL 
  AND "category" = 'vegetable';

UPDATE "Ingredient" 
SET "defaultShelfLifeDays" = 14
WHERE "defaultShelfLifeDays" IS NULL 
  AND "category" = 'dairy';

UPDATE "Ingredient" 
SET "defaultShelfLifeDays" = 365
WHERE "defaultShelfLifeDays" IS NULL 
  AND "category" = 'grain';

UPDATE "Ingredient" 
SET "defaultShelfLifeDays" = 365
WHERE "defaultShelfLifeDays" IS NULL 
  AND "category" = 'condiment';

UPDATE "Ingredient" 
SET "defaultShelfLifeDays" = 180
WHERE "defaultShelfLifeDays" IS NULL 
  AND "category" = 'other';

-- 5. УСТАНОВКА defaultPricePerUnit ДЛЯ ЗАПИСЕЙ БЕЗ НЕГО
-- =====================================================

UPDATE "Ingredient" 
SET "defaultPricePerUnit" = 0.02
WHERE "defaultPricePerUnit" IS NULL 
  AND "category" = 'protein';

UPDATE "Ingredient" 
SET "defaultPricePerUnit" = 0.01
WHERE "defaultPricePerUnit" IS NULL 
  AND "category" IN ('vegetable', 'grain', 'dairy');

UPDATE "Ingredient" 
SET "defaultPricePerUnit" = 0.03
WHERE "defaultPricePerUnit" IS NULL 
  AND "category" = 'condiment';

UPDATE "Ingredient" 
SET "defaultPricePerUnit" = 0.01
WHERE "defaultPricePerUnit" IS NULL 
  AND "category" = 'other';

-- 6. ПРОВЕРКА И ИСПРАВЛЕНИЕ ЕДИНИЦ ИЗМЕРЕНИЯ
-- =====================================================

-- Стандартизируем единицы измерения
-- g - граммы (по умолчанию для большинства)
-- ml - миллилитры (для жидкостей)
-- pcs - штуки (для штучных товаров)

-- Исправляем жидкости на ml
UPDATE "Ingredient" 
SET "unit" = 'ml'
WHERE "unit" = 'g' 
  AND ("name" LIKE '%Mleko%' 
    OR "name" LIKE '%Kefir%'
    OR "name" LIKE '%Maślanka%'
    OR "name" LIKE '%Śmietana%'
    OR "name" LIKE '%Olej%'
    OR "name" LIKE '%Oliwa%'
    OR "name" LIKE '%Ocet%'
    OR "name" LIKE '%Sos sojowy%'
    OR "name" LIKE '%Sos rybny%'
    OR "name" LIKE '%Sos teriyaki%'
    OR "name" LIKE '%Ekstrakt%'
    OR "name" LIKE '%woda%');

-- 7. ФИНАЛЬНАЯ ПРОВЕРКА - ВЫВОД СТАТИСТИКИ
-- =====================================================

-- Проверяем количество продуктов по категориям
SELECT 
    "category",
    COUNT(*) as count,
    COUNT(CASE WHEN "defaultShelfLifeDays" IS NULL THEN 1 END) as missing_shelf_life,
    COUNT(CASE WHEN "defaultPricePerUnit" IS NULL THEN 1 END) as missing_price
FROM "Ingredient"
GROUP BY "category"
ORDER BY "category";

-- Проверяем записи без обязательных полей
SELECT COUNT(*) as total_without_shelf_life
FROM "Ingredient" 
WHERE "defaultShelfLifeDays" IS NULL;

SELECT COUNT(*) as total_without_price
FROM "Ingredient" 
WHERE "defaultPricePerUnit" IS NULL;

-- Общая статистика
SELECT 
    COUNT(*) as total_ingredients,
    COUNT(DISTINCT "category") as total_categories,
    COUNT(DISTINCT "unit") as total_units
FROM "Ingredient";
