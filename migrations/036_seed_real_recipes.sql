-- Migration 036: Seed Real Recipes
-- Created: 2024-12-21
-- Purpose: Add authentic recipes from Polish, Italian, French cuisines

-- ==============================================
-- HELPER: Get ingredient ID by name
-- ==============================================
-- Note: These queries will be used in the INSERT statements below

-- ==============================================
-- 1. POLISH RECIPES
-- ==============================================

-- Recipe: Pierogi Ruskie (Polish Dumplings with Potato & Cheese)
DO $$
DECLARE
    recipe_id UUID;
    potato_id UUID;
    cheese_id UUID;
    onion_id UUID;
    flour_id UUID;
    egg_id UUID;
    salt_id UUID;
    butter_id UUID;
BEGIN
    -- Insert recipe
    INSERT INTO "Recipe" (
        "canonicalName", "localName", country, region, category, difficulty,
        "timeMinutes", servings, steps, "nutritionProfile", source
    ) VALUES (
        'Pierogi Ruskie',
        'Pierogi ruskie',
        'Poland',
        'Małopolska',
        'main',
        'medium',
        90,
        4,
        '[
            {"step": 1, "instruction": "Ugotuj ziemniaki do miękkości, osusz i rozgnieć na puree"},
            {"step": 2, "instruction": "Podsmaż cebulę na maśle do złocistości"},
            {"step": 3, "instruction": "Wymieszaj puree z serem i cebulą, dopraw solą"},
            {"step": 4, "instruction": "Przygotuj ciasto: wymieszaj mąkę, jajko, wodę i sól"},
            {"step": 5, "instruction": "Rozwałkuj ciasto, wykrój kółka szklanką"},
            {"step": 6, "instruction": "Nałóż farsz, zaklej brzegi, gotuj w osolonej wodzie 3-4 min"},
            {"step": 7, "instruction": "Podawaj z podsmażoną cebulą i śmietaną"}
        ]'::jsonb,
        '{"type": "balanced", "calories": 420}'::jsonb,
        '{"type": "traditional", "reference": "Polish national dish"}'::jsonb
    ) RETURNING id INTO recipe_id;

    -- Get ingredient IDs
    SELECT id INTO potato_id FROM "Ingredient" WHERE name ILIKE 'ziemniak%' LIMIT 1;
    SELECT id INTO cheese_id FROM "Ingredient" WHERE name ILIKE '%ser%twaróg%' OR name ILIKE 'twaróg%' LIMIT 1;
    SELECT id INTO onion_id FROM "Ingredient" WHERE name ILIKE 'cebula%' LIMIT 1;
    SELECT id INTO flour_id FROM "Ingredient" WHERE name ILIKE 'mąka%' LIMIT 1;
    SELECT id INTO egg_id FROM "Ingredient" WHERE name ILIKE 'jaj%' LIMIT 1;
    SELECT id INTO salt_id FROM "Ingredient" WHERE name ILIKE 'sól%' LIMIT 1;
    SELECT id INTO butter_id FROM "Ingredient" WHERE name ILIKE 'masło%' LIMIT 1;

    -- Insert ingredients
    IF potato_id IS NOT NULL THEN
        INSERT INTO "RecipeIngredient" ("recipeId", "ingredientId", "ingredientKey", quantity, unit, "sortOrder")
        VALUES (recipe_id, potato_id, 'potato', 500, 'g', 1);
    END IF;
    
    IF cheese_id IS NOT NULL THEN
        INSERT INTO "RecipeIngredient" ("recipeId", "ingredientId", "ingredientKey", quantity, unit, "sortOrder")
        VALUES (recipe_id, cheese_id, 'cottage_cheese', 250, 'g', 2);
    END IF;
    
    IF onion_id IS NOT NULL THEN
        INSERT INTO "RecipeIngredient" ("recipeId", "ingredientId", "ingredientKey", quantity, unit, "sortOrder")
        VALUES (recipe_id, onion_id, 'onion', 200, 'g', 3);
    END IF;
    
    IF flour_id IS NOT NULL THEN
        INSERT INTO "RecipeIngredient" ("recipeId", "ingredientId", "ingredientKey", quantity, unit, "sortOrder")
        VALUES (recipe_id, flour_id, 'flour', 400, 'g', 4);
    END IF;
    
    IF egg_id IS NOT NULL THEN
        INSERT INTO "RecipeIngredient" ("recipeId", "ingredientId", "ingredientKey", quantity, unit, "sortOrder")
        VALUES (recipe_id, egg_id, 'egg', 1, 'pcs', 5);
    END IF;
    
    IF butter_id IS NOT NULL THEN
        INSERT INTO "RecipeIngredient" ("recipeId", "ingredientId", "ingredientKey", quantity, unit, "sortOrder")
        VALUES (recipe_id, butter_id, 'butter', 50, 'g', 6);
    END IF;

    -- Add allergens
    INSERT INTO "RecipeAllergen" ("recipeId", "allergenId")
    SELECT recipe_id, id FROM "Allergen" WHERE name IN ('gluten', 'lactose', 'eggs');

    -- Add diet tags
    INSERT INTO "RecipeDietTag" ("recipeId", "dietTagId")
    SELECT recipe_id, id FROM "DietTag" WHERE name = 'vegetarian';

