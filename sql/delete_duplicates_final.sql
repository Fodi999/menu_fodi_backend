-- ============================================
-- УДАЛЕНИЕ ДУБЛИКАТОВ - ФИНАЛЬНАЯ ВЕРСИЯ
-- ============================================

BEGIN;

\echo '🗑️  Удаление дубликатов и тестовых ингредиентов...'
\echo ''

-- 1️⃣ Удаляем тестовый ингредиент "ingredient for deletion"
\echo '1. Удаление тестового ингредиента "ingredient for deletion"...'
DELETE FROM "Ingredient"
WHERE id = 'c8e7141a-c0c0-45ce-80b0-2584fc38c6f7';
\echo '   ✅ Удалено'

-- 2️⃣ Ice cream дубликат (оставляем dairy/ml, удаляем other/g)
\echo '2. Удаление дубликата ice cream (other/g)...'
DELETE FROM "Ingredient"
WHERE id = '21110781-0f61-4862-b800-9aaafe4cac92';  -- icecream, other, g
\echo '   ✅ Удалено (оставлен: ice cream, dairy, ml)'

-- 3️⃣ Tomato/Pomidor дубликат (оставляем с normalized tomato)
\echo '3. Удаление дубликата pomidor/tomato...'
DELETE FROM "Ingredient"
WHERE id = 'fc57dbf2-39bb-4f30-a8e2-cf6585074587';  -- Pomidor -> pomidor
\echo '   ✅ Удалено (оставлен: pomidor -> tomato)'

-- 4️⃣ Cucumber дубликат (оставляем ogórek szklarniowy)
\echo '4. Удаление дубликата огурец...'
DELETE FROM "Ingredient"
WHERE id = '59bf118a-9dae-4ca3-a262-776e18b58338';  -- Ogórek -> ogorek
\echo '   ✅ Удалено (оставлен: ogórek szklarniowy -> cucumber)'

\echo ''
\echo '📊 Итоговая статистика после удаления:'
SELECT COUNT(*) as total_ingredients
FROM "Ingredient";

\echo ''
\echo '✅ Удаление завершено!'
\echo ''
\echo '⚠️  ВНИМАНИЕ: Для применения изменений выполните: COMMIT;'
\echo '   Для отмены: ROLLBACK;'
\echo ''

-- Оставляем пользователю решить
-- COMMIT;
