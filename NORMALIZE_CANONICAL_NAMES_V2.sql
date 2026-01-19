-- =====================================================
-- НОРМАЛИЗАЦИЯ CANONICAL NAMES (ИСПРАВЛЕННАЯ ВЕРСИЯ)
-- =====================================================
-- Проблема: UNIQUE constraint уже есть, нужно сначала удалить дубликаты
-- Дата: 2026-01-18
-- =====================================================

BEGIN;

-- ✅ ШАГ 1: Временно удаляем UNIQUE constraint
ALTER TABLE "Recipe" DROP CONSTRAINT IF EXISTS "Recipe_canonicalName_key";

-- ✅ ШАГ 2: Нормализуем canonical names

-- 1️⃣ Яичница → scrambled_eggs
UPDATE "Recipe" 
SET "canonicalName" = 'scrambled_eggs' 
WHERE "canonicalName" = 'яичница' 
   OR "canonicalName" = 'Scrambled Eggs';

-- 2️⃣ Жареный лосось → fried_salmon (объединяем 6 вариантов)
UPDATE "Recipe" 
SET "canonicalName" = 'fried_salmon' 
WHERE "canonicalName" IN (
    'жареный_лосось',
    'лосось_жареный',
    'жареный_лосось_(микроскопический_тест)',
    'жареный_лосось_(реалистичный_тест)',
    'жареный_лосось_с_хрустящей_кожей',
    'домашний_рецепт_жареного_лосося'
);

-- 3️⃣ Grilled salmon → grilled_salmon
UPDATE "Recipe" 
SET "canonicalName" = 'grilled_salmon' 
WHERE "canonicalName" = 'grilled_salmon_with_jasmine_rice';

-- 4️⃣ Pierogi Ruskie → pierogi_ruskie
UPDATE "Recipe" 
SET "canonicalName" = 'pierogi_ruskie' 
WHERE "canonicalName" = 'Pierogi Ruskie' 
   OR "canonicalName" IS NULL AND "title" = 'Pierogi ruskie';

-- 5️⃣ Паста карбонара → pasta_carbonara
UPDATE "Recipe" 
SET "canonicalName" = 'pasta_carbonara' 
WHERE "canonicalName" = 'паста_карбонара_(авторский_рецепт)'
   OR "canonicalName" = 'Spaghetti Carbonara';

-- 6️⃣ Greek Salad → greek_salad
UPDATE "Recipe" 
SET "canonicalName" = 'greek_salad' 
WHERE "canonicalName" = 'Greek Salad';

-- 7️⃣ Polish Meat Dumplings → polish_meat_dumplings
UPDATE "Recipe" 
SET "canonicalName" = 'polish_meat_dumplings' 
WHERE "canonicalName" = 'Polish Meat Dumplings';

-- 8️⃣ Polish Chicken Soup → polish_chicken_soup
UPDATE "Recipe" 
SET "canonicalName" = 'polish_chicken_soup' 
WHERE "canonicalName" = 'Polish Chicken Soup';

-- 9️⃣ Pizza Margherita → pizza_margherita
UPDATE "Recipe" 
SET "canonicalName" = 'pizza_margherita' 
WHERE "canonicalName" = 'Pizza Margherita';

-- 🔟 Polish Hunters Stew → polish_hunters_stew
UPDATE "Recipe" 
SET "canonicalName" = 'polish_hunters_stew' 
WHERE "canonicalName" = 'Polish Hunters Stew';

-- 1️⃣1️⃣ Polish Breaded Pork Chop → polish_breaded_pork_chop
UPDATE "Recipe" 
SET "canonicalName" = 'polish_breaded_pork_chop' 
WHERE "canonicalName" = 'Polish Breaded Pork Chop';

-- 1️⃣2️⃣ Polish Potato Pancakes → polish_potato_pancakes
UPDATE "Recipe" 
SET "canonicalName" = 'polish_potato_pancakes' 
WHERE "canonicalName" = 'Polish Potato Pancakes';

-- ✅ ШАГ 3: Удаляем дубликаты (оставляем самый старый)

-- Создаём временную таблицу с ID дубликатов для удаления
CREATE TEMP TABLE recipes_to_delete AS
WITH ranked_recipes AS (
    SELECT 
        id,
        "canonicalName",
        "createdAt",
        ROW_NUMBER() OVER (
            PARTITION BY "canonicalName" 
            ORDER BY "createdAt" ASC NULLS LAST
        ) as rn
    FROM "Recipe"
    WHERE "canonicalName" IS NOT NULL AND "canonicalName" != ''
)
SELECT id, "canonicalName", "createdAt"
FROM ranked_recipes 
WHERE rn > 1;

-- Показываем что будет удалено
SELECT 
    'Будет удалено:' as status,
    COUNT(*) as count
FROM recipes_to_delete;

SELECT 
    r.id,
    r."canonicalName",
    r."title",
    r."createdAt"
FROM "Recipe" r
JOIN recipes_to_delete d ON r.id = d.id
ORDER BY r."canonicalName", r."createdAt";

-- ⚠️ УДАЛЕНИЕ (раскомментируй после проверки списка)
DELETE FROM "RecipeIngredient" WHERE "recipeId" IN (SELECT id FROM recipes_to_delete);
DELETE FROM "Recipe" WHERE id IN (SELECT id FROM recipes_to_delete);

-- ✅ ШАГ 4: Добавляем обратно constraints

-- NOT NULL constraint
ALTER TABLE "Recipe" ALTER COLUMN "canonicalName" SET NOT NULL;

-- UNIQUE constraint
ALTER TABLE "Recipe" ADD CONSTRAINT "Recipe_canonicalName_unique" UNIQUE ("canonicalName");

COMMIT;

-- =====================================================
-- ФИНАЛЬНАЯ ПРОВЕРКА
-- =====================================================

-- Показать все canonical names (должны быть уникальные, lowercase, с underscores)
SELECT 
    "canonicalName",
    COUNT(*) as count,
    STRING_AGG("title", ', ') as titles
FROM "Recipe"
GROUP BY "canonicalName"
ORDER BY "canonicalName";

-- Проверка что нет NULL
SELECT 
    'NULL canonical names' as check_type,
    COUNT(*) as count
FROM "Recipe"
WHERE "canonicalName" IS NULL OR "canonicalName" = '';

-- Проверка что нет кириллицы
SELECT 
    'Cyrillic canonical names' as check_type,
    COUNT(*) as count
FROM "Recipe"
WHERE "canonicalName" ~ '[а-яА-ЯёЁ]';

-- Проверка что нет пробелов
SELECT 
    'Canonical names with spaces' as check_type,
    COUNT(*) as count
FROM "Recipe"
WHERE "canonicalName" ~ ' ';

-- Итоговая статистика
SELECT 
    'Total recipes after cleanup' as metric,
    COUNT(*) as value
FROM "Recipe";
