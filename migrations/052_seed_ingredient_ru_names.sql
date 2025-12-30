-- Seed Russian (RU) names for common ingredients
-- Date: 2025-12-30
-- Description: Adds Russian translations for most common ingredients

-- Vegetables / Овощи
UPDATE "Ingredient" SET 
    name_ru = 'помидор',
    name_en = 'tomato',
    normalized_value = 'pomidor'
WHERE name_pl = 'pomidor' OR name = 'pomidor';

UPDATE "Ingredient" SET 
    name_ru = 'огурец',
    name_en = 'cucumber',
    normalized_value = 'ogurek'
WHERE name_pl = 'ogórek' OR name = 'ogórek';

UPDATE "Ingredient" SET 
    name_ru = 'лук',
    name_en = 'onion',
    normalized_value = 'cebula'
WHERE name_pl = 'cebula' OR name = 'cebula';

UPDATE "Ingredient" SET 
    name_ru = 'чеснок',
    name_en = 'garlic',
    normalized_value = 'czosnek'
WHERE name_pl = 'czosnek' OR name = 'czosnek';

UPDATE "Ingredient" SET 
    name_ru = 'картофель',
    name_en = 'potato',
    normalized_value = 'ziemniak'
WHERE name_pl = 'ziemniak' OR name = 'ziemniak';

UPDATE "Ingredient" SET 
    name_ru = 'морковь',
    name_en = 'carrot',
    normalized_value = 'marchew'
WHERE name_pl = 'marchew' OR name = 'marchew';

UPDATE "Ingredient" SET 
    name_ru = 'перец',
    name_en = 'pepper',
    normalized_value = 'papryka'
WHERE name_pl = 'papryka' OR name = 'papryka';

-- Protein / Белки
UPDATE "Ingredient" SET 
    name_ru = 'курица',
    name_en = 'chicken',
    normalized_value = 'kurczak'
WHERE name_pl = 'kurczak' OR name = 'kurczak';

UPDATE "Ingredient" SET 
    name_ru = 'говядина',
    name_en = 'beef',
    normalized_value = 'wolowina'
WHERE name_pl = 'wołowina' OR name = 'wołowina';

UPDATE "Ingredient" SET 
    name_ru = 'свинина',
    name_en = 'pork',
    normalized_value = 'wieprzowina'
WHERE name_pl = 'wieprzowina' OR name = 'wieprzowina';

UPDATE "Ingredient" SET 
    name_ru = 'яйцо',
    name_en = 'egg',
    normalized_value = 'jajko'
WHERE name_pl = 'jajko' OR name = 'jajko';

UPDATE "Ingredient" SET 
    name_ru = 'рыба',
    name_en = 'fish',
    normalized_value = 'ryba'
WHERE name_pl = 'ryba' OR name = 'ryba';

-- Dairy / Молочные продукты
UPDATE "Ingredient" SET 
    name_ru = 'молоко',
    name_en = 'milk',
    normalized_value = 'mleko'
WHERE name_pl = 'mleko' OR name = 'mleko';

UPDATE "Ingredient" SET 
    name_ru = 'сыр',
    name_en = 'cheese',
    normalized_value = 'ser'
WHERE name_pl = 'ser' OR name = 'ser';

UPDATE "Ingredient" SET 
    name_ru = 'масло',
    name_en = 'butter',
    normalized_value = 'maslo'
WHERE name_pl = 'masło' OR name = 'masło';

UPDATE "Ingredient" SET 
    name_ru = 'сметана',
    name_en = 'sour cream',
    normalized_value = 'smietana'
WHERE name_pl = 'śmietana' OR name = 'śmietana';

UPDATE "Ingredient" SET 
    name_ru = 'йогурт',
    name_en = 'yogurt',
    normalized_value = 'jogurt'
WHERE name_pl = 'jogurt' OR name = 'jogurt';

-- Grains / Крупы
UPDATE "Ingredient" SET 
    name_ru = 'рис',
    name_en = 'rice',
    normalized_value = 'ryz'
WHERE name_pl = 'ryż' OR name = 'ryż';

UPDATE "Ingredient" SET 
    name_ru = 'макароны',
    name_en = 'pasta',
    normalized_value = 'makaron'
WHERE name_pl = 'makaron' OR name = 'makaron';

UPDATE "Ingredient" SET 
    name_ru = 'хлеб',
    name_en = 'bread',
    normalized_value = 'chleb'
WHERE name_pl = 'chleb' OR name = 'chleb';

UPDATE "Ingredient" SET 
    name_ru = 'мука',
    name_en = 'flour',
    normalized_value = 'maka'
WHERE name_pl = 'mąka' OR name = 'mąka';

-- Condiments / Специи и соусы
UPDATE "Ingredient" SET 
    name_ru = 'соль',
    name_en = 'salt',
    normalized_value = 'sol'
WHERE name_pl = 'sól' OR name = 'sól';

UPDATE "Ingredient" SET 
    name_ru = 'перец',
    name_en = 'pepper',
    normalized_value = 'pieprz'
WHERE name_pl = 'pieprz' OR name = 'pieprz';

UPDATE "Ingredient" SET 
    name_ru = 'сахар',
    name_en = 'sugar',
    normalized_value = 'cukier'
WHERE name_pl = 'cukier' OR name = 'cukier';

UPDATE "Ingredient" SET 
    name_ru = 'растительное масло',
    name_en = 'oil',
    normalized_value = 'olej'
WHERE name_pl = 'olej' OR name = 'olej';

UPDATE "Ingredient" SET 
    name_ru = 'уксус',
    name_en = 'vinegar',
    normalized_value = 'ocet'
WHERE name_pl = 'ocet' OR name = 'ocet';

-- Update any remaining ingredients with normalized values
UPDATE "Ingredient"
SET normalized_value = LOWER(
    TRANSLATE(
        COALESCE(name_pl, name),
        'ąćęłńóśźżĄĆĘŁŃÓŚŹŻ',
        'acelnoszz ACELNOSZZ'
    )
)
WHERE normalized_value IS NULL;
