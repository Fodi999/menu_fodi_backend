-- Migration 062: Add translations for Russian-named ingredients
-- Date: 2026-01-06
-- Purpose: Add PL/EN translations for ingredients that have only Russian names
-- Note: This also includes cleanup of duplicate test entries

-- ========================================
-- SEAFOOD - Russian named ingredients
-- ========================================

-- King Shrimp (Креветки Королевские)
UPDATE "Ingredient" 
SET name_pl = 'Krewetki królewskie', 
    name_en = 'King prawns',
    name_ru = 'Креветки королевские',
    category = 'protein'
WHERE name = 'Креветки Королевские';

-- Tiger Shrimp (Креветки тигровые)
UPDATE "Ingredient" 
SET name_pl = 'Krewetki tygrysie', 
    name_en = 'Tiger prawns',
    name_ru = 'Креветки тигровые',
    category = 'protein'
WHERE name = 'Креветки тигровые';

-- Bream (Лещь)
UPDATE "Ingredient" 
SET name_pl = 'Leszcz', 
    name_en = 'Bream',
    name_ru = 'Лещ',
    category = 'protein'
WHERE name = 'Лещь';

-- Salmon - Farm (Лосось Фермерский) - keeping only one, delete duplicates
UPDATE "Ingredient" 
SET name_pl = 'Łosoś hodowlany', 
    name_en = 'Farm salmon',
    name_ru = 'Лосось фермерский',
    category = 'protein'
WHERE name = 'Лосось Фермерский' 
  AND id = (SELECT id FROM "Ingredient" WHERE name = 'Лосось Фермерский' LIMIT 1);

-- Salmon - Norwegian (Лосось норвежский)
UPDATE "Ingredient" 
SET name_pl = 'Łosoś norweski', 
    name_en = 'Norwegian salmon',
    name_ru = 'Лосось норвежский',
    category = 'protein'
WHERE name = 'Лосось норвежский'
  AND id = (SELECT id FROM "Ingredient" WHERE name = 'Лосось норвежский' LIMIT 1);

-- Salmon - Fresh (Лосось свежий)
UPDATE "Ingredient" 
SET name_pl = 'Łosoś świeży', 
    name_en = 'Fresh salmon',
    name_ru = 'Лосось свежий',
    category = 'protein'
WHERE name = 'Лосось свежий';

-- Salmon - Farm (variant spelling)
UPDATE "Ingredient" 
SET name_pl = 'Łosoś hodowlany', 
    name_en = 'Farm salmon',
    name_ru = 'Лосось фермерский',
    category = 'protein'
WHERE name = 'Лосось фермерский';

-- Salmon - Chilean (Лосось чилийский)
UPDATE "Ingredient" 
SET name_pl = 'Łosoś chilijski', 
    name_en = 'Chilean salmon',
    name_ru = 'Лосось чилийский',
    category = 'protein'
WHERE name = 'Лосось чилийский';

-- Tuna (Тунец)
UPDATE "Ingredient" 
SET name_pl = 'Tuńczyk', 
    name_en = 'Tuna',
    name_ru = 'Тунец',
    category = 'protein'
WHERE name = 'Тунец';

-- Tuna - Yellowfin (Тунец Желтопёрый)
UPDATE "Ingredient" 
SET name_pl = 'Tuńczyk żółtopłetwy', 
    name_en = 'Yellowfin tuna',
    name_ru = 'Тунец желтопёрый',
    category = 'protein'
WHERE name = 'Тунец Желтопёрый';

-- Tuna - Yellowfin (variant) - keep only one
UPDATE "Ingredient" 
SET name_pl = 'Tuńczyk żółtopłetwy', 
    name_en = 'Yellowfin tuna',
    name_ru = 'Тунец желтопёрый',
    category = 'protein'
WHERE name = 'Тунец желтоперый'
  AND id = (SELECT id FROM "Ingredient" WHERE name = 'Тунец желтоперый' LIMIT 1);

-- Tuna - Fresh (Тунец свежий)
UPDATE "Ingredient" 
SET name_pl = 'Tuńczyk świeży', 
    name_en = 'Fresh tuna',
    name_ru = 'Тунец свежий',
    category = 'protein'
WHERE name = 'Тунец свежий';

-- ========================================
-- OTHER BASIC INGREDIENTS
-- ========================================

-- Mineral water (Минеральная вода)
UPDATE "Ingredient" 
SET name_pl = 'Woda mineralna', 
    name_en = 'Mineral water',
    name_ru = 'Минеральная вода',
    category = 'other'