END $$;

-- ==============================================
-- Recipe: Bigos (Polish Hunter's Stew)
-- ==============================================
DO $$
DECLARE
    recipe_id UUID;
    cabbage_id UUID;
    sauerkraut_id UUID;
    pork_id UUID;
    sausage_id UUID;
    onion_id UUID;
    tomato_id UUID;
BEGIN
    INSERT INTO "Recipe" (
        "canonicalName", "localName", country, region, category, difficulty,
        "timeMinutes", servings, steps, "nutritionProfile", source
    ) VALUES (
        'Bigos',
        'Bigos myśliwski',
        'Poland',
        NULL,
        'main',
        'medium',
        180,
        6,
        '[
            {"step": 1, "instruction": "Pokrój kapustę i kiszoną kapustę w paski"},
            {"step": 2, "instruction": "Podsmaż cebulę, dodaj mięso pokrojone w kostkę"},
            {"step": 3, "instruction": "Dodaj kapustę, kiszoną kapustę, koncentrat pomidorowy"},
            {"step": 4, "instruction": "Duś na małym ogniu 2-3 godziny, dodając wodę w razie potrzeby"},
            {"step": 5, "instruction": "Dodaj pokrojoną kiełbasę 30 min przed końcem"},
            {"step": 6, "instruction": "Dopraw solą, pieprzem, listkiem laurowym"}
        ]'::jsonb,
        '{"type": "high-protein", "calories": 380}'::jsonb,
        '{"type": "traditional", "reference": "Traditional Polish stew"}'::jsonb
    ) RETURNING id INTO recipe_id;

    SELECT id INTO cabbage_id FROM "Ingredient" WHERE name ILIKE 'kapusta%' AND name NOT ILIKE '%kiszon%' LIMIT 1;
    SELECT id INTO sauerkraut_id FROM "Ingredient" WHERE name ILIKE '%kiszon%kapust%' LIMIT 1;
    SELECT id INTO pork_id FROM "Ingredient" WHERE name ILIKE '%wieprzow%' OR name ILIKE 'schab%' LIMIT 1;
    SELECT id INTO onion_id FROM "Ingredient" WHERE name ILIKE 'cebula%' LIMIT 1;

    IF cabbage_id IS NOT NULL THEN
        INSERT INTO "RecipeIngredient" ("recipeId", "ingredientId", "ingredientKey", quantity, unit, "sortOrder")
        VALUES (recipe_id, cabbage_id, 'cabbage', 500, 'g', 1);
    END IF;
    
    IF sauerkraut_id IS NOT NULL THEN
        INSERT INTO "RecipeIngredient" ("recipeId", "ingredientId", "ingredientKey", quantity, unit, "sortOrder")
        VALUES (recipe_id, sauerkraut_id, 'sauerkraut', 500, 'g', 2);
    END IF;
    
    IF pork_id IS NOT NULL THEN
        INSERT INTO "RecipeIngredient" ("recipeId", "ingredientId", "ingredientKey", quantity, unit, "sortOrder")
        VALUES (recipe_id, pork_id, 'pork', 400, 'g', 3);
    END IF;
    
    IF onion_id IS NOT NULL THEN
        INSERT INTO "RecipeIngredient" ("recipeId", "ingredientId", "ingredientKey", quantity, unit, "sortOrder")
        VALUES (recipe_id, onion_id, 'onion', 150, 'g', 4);
    END IF;

    INSERT INTO "RecipeDietTag" ("recipeId", "dietTagId")
    SELECT recipe_id, id FROM "DietTag" WHERE name = 'high-protein';

END $$;

-- ==============================================
-- 2. ITALIAN RECIPES
-- ==============================================

-- Recipe: Spaghetti Carbonara
DO $$
DECLARE
    recipe_id UUID;
    pasta_id UUID;
    egg_id UUID;
    cheese_id UUID;
    bacon_id UUID;
