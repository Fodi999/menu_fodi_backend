-- Капитализация ингредиентов (с автоматическим COMMIT)

-- 1️⃣ name_pl
UPDATE "Ingredient"
SET name_pl = CONCAT(UPPER(SUBSTRING(name_pl, 1, 1)), SUBSTRING(name_pl, 2))
WHERE name_pl IS NOT NULL 
  AND name_pl != '' 
  AND SUBSTRING(name_pl, 1, 1) = LOWER(SUBSTRING(name_pl, 1, 1));

-- 2️⃣ name_en
UPDATE "Ingredient"
SET name_en = CONCAT(UPPER(SUBSTRING(name_en, 1, 1)), SUBSTRING(name_en, 2))
WHERE name_en IS NOT NULL 
  AND name_en != '' 
  AND SUBSTRING(name_en, 1, 1) = LOWER(SUBSTRING(name_en, 1, 1));

-- 3️⃣ name_ru
UPDATE "Ingredient"
SET name_ru = CONCAT(UPPER(SUBSTRING(name_ru, 1, 1)), SUBSTRING(name_ru, 2))
WHERE name_ru IS NOT NULL 
  AND name_ru != '' 
  AND SUBSTRING(name_ru, 1, 1) = LOWER(SUBSTRING(name_ru, 1, 1));

-- 4️⃣ name (legacy)
UPDATE "Ingredient"
SET name = CONCAT(UPPER(SUBSTRING(name, 1, 1)), SUBSTRING(name, 2))
WHERE name IS NOT NULL 
  AND name != '' 
  AND SUBSTRING(name, 1, 1) = LOWER(SUBSTRING(name, 1, 1));

-- Проверка
SELECT 
    'Lowercase PL' as metric,
    COUNT(*) as remaining
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
  AND SUBSTRING(name_ru, 1, 1) = LOWER(SUBSTRING(name_ru, 1, 1));
