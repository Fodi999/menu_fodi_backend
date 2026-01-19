-- =====================================================
-- DRY RUN: Проверка перед миграцией canonical names
-- =====================================================
-- Цель: Проверить какие рецепты будут изменены
-- Дата: 2026-01-18
-- =====================================================

-- 1️⃣ Проверка NULL значений
SELECT 
    id, 
    "localName",
    "canonicalName",
    category
FROM "Recipe" 
WHERE "canonicalName" IS NULL OR "canonicalName" = '';

-- 2️⃣ Проверка локализованных canonical names (кириллица)
SELECT 
    id,
    "localName",
    "canonicalName",
    category
FROM "Recipe"
WHERE "canonicalName" ~ '[а-яА-ЯёЁ]'; -- Regex для кириллицы

-- 3️⃣ Проверка дубликатов canonical names
SELECT 
    "canonicalName", 
    COUNT(*) as count,
    array_agg(id) as recipe_ids,
    array_agg("localName") as local_names
FROM "Recipe"
GROUP BY "canonicalName"
HAVING COUNT(*) > 1
ORDER BY count DESC;

-- 4️⃣ Список всех canonical names для визуального контроля
SELECT 
    "canonicalName",
    "localName",
    category,
    COUNT(*) OVER (PARTITION BY "canonicalName") as duplicates
FROM "Recipe"
ORDER BY "canonicalName", "localName";

-- 5️⃣ Подсчёт рецептов по типу canonical name
SELECT 
    CASE 
        WHEN "canonicalName" IS NULL THEN 'NULL'
        WHEN "canonicalName" = '' THEN 'EMPTY'
        WHEN "canonicalName" ~ '[а-яА-ЯёЁ]' THEN 'LOCALIZED (Cyrillic)'
        WHEN "canonicalName" ~ '[A-Za-z_]+' THEN 'ENGLISH SLUG (OK)'
        ELSE 'OTHER'
    END as canonical_type,
    COUNT(*) as count
FROM "Recipe"
GROUP BY canonical_type
ORDER BY count DESC;

-- 6️⃣ Предпросмотр изменений для яичницы
SELECT 
    id,
    "canonicalName" as old_canonical,
    'scrambled_eggs' as new_canonical,
    "localName"
FROM "Recipe"
WHERE "canonicalName" IN ('яичница', 'Scrambled Eggs');

-- 7️⃣ Предпросмотр изменений для лосося
SELECT 
    id,
    "canonicalName" as old_canonical,
    'fried_salmon' as new_canonical,
    "localName"
FROM "Recipe"
WHERE "canonicalName" IN (
    'жареный_лосось',
    'лосось_жареный',
    'жареный_лосось_(микроскопический_тест)',
    'жареный_лосось_(реалистичный_тест)',
    'жареный_лосось_с_хрустящей_кожей',
    'домашний_рецепт_жареного_лосося'
);

-- 8️⃣ Итоговая статистика
SELECT 
    'Total Recipes' as metric,
    COUNT(*) as value
FROM "Recipe"
UNION ALL
SELECT 
    'With NULL canonical',
    COUNT(*)
FROM "Recipe"
WHERE "canonicalName" IS NULL OR "canonicalName" = ''
UNION ALL
SELECT 
    'With Localized canonical',
    COUNT(*)
FROM "Recipe"
WHERE "canonicalName" ~ '[а-яА-ЯёЁ]'
UNION ALL
SELECT 
    'With English slug',
    COUNT(*)
FROM "Recipe"
WHERE "canonicalName" ~ '^[a-z_]+$'
UNION ALL
SELECT 
    'Duplicate canonical names',
    COUNT(DISTINCT "canonicalName")
FROM "Recipe"
GROUP BY "canonicalName"
HAVING COUNT(*) > 1;