BEGIN
    INSERT INTO "Recipe" (
        "canonicalName", "localName", country, region, category, difficulty,
        "timeMinutes", servings, steps, "nutritionProfile", source
    ) VALUES (
        'Spaghetti Carbonara',
        'Spaghetti alla Carbonara',
        'Italy',
        'Lazio',
        'main',
        'easy',
        25,
        4,
        '[
            {"step": 1, "instruction": "Gotuj spaghetti al dente w osolonej wodzie"},
            {"step": 2, "instruction": "Podsmaż boczek na patelni do chrupkości"},
            {"step": 3, "instruction": "Wymieszaj żółtka z startym serem pecorino"},
            {"step": 4, "instruction": "Odcedź makaron, zachowaj szklankę wody"},
            {"step": 5, "instruction": "Wymieszaj gorący makaron z żółtkami poza ogniem"},
            {"step": 6, "instruction": "Dodaj boczek i wodę z makaronu dla kremowości"},
            {"step": 7, "instruction": "Podawaj natychmiast z pieprzem"}
        ]'::jsonb,
        '{"type": "high-protein", "calories": 550}'::jsonb,
        '{"type": "traditional", "reference": "Roman classic recipe"}'::jsonb
    ) RETURNING id INTO recipe_id;

    SELECT id INTO pasta_id FROM "Ingredient" WHERE name ILIKE 'makaron%' OR name ILIKE 'spaghetti%' LIMIT 1;
    SELECT id INTO egg_id FROM "Ingredient" WHERE name ILIKE 'jaj%' LIMIT 1;
    SELECT id INTO cheese_id FROM "Ingredient" WHERE name ILIKE '%parmezan%' OR name ILIKE '%pecorino%' LIMIT 1;
    SELECT id INTO bacon_id FROM "Ingredient" WHERE name ILIKE 'boczek%' OR name ILIKE 'bekon%' LIMIT 1;

    IF pasta_id IS NOT NULL THEN
        INSERT INTO "RecipeIngredient" ("recipeId", "ingredientId", "ingredientKey", quantity, unit, "sortOrder")
        VALUES (recipe_id, pasta_id, 'spaghetti', 400, 'g', 1);
    END IF;
    
    IF egg_id IS NOT NULL THEN
        INSERT INTO "RecipeIngredient" ("recipeId", "ingredientId", "ingredientKey", quantity, unit, "sortOrder")
        VALUES (recipe_id, egg_id, 'egg', 4, 'pcs', 2);
    END IF;
    
    IF cheese_id IS NOT NULL THEN
        INSERT INTO "RecipeIngredient" ("recipeId", "ingredientId", "ingredientKey", quantity, unit, "sortOrder")
        VALUES (recipe_id, cheese_id, 'pecorino', 100, 'g', 3);
    END IF;
    
    IF bacon_id IS NOT NULL THEN
        INSERT INTO "RecipeIngredient" ("recipeId", "ingredientId", "ingredientKey", quantity, unit, "sortOrder")
        VALUES (recipe_id, bacon_id, 'guanciale', 150, 'g', 4);
    END IF;

    INSERT INTO "RecipeAllergen" ("recipeId", "allergenId")
    SELECT recipe_id, id FROM "Allergen" WHERE name IN ('gluten', 'eggs', 'lactose');

    INSERT INTO "RecipeDietTag" ("recipeId", "dietTagId")
    SELECT recipe_id, id FROM "DietTag" WHERE name = 'high-protein';

END $$;

-- ==============================================
-- Recipe: Margherita Pizza
-- ==============================================
DO $$
DECLARE
    recipe_id UUID;
    flour_id UUID;
    tomato_id UUID;
    mozzarella_id UUID;
    basil_id UUID;
    oil_id UUID;
