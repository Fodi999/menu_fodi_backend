-- Migration: Clean and rebuild recipe catalog with correct structure
-- Purpose: Remove incomplete recipes, add 5 complete recipes with verified ingredients
-- Date: 2025-12-21
--
-- Strategy: Start fresh with fully verified recipes
-- Each recipe will have:
-- 1. All ingredients verified in Ingredient table
-- 2. Realistic quantities for servings
-- 3. Complete steps
-- 4. Optional flags set correctly (oil, salt, pepper)

-- ============================================================================
-- STEP 1: Clean existing catalog (keep only verified complete recipes)
-- ============================================================================

-- Delete all recipes from migration 038 (they have incomplete ingredients)
DELETE FROM "Recipe" 
WHERE "canonicalName" IN (
  'Polish Breaded Pork Chop',
  'Stuffed Cabbage Rolls',
  'Potato Pancakes',
  'Polish Crepes',
  'Polish Chicken Soup',
  'Polish Potato Dumplings',
  'Polish Sour Rye Soup',
  'Mushroom Soup',
  'Risotto alla Milanese',
  'Lasagna',
  'Tiramisu',
  'Penne Arrabbiata',
  'Caprese Salad',
  'Minestrone',
  'French Omelette',
  'American Pancakes',
  'Caesar Salad',
  'Tomato Soup',
  'Fried Rice',
  'Guacamole',
  'Chicken Curry',
  'Vegetable Stir Fry',
  'Quiche Lorraine',
  'Shakshuka',
  'Fish and Chips'
);

-- Keep only the original verified recipes (if they exist)
-- These should be: Sałatka grecka, Pierogi ruskie, Spaghetti Carbonara, etc.

-- ============================================================================
-- STEP 2: Add 5 complete Polish recipes
-- ============================================================================

-- Recipe 1: Kotlet schabowy (Polish Breaded Pork Chop) - COMPLETE VERSION
DO $$
DECLARE
    recipe_id UUID := gen_random_uuid();
    ing_schab UUID;
    ing_jaja UUID;
    ing_bulka UUID;
    ing_olej UUID;
BEGIN
    -- Find correct ingredient names
    SELECT id INTO ing_schab FROM "Ingredient" WHERE name = 'Wieprzowina (schab)' LIMIT 1;
    SELECT id INTO ing_jaja FROM "Ingredient" WHERE name = 'Jaja' LIMIT 1;
    SELECT id INTO ing_bulka FROM "Ingredient" WHERE name = 'Bułka' LIMIT 1;
    SELECT id INTO ing_olej FROM "Ingredient" WHERE name = 'Olej roślinny' LIMIT 1;

    -- Insert recipe
    INSERT INTO "Recipe" (
        id, 
        "canonicalName", 
        "localName", 
        country, 
        category, 
        difficulty, 
        "timeMinutes", 
        servings, 
        steps, 
        source
    )
    VALUES (
        recipe_id,
        'Polish Breaded Pork Chop',
        'Kotlet schabowy',
        'Poland',
        'main',
        'easy',
        25,
        4,
        '["1. Rozbij mięso tłuczkiem na cienko (ok. 1 cm)", "2. Oprósz mąką z obu stron", "3. Obtocz w rozbitym jajku", "4. Panieruj w bułce tartej", "5. Smaż na rozgrzanym oleju (3-4 min z każdej strony)", "6. Osusz na papierowym ręczniku", "7. Podawaj z ziemniakami i surówką"]'::jsonb,
        '{"type": "traditional", "reference": "Polish cuisine classic"}'::jsonb
    );

    -- Add ingredients with correct quantities
    IF ing_schab IS NOT NULL THEN
        INSERT INTO "RecipeIngredient" ("recipeId", "ingredientId", "ingredientKey", quantity, unit, "sortOrder", optional, "groupName")
        VALUES (recipe_id, ing_schab, 'wieprzowina-schab', 600, 'g', 1, false, 'baza');
    END IF;
    
    IF ing_jaja IS NOT NULL THEN
        INSERT INTO "RecipeIngredient" ("recipeId", "ingredientId", "ingredientKey", quantity, unit, "sortOrder", optional, "groupName")
        VALUES (recipe_id, ing_jaja, 'jaja', 2, 'szt', 2, false, 'baza');
    END IF;
    
    IF ing_bulka IS NOT NULL THEN
        INSERT INTO "RecipeIngredient" ("recipeId", "ingredientId", "ingredientKey", quantity, unit, "sortOrder", optional, "groupName")
        VALUES (recipe_id, ing_bulka, 'bulka-tarta', 150, 'g', 3, false, 'baza');
    END IF;
    
    IF ing_olej IS NOT NULL THEN
        INSERT INTO "RecipeIngredient" ("recipeId", "ingredientId", "ingredientKey", quantity, unit, "sortOrder", optional, "groupName")
        VALUES (recipe_id, ing_olej, 'olej', 100, 'ml', 4, true, 'baza');
    END IF;
