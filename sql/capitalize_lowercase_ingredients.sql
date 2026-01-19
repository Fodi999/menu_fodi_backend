-- ============================================
-- Очистка и капитализация lowercase ингредиентов
-- ============================================

BEGIN;

\echo '🔄 Начинаем обновление ингредиентов с маленькой буквы...'
\echo ''

-- 1️⃣ Обновляем name_pl (польский)
\echo '📝 1. Обновление name_pl (польский)...'
UPDATE "Ingredient"
SET name_pl = CONCAT(UPPER(SUBSTRING(name_pl, 1, 1)), SUBSTRING(name_pl, 2))
WHERE name_pl IS NOT NULL 
  AND name_pl != '' 
  AND SUBSTRING(name_pl, 1, 1) = LOWER(SUBSTRING(name_pl, 1, 1));

\echo '   ✅ Обновлено польских названий'

-- 2️⃣ Обновляем name_en (английский)
\echo '📝 2. Обновление name_en (английский)...'
UPDATE "Ingredient"
SET name_en = CONCAT(UPPER(SUBSTRING(name_en, 1, 1)), SUBSTRING(name_en, 2))
WHERE name_en IS NOT NULL 
  AND name_en != '' 
  AND SUBSTRING(name_en, 1, 1) = LOWER(SUBSTRING(name_en, 1, 1));

\echo '   ✅ Обновлено английских названий'

-- 3️⃣ Обновляем name_ru (русский)
\echo '📝 3. Обновление name_ru (русский)...'
UPDATE "Ingredient"
SET name_ru = CONCAT(UPPER(SUBSTRING(name_ru, 1, 1)), SUBSTRING(name_ru, 2))
WHERE name_ru IS NOT NULL 
  AND name_ru != '' 
  AND SUBSTRING(name_ru, 1, 1) = LOWER(SUBSTRING(name_ru, 1, 1));

\echo '   ✅ Обновлено русских названий'

-- 4️⃣ Обновляем legacy поле name (английский)
\echo '📝 4. Обновление legacy поля name...'
UPDATE "Ingredient"
SET name = CONCAT(UPPER(SUBSTRING(name, 1, 1)), SUBSTRING(name, 2))
WHERE name IS NOT NULL 
  AND name != '' 
  AND SUBSTRING(name, 1, 1) = LOWER(SUBSTRING(name, 1, 1));

\echo '   ✅ Обновлено legacy названий'

\echo ''
\echo '📊 Итоговая статистика:'

-- Проверяем результат
SELECT 
    'Осталось lowercase PL' as metric,
    COUNT(*) as value
FROM "Ingredient"
WHERE name_pl IS NOT NULL 
  AND name_pl != '' 
  AND SUBSTRING(name_pl, 1, 1) = LOWER(SUBSTRING(name_pl, 1, 1))
UNION ALL
SELECT 
    'Осталось lowercase EN',
    COUNT(*)
FROM "Ingredient"
WHERE name_en IS NOT NULL 
  AND name_en != '' 
  AND SUBSTRING(name_en, 1, 1) = LOWER(SUBSTRING(name_en, 1, 1))
UNION ALL
SELECT 
    'Осталось lowercase RU',
    COUNT(*)
FROM "Ingredient"
WHERE name_ru IS NOT NULL 
  AND name_ru != '' 
  AND SUBSTRING(name_ru, 1, 1) = LOWER(SUBSTRING(name_ru, 1, 1))
UNION ALL
SELECT 
    'Всего ингредиентов',
    COUNT(*)
FROM "Ingredient";

\echo ''
\echo '✅ Обновление завершено успешно!'
\echo ''
\echo '⚠️  ВНИМАНИЕ: Изменения еще не применены!'
\echo '   Для применения: COMMIT;'
\echo '   Для отмены: ROLLBACK;'
\echo ''

-- Не коммитим автоматически - пусть пользователь проверит
-- COMMIT;