BEGIN
    INSERT INTO "Recipe" (
        "canonicalName", "localName", country, region, category, difficulty,
        "timeMinutes", servings, steps, "nutritionProfile", source
    ) VALUES (
        'Pizza Margherita',
        'Pizza Margherita',
        'Italy',
        'Campania',
        'main',
        'medium',
        120,
        2,
        '[
            {"step": 1, "instruction": "Wymieszaj mąkę, drożdże, wodę i sól, wyrób ciasto"},
            {"step": 2, "instruction": "Odstaw ciasto do wyrośnięcia na 1 godzinę"},
            {"step": 3, "instruction": "Rozwałkuj ciasto na cienki placek"},
            {"step": 4, "instruction": "Rozprowadź sos pomidorowy, dodaj mozzarellę"},
            {"step": 5, "instruction": "Piecz w 250°C przez 10-12 minut"},
            {"step": 6, "instruction": "Dodaj świeżą bazylię przed podaniem"}
        ]'::jsonb,
        '{"type": "balanced", "calories": 480}'::jsonb,
        '{"type": "traditional", "reference": "Neapolitan pizza classic"}'::jsonb
    ) RETURNING id INTO recipe_id;

    SELECT id INTO flour_id FROM "Ingredient" WHERE name ILIKE 'mąka%' LIMIT 1;
    SELECT id INTO tomato_id FROM "Ingredient" WHERE name ILIKE '%pomidor%' LIMIT 1;
    SELECT id INTO mozzarella_id FROM "Ingredient" WHERE name ILIKE 'mozzarella%' LIMIT 1;
    SELECT id INTO oil_id FROM "Ingredient" WHERE name ILIKE 'oliwa%' LIMIT 1;

    IF flour_id IS NOT NULL THEN
        INSERT INTO "RecipeIngredient" ("recipeId", "ingredientId", "ingredientKey", quantity, unit, "sortOrder")
        VALUES (recipe_id, flour_id, 'flour', 300, 'g', 1);
    END IF;
    
    IF tomato_id IS NOT NULL THEN
        INSERT INTO "RecipeIngredient" ("recipeId", "ingredientId", "ingredientKey", quantity, unit, "sortOrder")
        VALUES (recipe_id, tomato_id, 'tomato_sauce', 150, 'ml', 2);
    END IF;
    
    IF mozzarella_id IS NOT NULL THEN
        INSERT INTO "RecipeIngredient" ("recipeId", "ingredientId", "ingredientKey", quantity, unit, "sortOrder")
        VALUES (recipe_id, mozzarella_id, 'mozzarella', 200, 'g', 3);
    END IF;
    
    IF oil_id IS NOT NULL THEN
        INSERT INTO "RecipeIngredient" ("recipeId", "ingredientId", "ingredientKey", quantity, unit, optional, "sortOrder")
        VALUES (recipe_id, oil_id, 'olive_oil', 20, 'ml', TRUE, 4);
    END IF;

    INSERT INTO "RecipeAllergen" ("recipeId", "allergenId")
    SELECT recipe_id, id FROM "Allergen" WHERE name IN ('gluten', 'lactose');

    INSERT INTO "RecipeDietTag" ("recipeId", "dietTagId")
    SELECT recipe_id, id FROM "DietTag" WHERE name = 'vegetarian';

END $$;

-- ==============================================
-- 3. SIMPLE RECIPES
-- ==============================================

-- Recipe: Scrambled Eggs (Jajecznica)
DO $$
DECLARE
    recipe_id UUID;
    egg_id UUID;
    butter_id UUID;
    milk_id UUID;
BEGIN
    INSERT INTO "Recipe" (
        "canonicalName", "localName", country, region, category, difficulty,
        "timeMinutes", servings, steps, "nutritionProfile", source
    ) VALUES (
        'Scrambled Eggs',
        'Jajecznica',
        'Poland',
        NULL,
        'main',
        'easy',
        10,
        2,
        '[
            {"step": 1, "instruction": "Rozbij jajka do miski, dodaj mleko, wymieszaj"},
            {"step": 2, "instruction": "Rozgrzej masło na patelni"},
            {"step": 3, "instruction": "Wlej jajka, smaż mieszając do uzyskania kremowej konsystencji"},
            {"step": 4, "instruction": "Dopraw solą i pieprzem"}
        ]'::jsonb,
        '{"type": "high-protein", "calories": 280}'::jsonb,
        '{"type": "traditional", "reference": "Classic breakfast"}'::jsonb
    ) RETURNING id INTO recipe_id;

    SELECT id INTO egg_id FROM "Ingredient" WHERE name ILIKE 'jaj%' LIMIT 1;
    SELECT id INTO butter_id FROM "Ingredient" WHERE name ILIKE 'masło%' LIMIT 1;
    SELECT id INTO milk_id FROM "Ingredient" WHERE name ILIKE 'mleko%' LIMIT 1;

    IF egg_id IS NOT NULL THEN
        INSERT INTO "RecipeIngredient" ("recipeId", "ingredientId", "ingredientKey", quantity, unit, "sortOrder")
        VALUES (recipe_id, egg_id, 'egg', 4, 'pcs', 1);
    END IF;
    
    IF butter_id IS NOT NULL THEN
        INSERT INTO "RecipeIngredient" ("recipeId", "ingredientId", "ingredientKey", quantity, unit, "sortOrder")
        VALUES (recipe_id, butter_id, 'butter', 20, 'g', 2);
    END IF;
    
    IF milk_id IS NOT NULL THEN
        INSERT INTO "RecipeIngredient" ("recipeId", "ingredientId", "ingredientKey", quantity, unit, optional, "sortOrder")
        VALUES (recipe_id, milk_id, 'milk', 50, 'ml', TRUE, 3);
    END IF;

    INSERT INTO "RecipeAllergen" ("recipeId", "allergenId")
    SELECT recipe_id, id FROM "Allergen" WHERE name IN ('eggs', 'lactose');

    INSERT INTO "RecipeDietTag" ("recipeId", "dietTagId")
    SELECT recipe_id, id FROM "DietTag" WHERE name IN ('vegetarian', 'gluten-free');

