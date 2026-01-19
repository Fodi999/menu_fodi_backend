-- ============================================
-- Удаление дубликатов и тестовых ингредиентов
-- ============================================

BEGIN;

\echo '🗑️  Начинаем очистку дубликатов и тестовых данных...'
\echo ''

-- 1️⃣ Найдем явные дубликаты (ice cream x2, pomidor/tomato x2)
\echo '🔍 1. Дубликаты ice cream (lody):'
SELECT id, name_pl, name_en, name_ru, normalized_value, category, unit
FROM "Ingredient"
WHERE normalized_value IN ('icecream', 'ice cream')
   OR (name_pl = 'lody');

\echo ''
\echo '🔍 2. Дубликаты pomidor/tomato:'
SELECT id, name_pl, name_en, name_ru, normalized_value, category, unit
FROM "Ingredient"
WHERE normalized_value = 'pomidor'
   OR (name_pl = 'pomidor' OR name_pl = 'Pomidor')
   OR (name_en = 'tomato' AND name_pl IN ('pomidor', 'Pomidor'));

\echo ''
\echo '🔍 3. Дубликаты cucumber (ogórek):'
SELECT id, name_pl, name_en, name_ru, normalized_value, category, unit
FROM "Ingredient"
WHERE normalized_value IN ('cucumber', 'ogorek')
   OR name_pl IN ('Ogórek', 'ogórek szklarniowy');

\echo ''
\echo '🔍 4. Дубликаты cabbage (kapusta):'
SELECT id, name_pl, name_en, name_ru, normalized_value, category, unit
FROM "Ingredient"
WHERE normalized_value LIKE '%kapusta%'
   OR name_en = 'cabbage';

\echo ''
\echo '⚠️  ВНИМАНИЕ: Проверьте дубликаты выше!'
\echo ''
\echo '❓ Хотите удалить дубликаты и тестовые данные?'
\echo '   Если да, раскомментируйте DELETE запросы ниже'
\echo ''

-- ============================================
-- УДАЛЕНИЕ (раскомментируйте после проверки)
-- ============================================

-- 🗑️ Удаляем тестовый ингредиент "ingredient for deletion"
-- DELETE FROM "Ingredient"
-- WHERE id = 'c8e7141a-c0c0-45ce-80b0-2584fc38c6f7'
--    OR name_pl = 'składnik do usunięcia'
--    OR normalized_value = 'ingredient_for_deletion';

-- 🗑️ Удаляем дубликат ice cream (оставляем dairy/ml, удаляем other/g)
-- DELETE FROM "Ingredient"
-- WHERE id = '21110781-0f61-4862-b800-9aaafe4cac92'  -- other/g версия
--    OR (normalized_value = 'icecream' AND category = 'other');

-- 🗑️ Удаляем дубликат tomato/pomidor (оставляем один, удаляем другой)
-- -- Оставляем: fc57dbf2-39bb-4f30-a8e2-cf6585074587 (Pomidor -> tomato, pomidor)
-- -- Удаляем: fb61f17e-4c07-4a2f-9927-ac788d329e6d (pomidor -> tomato, tomato)
-- DELETE FROM "Ingredient"
-- WHERE id = 'fb61f17e-4c07-4a2f-9927-ac788d329e6d';

-- 🗑️ Удаляем лишний cucumber (оставляем огурец szklarniowy)
-- -- Оставляем: 2e1c5ba4-1fe4-4407-aaa7-3e5570294f9a (ogórek szklarniowy)
-- -- Удаляем: 59bf118a-9dae-4ca3-a262-776e18b58338 (Ogórek обычный)
-- DELETE FROM "Ingredient"
-- WHERE id = '59bf118a-9dae-4ca3-a262-776e18b58338';

\echo ''
\echo '📝 Скрипт готов к использованию.'
\echo ''
\echo '📋 Рекомендации:'
\echo '   1. Сначала запустите БЕЗ удаления (просмотрите дубликаты)'
\echo '   2. Раскомментируйте DELETE запросы'
\echo '   3. Запустите снова для удаления'
\echo '   4. Проверьте результат'
\echo '   5. COMMIT для применения или ROLLBACK для отмены'
\echo ''

-- Не коммитим автоматически
-- COMMIT;