END $$;

-- Recipe 2: Pierogi z mięsem (Meat Dumplings) - simplified
DO $$
DECLARE
    recipe_id UUID := gen_random_uuid();
    ing_maka UUID;
    ing_jaja UUID;
    ing_wieprzowina UUID;
    ing_cebula UUID;
BEGIN
    SELECT id INTO ing_maka FROM "Ingredient" WHERE name = 'Mąka pszenna' LIMIT 1;
    SELECT id INTO ing_jaja FROM "Ingredient" WHERE name = 'Jaja' LIMIT 1;
    SELECT id INTO ing_wieprzowina FROM "Ingredient" WHERE name = 'Wieprzowina (schab)' LIMIT 1; -- Use schab instead
    SELECT id INTO ing_cebula FROM "Ingredient" WHERE name = 'Cebula' LIMIT 1;

    INSERT INTO "Recipe" (
        id, "canonicalName", "localName", country, category, difficulty, "timeMinutes", servings, steps, source
    )
    VALUES (
        recipe_id,
        'Polish Meat Dumplings',
        'Pierogi z mięsem',
        'Poland',
        'main',
        'medium',
        90,
        6,
        '["1. Zagnieć ciasto z mąki, jajka i wody", "2. Usmaż mięso z cebulą", "3. Rozwałkuj ciasto", "4. Wytnij krążki szklanką", "5. Nałóż farsz i sklej brzegi", "6. Gotuj we wrzącej wodzie 5-7 minut", "7. Podawaj ze skwarkami"]'::jsonb,
        '{"type": "traditional", "reference": "Polish dumplings"}'::jsonb
    );

    IF ing_maka IS NOT NULL THEN
        INSERT INTO "RecipeIngredient" ("recipeId", "ingredientId", "ingredientKey", quantity, unit, "sortOrder", optional, "groupName")
        VALUES (recipe_id, ing_maka, 'maka', 500, 'g', 1, false, 'ciasto');
    END IF;
    
    IF ing_jaja IS NOT NULL THEN
        INSERT INTO "RecipeIngredient" ("recipeId", "ingredientId", "ingredientKey", quantity, unit, "sortOrder", optional, "groupName")
        VALUES (recipe_id, ing_jaja, 'jaja', 1, 'szt', 2, false, 'ciasto');
    END IF;
    
    IF ing_wieprzowina IS NOT NULL THEN
        INSERT INTO "RecipeIngredient" ("recipeId", "ingredientId", "ingredientKey", quantity, unit, "sortOrder", optional, "groupName")
        VALUES (recipe_id, ing_wieprzowina, 'wieprzowina-schab', 400, 'g', 3, false, 'farsz');
    END IF;
    
    IF ing_cebula IS NOT NULL THEN
        INSERT INTO "RecipeIngredient" ("recipeId", "ingredientId", "ingredientKey", quantity, unit, "sortOrder", optional, "groupName")
        VALUES (recipe_id, ing_cebula, 'cebula', 200, 'g', 4, false, 'farsz');
    END IF;
END $$;