WHERE name = 'Минеральная вода';

-- Flour (Мука)
UPDATE "Ingredient" 
SET name_pl = 'Mąka', 
    name_en = 'Flour',
    name_ru = 'Мука',
    category = 'grain'
WHERE name = 'Мука';

-- Salt (Соль)
UPDATE "Ingredient" 
SET name_pl = 'Sól', 
    name_en = 'Salt',
    name_ru = 'Соль',
    category = 'condiment'
WHERE name = 'Соль';

-- Eggs (Яица - typo of Яйца)
UPDATE "Ingredient" 
SET name_pl = 'Jaja', 
    name_en = 'Eggs',
    name_ru = 'Яйца',
    category = 'protein'
WHERE name = 'Яица';

-- ========================================
-- DELETE TEST DATA AND DUPLICATES
-- ========================================

-- Delete test entries
DELETE FROM "Ingredient" WHERE name = 'Тестовый лосось через API';
DELETE FROM "Ingredient" WHERE name = 'Тестовый угорь';

-- Delete duplicate "Лосось Фермерский" (keep first one)
DELETE FROM "Ingredient" 
WHERE name = 'Лосось Фермерский' 
  AND id NOT IN (SELECT id FROM "Ingredient" WHERE name = 'Лосось Фермерский' LIMIT 1);

-- Delete duplicate "Лосось норвежский" (keep first one)
DELETE FROM "Ingredient" 
WHERE name = 'Лосось норвежский' 
  AND id NOT IN (SELECT id FROM "Ingredient" WHERE name = 'Лосось норвежский' LIMIT 1);

-- Delete duplicate "Тунец желтоперый" (keep first one)
DELETE FROM "Ingredient" 
WHERE name = 'Тунец желтоперый' 
  AND id NOT IN (SELECT id FROM "Ingredient" WHERE name = 'Тунец желтоперый' LIMIT 1);

-- ========================================
-- UPDATE normalized_value
-- ========================================

UPDATE "Ingredient"
SET normalized_value = LOWER(
    TRANSLATE(
        COALESCE(name_pl, name),
        'ąćęłńóśźżĄĆĘŁŃÓŚŹŻ',
        'acelnoszz ACELNOSZZ'
    )
)
WHERE name_en IS NOT NULL AND name_ru IS NOT NULL;

-- ========================================
-- VERIFICATION
-- ========================================

DO $$
DECLARE
    total_count INTEGER;
    with_en INTEGER;
    with_ru INTEGER;
    with_pl INTEGER;
    complete INTEGER;
    without_translations INTEGER;
BEGIN
    SELECT COUNT(*) INTO total_count FROM "Ingredient";
    SELECT COUNT(*) INTO with_en FROM "Ingredient" WHERE name_en IS NOT NULL;
    SELECT COUNT(*) INTO with_ru FROM "Ingredient" WHERE name_ru IS NOT NULL;
    SELECT COUNT(*) INTO with_pl FROM "Ingredient" WHERE name_pl IS NOT NULL;
    SELECT COUNT(*) INTO complete FROM "Ingredient" WHERE name_en IS NOT NULL AND name_ru IS NOT NULL AND name_pl IS NOT NULL;
    SELECT COUNT(*) INTO without_translations FROM "Ingredient" WHERE name_en IS NULL OR name_ru IS NULL OR name_pl IS NULL;
    
    RAISE NOTICE '========================================';
    RAISE NOTICE 'Migration 062 completed!';
    RAISE NOTICE '========================================';
    RAISE NOTICE 'Total ingredients: %', total_count;
    RAISE NOTICE 'With Polish: % (%.1f%%)', with_pl, (with_pl::float / total_count * 100);
    RAISE NOTICE 'With English: % (%.1f%%)', with_en, (with_en::float / total_count * 100);
    RAISE NOTICE 'With Russian: % (%.1f%%)', with_ru, (with_ru::float / total_count * 100);
    RAISE NOTICE 'Complete (all 3 languages): % (%.1f%%)', complete, (complete::float / total_count * 100);
    RAISE NOTICE 'Still incomplete: %', without_translations;
    RAISE NOTICE '========================================';
END $$;

-- Show sample of seafood ingredients
SELECT 
    name_pl as "Polish",
    name_en as "English", 
    name_ru as "Russian",
    category
FROM "Ingredient" 
WHERE name_pl LIKE 'Łosoś%' OR name_pl LIKE 'Tuńczyk%' OR name_pl LIKE 'Krewetki%'
ORDER BY category, name_pl
LIMIT 10;