END $$;

-- ==============================================
-- Recipe: Greek Salad
-- ==============================================
DO $$
DECLARE
    recipe_id UUID;
    tomato_id UUID;
    cucumber_id UUID;
    feta_id UUID;
    olive_id UUID;
    onion_id UUID;
    oil_id UUID;
BEGIN
    INSERT INTO "Recipe" (
        "canonicalName", "localName", country, region, category, difficulty,
        "timeMinutes", servings, steps, "nutritionProfile", source
    ) VALUES (
        'Greek Salad',
        'Sałatka grecka',
        'Greece',
        NULL,
        'salad',
        'easy',
        15,
        4,
        '[
            {"step": 1, "instruction": "Pokrój pomidory i ogórka w duże kawałki"},
            {"step": 2, "instruction": "Dodaj pokrojoną czerwoną cebulę w piórka"},
            {"step": 3, "instruction": "Dodaj oliwki i ser feta w kostkach"},
            {"step": 4, "instruction": "Polej oliwą z oliwek, posyp oregano"},
            {"step": 5, "instruction": "Wymieszaj delikatnie przed podaniem"}
        ]'::jsonb,
        '{"type": "balanced", "calories": 220}'::jsonb,
        '{"type": "traditional", "reference": "Mediterranean classic"}'::jsonb
    ) RETURNING id INTO recipe_id;

    SELECT id INTO tomato_id FROM "Ingredient" WHERE name ILIKE 'pomidor%' LIMIT 1;
    SELECT id INTO cucumber_id FROM "Ingredient" WHERE name ILIKE 'ogórek%' LIMIT 1;
    SELECT id INTO feta_id FROM "Ingredient" WHERE name ILIKE 'feta%' LIMIT 1;
    SELECT id INTO onion_id FROM "Ingredient" WHERE name ILIKE 'cebula%' LIMIT 1;
    SELECT id INTO oil_id FROM "Ingredient" WHERE name ILIKE 'oliwa%' LIMIT 1;

    IF tomato_id IS NOT NULL THEN
        INSERT INTO "RecipeIngredient" ("recipeId", "ingredientId", "ingredientKey", quantity, unit, "sortOrder")
        VALUES (recipe_id, tomato_id, 'tomato', 400, 'g', 1);
    END IF;
    
    IF cucumber_id IS NOT NULL THEN
        INSERT INTO "RecipeIngredient" ("recipeId", "ingredientId", "ingredientKey", quantity, unit, "sortOrder")
        VALUES (recipe_id, cucumber_id, 'cucumber', 200, 'g', 2);
    END IF;
    
    IF feta_id IS NOT NULL THEN
        INSERT INTO "RecipeIngredient" ("recipeId", "ingredientId", "ingredientKey", quantity, unit, "sortOrder")
        VALUES (recipe_id, feta_id, 'feta', 150, 'g', 3);
    END IF;
    
    IF onion_id IS NOT NULL THEN
        INSERT INTO "RecipeIngredient" ("recipeId", "ingredientId", "ingredientKey", quantity, unit, "sortOrder")
        VALUES (recipe_id, onion_id, 'red_onion', 100, 'g', 4);
    END IF;
    
    IF oil_id IS NOT NULL THEN
        INSERT INTO "RecipeIngredient" ("recipeId", "ingredientId", "ingredientKey", quantity, unit, "sortOrder")
        VALUES (recipe_id, oil_id, 'olive_oil', 30, 'ml', 5);
    END IF;

    INSERT INTO "RecipeAllergen" ("recipeId", "allergenId")
    SELECT recipe_id, id FROM "Allergen" WHERE name = 'lactose';

    INSERT INTO "RecipeDietTag" ("recipeId", "dietTagId")
    SELECT recipe_id, id FROM "DietTag" WHERE name IN ('vegetarian', 'gluten-free', 'low-carb');

END $$;
