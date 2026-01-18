-- =====================================================
-- НОРМАЛИЗАЦИЯ CANONICAL NAMES (КРИТИЧНО)
-- =====================================================
-- Цель: Привести все canonicalName к английским slug'ам
-- Дата: 2026-01-18
-- =====================================================

BEGIN;

-- 1️⃣ Яичница → scrambled_eggs
UPDATE "Recipe" 
SET "canonicalName" = 'scrambled_eggs' 
WHERE "canonicalName" = 'яичница' 
   OR "canonicalName" = 'Scrambled Eggs';

-- 2️⃣ Жареный лосось → fried_salmon (объединяем 5 вариантов)
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

-- 3️⃣ Grilled salmon → grilled_salmon (уже правильно)
UPDATE "Recipe" 
SET "canonicalName" = 'grilled_salmon' 
WHERE "canonicalName" = 'grilled_salmon_with_jasmine_rice';

-- 4️⃣ Pierogi Ruskie → pierogi_ruskie (4 дубликата)
UPDATE "Recipe" 
SET "canonicalName" = 'pierogi_ruskie' 
WHERE "canonicalName" = 'Pierogi Ruskie' 
   OR "localName" = 'Pierogi ruskie'
   OR ("canonicalName" IS NULL AND "localName" = 'Pierogi ruskie');

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

-- =====================================================
-- УДАЛЕНИЕ ДУБЛИКАТОВ (оставляем только 1 на canonical)
-- =====================================================

-- Создаём временную таблицу с ID рецептов для удаления
CREATE TEMP TABLE recipes_to_delete AS
WITH ranked_recipes AS (
    SELECT 
        id,
        "canonicalName",
        ROW_NUMBER() OVER (
            PARTITION BY "canonicalName" 
            ORDER BY "createdAt" ASC  -- Оставляем самый старый
        ) as rn
    FROM "Recipe"
    WHERE "canonicalName" IS NOT NULL
)
SELECT id 
FROM ranked_recipes 
WHERE rn > 1;  -- Удаляем все дубликаты, кроме первого

-- Показываем, что будет удалено
SELECT 
    r.id,
    r."canonicalName",
    r."title",
    r."createdAt"
FROM "Recipe" r
JOIN recipes_to_delete d ON r.id = d.id
ORDER BY r."canonicalName", r."createdAt";

-- ⚠️ РАСКОММЕНТИРУЙ СЛЕДУЮЩУЮ СТРОКУ, КОГДА ПРОВЕРИШЬ СПИСОК ВЫШЕ
-- DELETE FROM "Recipe" WHERE id IN (SELECT id FROM recipes_to_delete);

-- =====================================================
-- ДОБАВИТЬ CONSTRAINT: canonicalName обязателен
-- =====================================================

-- ⚠️ ВЫПОЛНИТЬ ТОЛЬКО ПОСЛЕ УДАЛЕНИЯ ДУБЛИКАТОВ И ЗАПОЛНЕНИЯ ВСЕХ canonicalName
-- ALTER TABLE "Recipe" ALTER COLUMN "canonicalName" SET NOT NULL;
-- ALTER TABLE "Recipe" ADD CONSTRAINT "Recipe_canonicalName_unique" UNIQUE ("canonicalName");

COMMIT;

-- =====================================================
-- ПРОВЕРКА РЕЗУЛЬТАТА
-- =====================================================

-- Показать все уникальные canonical names
SELECT 
    "canonicalName",
    COUNT(*) as count,
    STRING_AGG("title", ', ') as titles
FROM "Recipe"
WHERE "canonicalName" IS NOT NULL
GROUP BY "canonicalName"
ORDER BY count DESC, "canonicalName";

-- Показать рецепты БЕЗ canonical name
SELECT 
    id,
    "title",
    "localName"
FROM "Recipe"
WHERE "canonicalName" IS NULL OR "canonicalName" = '';
