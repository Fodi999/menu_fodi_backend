-- =====================================================
-- BACKUP: Recipe Catalog Tables
-- Дата: 2026-01-18
-- Цель: Backup перед нормализацией canonical names
-- =====================================================

-- Создаём временные backup таблицы
CREATE TABLE IF NOT EXISTS backup_recipe_20260118 AS
SELECT * FROM "Recipe";

CREATE TABLE IF NOT EXISTS backup_catalog_ingredient_20260118 AS
SELECT * FROM "CatalogIngredient";

CREATE TABLE IF NOT EXISTS backup_recipe_ingredient_20260118 AS
SELECT * FROM "RecipeIngredient";

-- Проверка backup
SELECT 'Recipe backup' as table_name, COUNT(*) as records FROM backup_recipe_20260118
UNION ALL
SELECT 'CatalogIngredient backup', COUNT(*) FROM backup_catalog_ingredient_20260118
UNION ALL
SELECT 'RecipeIngredient backup', COUNT(*) FROM backup_recipe_ingredient_20260118;

-- Проверка canonical names перед миграцией
SELECT 
    'NULL or empty' as issue_type,
    COUNT(*) as count
FROM "Recipe"
WHERE "canonicalName" IS NULL OR "canonicalName" = ''
UNION ALL
SELECT 
    'Cyrillic (localized)',
    COUNT(*)
FROM "Recipe"
WHERE "canonicalName" ~ '[а-яА-ЯёЁ]'
UNION ALL
SELECT 
    'With spaces (not slug)',
    COUNT(*)
FROM "Recipe"
WHERE "canonicalName" ~ ' '
UNION ALL
SELECT 
    'Total recipes',
    COUNT(*)
FROM "Recipe";