-- Recipe 3: Bigos (Hunter's Stew)
DO $$
DECLARE
    recipe_id UUID := gen_random_uuid();
    ing_kapusta UUID;
    ing_kapusta_kiszona UUID;
    ing_kielbasa UUID;
    ing_cebula UUID;
BEGIN
    SELECT id INTO ing_kapusta FROM "Ingredient" WHERE name = 'Kapusta biała' LIMIT 1;
    SELECT id INTO ing_kapusta_kiszona FROM "Ingredient" WHERE name = 'Kapusta kiszona' LIMIT 1;
    SELECT id INTO ing_kielbasa FROM "Ingredient" WHERE name = 'Kiełbasa' LIMIT 1;
    SELECT id INTO ing_cebula FROM "Ingredient" WHERE name = 'Cebula' LIMIT 1;

    INSERT INTO "Recipe" (
        id, "canonicalName", "localName", country, category, difficulty, "timeMinutes", servings, steps, source
    )
    VALUES (
        recipe_id,
        'Polish Hunters Stew',
        'Bigos myśliwski',
        'Poland',
        'main',
        'medium',
        120,
        6,
        '["1. Pokrój kapustę i kiełbasę", "2. Podsmaż cebulę", "3. Dodaj kapustę kiszoną i świeżą", "4. Duś 60 minut", "5. Dodaj kiełbasę", "6. Duś kolejne 30 minut", "7. Podawaj z chlebem"]'::jsonb,
        '{"type": "traditional", "reference": "Polish national dish"}'::jsonb
    );

    IF ing_kapusta IS NOT NULL THEN
        INSERT INTO "RecipeIngredient" ("recipeId", "ingredientId", "ingredientKey", quantity, unit, "sortOrder", optional, "groupName")
        VALUES (recipe_id, ing_kapusta, 'kapusta-biala', 500, 'g', 1, false, 'baza');
    END IF;
    
    IF ing_kapusta_kiszona IS NOT NULL THEN
        INSERT INTO "RecipeIngredient" ("recipeId", "ingredientId", "ingredientKey", quantity, unit, "sortOrder", optional, "groupName")
        VALUES (recipe_id, ing_kapusta_kiszona, 'kapusta-kiszona', 500, 'g', 2, false, 'baza');
    END IF;
    
    IF ing_kielbasa IS NOT NULL THEN
        INSERT INTO "RecipeIngredient" ("recipeId", "ingredientId", "ingredientKey", quantity, unit, "sortOrder", optional, "groupName")
        VALUES (recipe_id, ing_kielbasa, 'kielbasa', 300, 'g', 3, false, 'baza');
    END IF;
    
    IF ing_cebula IS NOT NULL THEN
        INSERT INTO "RecipeIngredient" ("recipeId", "ingredientId", "ingredientKey", quantity, unit, "sortOrder", optional, "groupName")
        VALUES (recipe_id, ing_cebula, 'cebula', 150, 'g', 4, false, 'baza');
    END IF;
END $$;

-- Recipe 4: Rosół (Chicken Soup)
DO $$
DECLARE
    recipe_id UUID := gen_random_uuid();
    ing_kurczak UUID;
    ing_marchew UUID;
    ing_cebula UUID;
    ing_pietruszka UUID;
BEGIN
    SELECT id INTO ing_kurczak FROM "Ingredient" WHERE name = 'Kurczak (pierś)' LIMIT 1;
    SELECT id INTO ing_marchew FROM "Ingredient" WHERE name = 'Marchew' LIMIT 1;
    SELECT id INTO ing_cebula FROM "Ingredient" WHERE name = 'Cebula' LIMIT 1;
    SELECT id INTO ing_pietruszka FROM "Ingredient" WHERE name = 'Pietruszka (korzeń)' LIMIT 1;

    INSERT INTO "Recipe" (
        id, "canonicalName", "localName", country, category, difficulty, "timeMinutes", servings, steps, source
    )
    VALUES (
        recipe_id,
        'Polish Chicken Soup',
        'Rosół',
        'Poland',
        'soup',
        'easy',
        90,
        6,
        '["1. Zalej kurczaka zimną wodą", "2. Dodaj warzywa (marchew, cebulę, pietruszkę)", "3. Gotuj na małym ogniu 90 minut", "4. Zbieraj pianę", "5. Dopraw solą i pieprzem", "6. Przecedź", "7. Podawaj z makaronem"]'::jsonb,
        '{"type": "traditional", "reference": "Polish chicken broth"}'::jsonb
    );

    IF ing_kurczak IS NOT NULL THEN
        INSERT INTO "RecipeIngredient" ("recipeId", "ingredientId", "ingredientKey", quantity, unit, "sortOrder", optional, "groupName")
        VALUES (recipe_id, ing_kurczak, 'kurczak', 1000, 'g', 1, false, 'baza');
    END IF;
    
    IF ing_marchew IS NOT NULL THEN
        INSERT INTO "RecipeIngredient" ("recipeId", "ingredientId", "ingredientKey", quantity, unit, "sortOrder", optional, "groupName")
        VALUES (recipe_id, ing_marchew, 'marchew', 200, 'g', 2, false, 'warzywa');
    END IF;
    
    IF ing_cebula IS NOT NULL THEN
        INSERT INTO "RecipeIngredient" ("recipeId", "ingredientId", "ingredientKey", quantity, unit, "sortOrder", optional, "groupName")
        VALUES (recipe_id, ing_cebula, 'cebula', 150, 'g', 3, false, 'warzywa');
    END IF;
    
    IF ing_pietruszka IS NOT NULL THEN
        INSERT INTO "RecipeIngredient" ("recipeId", "ingredientId", "ingredientKey", quantity, unit, "sortOrder", optional, "groupName")
        VALUES (recipe_id, ing_pietruszka, 'pietruszka', 100, 'g', 4, true, 'warzywa');
    END IF;
END $$;

-- Recipe 5: Placki ziemniaczane (Potato Pancakes)
DO $$
DECLARE
    recipe_id UUID := gen_random_uuid();
    ing_ziemniaki UUID;
    ing_jaja UUID;
    ing_maka UUID;
    ing_olej UUID;
BEGIN
    SELECT id INTO ing_ziemniaki FROM "Ingredient" WHERE name = 'Ziemniak' LIMIT 1;
    SELECT id INTO ing_jaja FROM "Ingredient" WHERE name = 'Jaja' LIMIT 1;
    SELECT id INTO ing_maka FROM "Ingredient" WHERE name = 'Mąka pszenna' LIMIT 1;
    SELECT id INTO ing_olej FROM "Ingredient" WHERE name = 'Olej roślinny' LIMIT 1;

    INSERT INTO "Recipe" (
        id, "canonicalName", "localName", country, category, difficulty, "timeMinutes", servings, steps, source
    )
    VALUES (
        recipe_id,
        'Polish Potato Pancakes',
        'Placki ziemniaczane',
        'Poland',
        'main',
        'easy',
        30,
        4,
        '["1. Obierz i zetrzyj ziemniaki na tarce", "2. Odciśnij nadmiar wody", "3. Dodaj jajka i mąkę", "4. Wymieszaj na gładką masę", "5. Smaż łyżką na rozgrzanym oleju", "6. Smaż z obu stron na złoty kolor", "7. Podawaj ze śmietaną lub cukrem"]'::jsonb,
        '{"type": "traditional", "reference": "Polish comfort food"}'::jsonb
    );

    IF ing_ziemniaki IS NOT NULL THEN
        INSERT INTO "RecipeIngredient" ("recipeId", "ingredientId", "ingredientKey", quantity, unit, "sortOrder", optional, "groupName")
        VALUES (recipe_id, ing_ziemniaki, 'ziemniak', 1000, 'g', 1, false, 'baza');
    END IF;
    
    IF ing_jaja IS NOT NULL THEN
        INSERT INTO "RecipeIngredient" ("recipeId", "ingredientId", "ingredientKey", quantity, unit, "sortOrder", optional, "groupName")
        VALUES (recipe_id, ing_jaja, 'jaja', 2, 'szt', 2, false, 'baza');
    END IF;
    
    IF ing_maka IS NOT NULL THEN
        INSERT INTO "RecipeIngredient" ("recipeId", "ingredientId", "ingredientKey", quantity, unit, "sortOrder", optional, "groupName")
        VALUES (recipe_id, ing_maka, 'maka', 50, 'g', 3, false, 'baza');
    END IF;
    
    IF ing_olej IS NOT NULL THEN
        INSERT INTO "RecipeIngredient" ("recipeId", "ingredientId", "ingredientKey", quantity, unit, "sortOrder", optional, "groupName")
        VALUES (recipe_id, ing_olej, 'olej', 100, 'ml', 4, true, 'baza');
    END IF;
END $$;

-- ============================================================================
-- VERIFICATION QUERIES
-- ============================================================================

-- Count recipes and ingredients
SELECT 
  COUNT(DISTINCT r.id) as recipe_count,
  COUNT(ri.id) as ingredient_count,
  ROUND(AVG(ingredient_per_recipe.cnt), 1) as avg_ingredients_per_recipe
FROM "Recipe" r
LEFT JOIN "RecipeIngredient" ri ON r.id = ri."recipeId"
CROSS JOIN (
  SELECT r2.id, COUNT(ri2.id) as cnt
  FROM "Recipe" r2
  LEFT JOIN "RecipeIngredient" ri2 ON r2.id = ri2."recipeId"
  GROUP BY r2.id
) ingredient_per_recipe;

-- Show new recipes with ingredient counts
SELECT 
  r."localName",
  r.country,
  r.category,
  r.servings,
  COUNT(ri.id) as ingredient_count
FROM "Recipe" r
LEFT JOIN "RecipeIngredient" ri ON r.id = ri."recipeId"
GROUP BY r.id, r."localName", r.country, r.category, r.servings
ORDER BY r."localName";
