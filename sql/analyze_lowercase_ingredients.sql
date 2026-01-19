-- ============================================
-- Проверка и очистка ингредиентов с lowercase
-- и дубликатов по normalized_value
-- ============================================

-- 1️⃣ АНАЛИЗ: Найти ингредиенты с маленькой буквы
-- ============================================

\echo '🔍 1. Ингредиенты с маленькой буквы (PL):'
SELECT id, name_pl, name_en, name_ru, normalized_value, category, unit
FROM "Ingredient"
WHERE name_pl IS NOT NULL 
  AND name_pl != '' 
  AND SUBSTRING(name_pl, 1, 1) = LOWER(SUBSTRING(name_pl, 1, 1))
ORDER BY name_pl;

\echo ''
\echo '🔍 2. Ингредиенты с маленькой буквы (EN):'
SELECT id, name_pl, name_en, name_ru, normalized_value, category, unit
FROM "Ingredient"
WHERE name_en IS NOT NULL 
  AND name_en != '' 
  AND SUBSTRING(name_en, 1, 1) = LOWER(SUBSTRING(name_en, 1, 1))
ORDER BY name_en;

\echo ''
\echo '🔍 3. Ингредиенты с маленькой буквы (RU):'
SELECT id, name_pl, name_en, name_ru, normalized_value, category, unit
FROM "Ingredient"
WHERE name_ru IS NOT NULL 
  AND name_ru != '' 
  AND SUBSTRING(name_ru, 1, 1) = LOWER(SUBSTRING(name_ru, 1, 1))
ORDER BY name_ru;

-- 2️⃣ АНАЛИЗ: Найти дубликаты по normalized_value
-- ============================================

\echo ''
\echo '🔍 4. Дубликаты по normalized_value:'
SELECT 
    normalized_value,
    COUNT(*) as count,
    STRING_AGG(id::text, ', ') as ids,
    STRING_AGG(COALESCE(name_en, name_pl, name_ru, 'N/A'), ' / ') as names
FROM "Ingredient"
WHERE normalized_value IS NOT NULL
GROUP BY normalized_value
HAVING COUNT(*) > 1
ORDER BY count DESC, normalized_value;

-- 3️⃣ АНАЛИЗ: Статистика по auto_translated
-- ============================================

\echo ''
\echo '📊 5. Статистика по источнику:'
SELECT 
    auto_translated,
    COUNT(*) as count,
    ROUND(COUNT(*) * 100.0 / SUM(COUNT(*)) OVER (), 2) as percentage
FROM "Ingredient"
GROUP BY auto_translated
ORDER BY auto_translated DESC;

-- 4️⃣ АНАЛИЗ: Ингредиенты используемые в рецептах
-- ============================================

\echo ''
\echo '📊 6. Какие lowercase ингредиенты используются в рецептах:'
SELECT 
    i.id,
    i.name_en,
    i.name_pl,
    i.name_ru,
    i.normalized_value,
    COUNT(DISTINCT ci."recipeId") as recipe_count
FROM "Ingredient" i
LEFT JOIN "CatalogIngredient" ci ON i.id = ci."ingredientId"
WHERE (
    (i.name_pl IS NOT NULL AND i.name_pl != '' AND SUBSTRING(i.name_pl, 1, 1) = LOWER(SUBSTRING(i.name_pl, 1, 1)))
    OR
    (i.name_en IS NOT NULL AND i.name_en != '' AND SUBSTRING(i.name_en, 1, 1) = LOWER(SUBSTRING(i.name_en, 1, 1)))
    OR
    (i.name_ru IS NOT NULL AND i.name_ru != '' AND SUBSTRING(i.name_ru, 1, 1) = LOWER(SUBSTRING(i.name_ru, 1, 1)))
)
GROUP BY i.id, i.name_en, i.name_pl, i.name_ru, i.normalized_value
ORDER BY recipe_count DESC, i.name_en;

-- 5️⃣ АНАЛИЗ: Общая статистика
-- ============================================

\echo ''
\echo '📊 7. Общая статистика:'
SELECT 
    'Total ingredients' as metric,
    COUNT(*) as value
FROM "Ingredient"
UNION ALL
SELECT 
    'Lowercase PL',
    COUNT(*)
FROM "Ingredient"
WHERE name_pl IS NOT NULL 
  AND name_pl != '' 
  AND SUBSTRING(name_pl, 1, 1) = LOWER(SUBSTRING(name_pl, 1, 1))
UNION ALL
SELECT 
    'Lowercase EN',
    COUNT(*)
FROM "Ingredient"
WHERE name_en IS NOT NULL 
  AND name_en != '' 
  AND SUBSTRING(name_en, 1, 1) = LOWER(SUBSTRING(name_en, 1, 1))
UNION ALL
SELECT 
    'Lowercase RU',
    COUNT(*)
FROM "Ingredient"
WHERE name_ru IS NOT NULL 
  AND name_ru != '' 
  AND SUBSTRING(name_ru, 1, 1) = LOWER(SUBSTRING(name_ru, 1, 1))
UNION ALL
SELECT 
    'Duplicates by normalized_value',
    COUNT(DISTINCT normalized_value)
FROM "Ingredient"
WHERE normalized_value IN (
    SELECT normalized_value
    FROM "Ingredient"
    WHERE normalized_value IS NOT NULL
    GROUP BY normalized_value
    HAVING COUNT(*) > 1
);

\echo ''
\echo '✅ Анализ завершен. Проверьте результаты выше.'
\echo ''
\echo '⚠️  ВНИМАНИЕ: Перед удалением убедитесь, что:'
\echo '   1. Ингредиенты не используются в рецептах'
\echo '   2. У вас есть backup базы данных'
\echo '   3. Вы понимаете последствия удаления'
\echo ''
\echo '📝 Следующий шаг: Если нужно удалить, запустите:'
\echo '   psql "$DATABASE_URL" -f cleanup_lowercase_ingredients.sql'
