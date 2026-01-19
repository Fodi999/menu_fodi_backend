-- =====================================================
-- PERFORMANCE OPTIMIZATION: Recipe Queries
-- =====================================================
-- Проблема: SELECT * FROM "Recipe" ORDER BY "createdAt" DESC ~ 295ms
-- Решение: Индекс на сортировку
-- Ожидаемый результат: 10-20ms (вместо 295ms)
-- =====================================================

BEGIN;

-- ✅ Индекс для сортировки по createdAt (DESC)
-- Используется в:
-- - GET /api/recipes (список рецептов)
-- - GET /api/admin/recipes (админ панель)
-- - Любые запросы с ORDER BY "createdAt" DESC
CREATE INDEX IF NOT EXISTS idx_recipe_created_at_desc
ON "Recipe" ("createdAt" DESC);

-- ✅ Индекс для фильтрации по category
-- Используется в:
-- - GET /api/recipes?category=main
-- - Статистика по категориям
CREATE INDEX IF NOT EXISTS idx_recipe_category
ON "Recipe" (category);

-- ✅ Индекс для поиска по canonicalName (уже есть UNIQUE, но добавим для ясности)
-- Используется в:
-- - GET /api/recipes/:canonicalName
-- - AI Recommendation (поиск лучшего рецепта)
-- Примечание: UNIQUE constraint уже создаёт индекс, но явно укажем
-- CREATE INDEX IF NOT EXISTS idx_recipe_canonical_name
-- ON "Recipe" ("canonicalName");
-- ^ Не нужен, так как есть UNIQUE constraint

-- ✅ Композитный индекс для фильтрации + сортировки
-- Используется в:
-- - GET /api/recipes?category=main&sort=createdAt
CREATE INDEX IF NOT EXISTS idx_recipe_category_created_at
ON "Recipe" (category, "createdAt" DESC);

-- ✅ Индекс для фильтрации по difficulty
-- Используется в:
-- - GET /api/recipes?difficulty=easy
CREATE INDEX IF NOT EXISTS idx_recipe_difficulty
ON "Recipe" (difficulty);

COMMIT;

-- =====================================================
-- ПРОВЕРКА ИНДЕКСОВ
-- =====================================================

-- Список всех индексов на таблице Recipe
SELECT 
    indexname,
    indexdef
FROM pg_indexes
WHERE tablename = 'Recipe'
ORDER BY indexname;

-- Размер индексов (для мониторинга)
SELECT 
    schemaname,
    tablename,
    indexname,
    pg_size_pretty(pg_relation_size(indexrelid)) as index_size
FROM pg_stat_user_indexes
WHERE schemaname = 'public' AND tablename = 'Recipe'
ORDER BY pg_relation_size(indexrelid) DESC;

-- =====================================================
-- ТЕСТИРОВАНИЕ ПРОИЗВОДИТЕЛЬНОСТИ
-- =====================================================

-- Тест 1: Базовый запрос списка рецептов
EXPLAIN ANALYZE
SELECT * FROM "Recipe"
ORDER BY "createdAt" DESC
LIMIT 50;

-- Тест 2: Фильтрация по категории
EXPLAIN ANALYZE
SELECT * FROM "Recipe"
WHERE category = 'main'
ORDER BY "createdAt" DESC
LIMIT 50;

-- Тест 3: Поиск по canonicalName
EXPLAIN ANALYZE
SELECT * FROM "Recipe"
WHERE "canonicalName" = 'fried_salmon';

-- Тест 4: Count (пока быстро, но на будущее учесть)
EXPLAIN ANALYZE
SELECT COUNT(*) FROM "Recipe";

-- =====================================================
-- ОЖИДАЕМЫЕ РЕЗУЛЬТАТЫ
-- =====================================================

-- ДО оптимизации:
-- - SELECT ... ORDER BY "createdAt" DESC: ~295ms (first run), ~125ms (cached)
-- - Full table scan + sort in memory

-- ПОСЛЕ оптимизации:
-- - SELECT ... ORDER BY "createdAt" DESC: ~10-20ms
-- - Index scan (no sort needed)
-- - Latency стабильна независимо от размера таблицы

-- =====================================================
-- MAINTENANCE (на будущее)
-- =====================================================

-- Если таблица вырастет до 100k+ рецептов:

-- 1. Approximate COUNT (PostgreSQL 9.2+)
-- SELECT reltuples::bigint AS approximate_row_count
-- FROM pg_class
-- WHERE relname = 'Recipe';

-- 2. Materialized View для статистики
-- CREATE MATERIALIZED VIEW recipe_stats AS
-- SELECT 
--     COUNT(*) as total_recipes,
--     COUNT(*) FILTER (WHERE category = 'main') as main_count,
--     COUNT(*) FILTER (WHERE category = 'salad') as salad_count,
--     COUNT(*) FILTER (WHERE category = 'soup') as soup_count
-- FROM "Recipe";
-- 
-- REFRESH MATERIALIZED VIEW recipe_stats; -- run periodically

-- 3. Partitioning по category (если 1M+ рецептов)
-- CREATE TABLE "Recipe_main" PARTITION OF "Recipe" FOR VALUES IN ('main');
-- CREATE TABLE "Recipe_salad" PARTITION OF "Recipe" FOR VALUES IN ('salad');
