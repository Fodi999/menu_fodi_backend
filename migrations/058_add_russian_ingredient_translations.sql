-- Migration: Add Russian translations for common ingredients
-- Date: 2026-01-04

-- Meat & Proteins
UPDATE "Ingredient" SET name_ru = 'Говядина', name_en = 'Beef' WHERE name_pl = 'Wołowina (rostbef)';
UPDATE "Ingredient" SET name_ru = 'Свинина', name_en = 'Pork' WHERE name_pl ILIKE '%wieprzow%';
UPDATE "Ingredient" SET name_ru = 'Курица', name_en = 'Chicken' WHERE name_pl ILIKE '%kurczak%' OR name_pl ILIKE '%kura%';
UPDATE "Ingredient" SET name_ru = 'Индейка', name_en = 'Turkey' WHERE name_pl ILIKE '%indyk%';
UPDATE "Ingredient" SET name_ru = 'Рыба', name_en = 'Fish' WHERE name_pl ILIKE '%ryba%' AND name_ru IS NULL;
UPDATE "Ingredient" SET name_ru = 'Лосось', name_en = 'Salmon' WHERE name_pl ILIKE '%łosoś%';
UPDATE "Ingredient" SET name_ru = 'Тунец', name_en = 'Tuna' WHERE name_pl ILIKE '%tuńczyk%';
UPDATE "Ingredient" SET name_ru = 'Бекон', name_en = 'Bacon' WHERE name_pl ILIKE '%boczek%' OR name_pl ILIKE '%bacon%';

-- Dairy
UPDATE "Ingredient" SET name_ru = 'Молоко', name_en = 'Milk' WHERE name_pl ILIKE '%mleko%';
UPDATE "Ingredient" SET name_ru = 'Сыр', name_en = 'Cheese' WHERE name_pl = 'Ser' AND name_ru IS NULL;
UPDATE "Ingredient" SET name_ru = 'Сыр фета', name_en = 'Feta cheese' WHERE name_pl ILIKE '%feta%';
UPDATE "Ingredient" SET name_ru = 'Сыр пармезан', name_en = 'Parmesan' WHERE name_pl ILIKE '%parmezan%';
UPDATE "Ingredient" SET name_ru = 'Сыр моцарелла', name_en = 'Mozzarella' WHERE name_pl ILIKE '%mozzarella%';
UPDATE "Ingredient" SET name_ru = 'Сливки', name_en = 'Cream' WHERE name_pl ILIKE '%śmietana%' AND name_pl ILIKE '%30%';
UPDATE "Ingredient" SET name_ru = 'Сметана', name_en = 'Sour cream' WHERE name_pl ILIKE '%śmietana%' AND name_pl ILIKE '%18%';
UPDATE "Ingredient" SET name_ru = 'Йогурт', name_en = 'Yogurt' WHERE name_pl ILIKE '%jogurt%';

-- Grains & Pasta
UPDATE "Ingredient" SET name_ru = 'Мука', name_en = 'Flour' WHERE name_pl ILIKE '%mąka%';
UPDATE "Ingredient" SET name_ru = 'Рис', name_en = 'Rice' WHERE name_pl ILIKE '%ryż%';
UPDATE "Ingredient" SET name_ru = 'Макароны', name_en = 'Pasta' WHERE name_pl ILIKE '%makaron%' AND name_ru IS NULL;
UPDATE "Ingredient" SET name_ru = 'Хлеб', name_en = 'Bread' WHERE name_pl ILIKE '%chleb%' OR name_pl ILIKE '%bułka%';

-- Condiments & Spices
UPDATE "Ingredient" SET name_ru = 'Перец черный', name_en = 'Black pepper' WHERE name_pl ILIKE '%pieprz czarny%';
UPDATE "Ingredient" SET name_ru = 'Орегано', name_en = 'Oregano' WHERE name_pl ILIKE '%oregano%';
UPDATE "Ingredient" SET name_ru = 'Базилик', name_en = 'Basil' WHERE name_pl ILIKE '%bazylia%';
UPDATE "Ingredient" SET name_ru = 'Петрушка', name_en = 'Parsley' WHERE name_pl ILIKE '%pietruszka%';
UPDATE "Ingredient" SET name_ru = 'Укроп', name_en = 'Dill' WHERE name_pl ILIKE '%koper%';
UPDATE "Ingredient" SET name_ru = 'Сахар', name_en = 'Sugar' WHERE name_pl ILIKE '%cukier%';
UPDATE "Ingredient" SET name_ru = 'Уксус', name_en = 'Vinegar' WHERE name_pl ILIKE '%ocet%';
UPDATE "Ingredient" SET name_ru = 'Оливковое масло', name_en = 'Olive oil' WHERE name_pl ILIKE '%oliwa%' AND name_ru IS NULL;
UPDATE "Ingredient" SET name_ru = 'Растительное масло', name_en = 'Vegetable oil' WHERE name_pl ILIKE '%olej%' AND name_pl NOT ILIKE '%oliwa%';

-- Fruits
UPDATE "Ingredient" SET name_ru = 'Лимон', name_en = 'Lemon' WHERE name_pl ILIKE '%cytryna%';
UPDATE "Ingredient" SET name_ru = 'Яблоко', name_en = 'Apple' WHERE name_pl ILIKE '%jabłko%';
UPDATE "Ingredient" SET name_ru = 'Банан', name_en = 'Banana' WHERE name_pl ILIKE '%banan%';
UPDATE "Ingredient" SET name_ru = 'Апельсин', name_en = 'Orange' WHERE name_pl ILIKE '%pomarańcz%';

-- Vegetables (beyond what already exists)
UPDATE "Ingredient" SET name_ru = 'Салат', name_en = 'Lettuce' WHERE name_pl ILIKE '%sałata%';
UPDATE "Ingredient" SET name_ru = 'Шпинат', name_en = 'Spinach' WHERE name_pl ILIKE '%szpinak%';
UPDATE "Ingredient" SET name_ru = 'Грибы', name_en = 'Mushrooms' WHERE name_pl ILIKE '%pieczark%' OR name_pl ILIKE '%grzyb%';
UPDATE "Ingredient" SET name_ru = 'Кабачок', name_en = 'Zucchini' WHERE name_pl ILIKE '%cukinia%';
UPDATE "Ingredient" SET name_ru = 'Баклажан', name_en = 'Eggplant' WHERE name_pl ILIKE '%bakłażan%';
UPDATE "Ingredient" SET name_ru = 'Брокколи', name_en = 'Broccoli' WHERE name_pl ILIKE '%brokuł%';

-- Other common items
UPDATE "Ingredient" SET name_ru = 'Вода', name_en = 'Water' WHERE name_pl ILIKE '%woda%';
UPDATE "Ingredient" SET name_ru = 'Томатная паста', name_en = 'Tomato paste' WHERE name_pl ILIKE '%koncentrat pomidorowy%';
UPDATE "Ingredient" SET name_ru = 'Томатный соус', name_en = 'Tomato sauce' WHERE name_pl ILIKE '%passata%' OR (name_pl ILIKE '%sos%' AND name_pl ILIKE '%pomidor%');

-- Log completion
DO $$
DECLARE
    updated_count INT;
BEGIN
    SELECT COUNT(*) INTO updated_count FROM "Ingredient" WHERE name_ru IS NOT NULL;
    RAISE NOTICE 'Migration 058 completed. Total ingredients with Russian translations: %', updated_count;
END $$;
