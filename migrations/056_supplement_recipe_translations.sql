-- Additional translations for existing recipes

-- Scrambled Eggs (Jajecznica)
UPDATE "Recipe"
SET 
    steps_en = '["1. Beat eggs in a bowl", "2. Melt butter in a pan over medium heat", "3. Pour eggs into pan", "4. Gently stir with spatula until creamy", "5. Season with salt and pepper"]'::jsonb,
    steps_ru = '["1. Взбейте яйца в миске", "2. Растопите сливочное масло на сковороде на среднем огне", "3. Вылейте яйца на сковороду", "4. Аккуратно помешивайте лопаткой до кремообразной консистенции", "5. Приправьте солью и перцем"]'::jsonb
WHERE "canonicalName" = 'Scrambled Eggs';

-- Spaghetti Carbonara
UPDATE "Recipe"
SET 
    steps_en = '["1. Cook spaghetti according to package directions", "2. Fry bacon until crispy", "3. Mix eggs with grated Parmesan", "4. Drain pasta, add to bacon pan", "5. Remove from heat, add egg mixture, toss quickly"]'::jsonb,
    steps_ru = '["1. Отварите спагетти согласно инструкции на упаковке", "2. Обжарьте бекон до хрустящей корочки", "3. Смешайте яйца с тертым пармезаном", "4. Слейте пасту, добавьте в сковороду с беконом", "5. Снимите с огня, добавьте яичную смесь, быстро перемешайте"]'::jsonb
WHERE "canonicalName" = 'Spaghetti Carbonara';

-- Pizza Margherita
UPDATE "Recipe"
SET 
    steps_en = '["1. Roll out pizza dough", "2. Spread tomato sauce", "3. Add mozzarella slices", "4. Drizzle with olive oil", "5. Bake at 250°C for 10-12 minutes", "6. Garnish with fresh basil"]'::jsonb,
    steps_ru = '["1. Раскатайте тесто для пиццы", "2. Намажьте томатным соусом", "3. Выложите ломтики моцареллы", "4. Полейте оливковым маслом", "5. Выпекайте при 250°C 10-12 минут", "6. Украсьте свежим базиликом"]'::jsonb
WHERE "canonicalName" = 'Pizza Margherita';

-- Polish Hunters Stew (Bigos)
UPDATE "Recipe"
SET 
    steps_en = '["1. Chop sauerkraut and fresh cabbage", "2. Sauté onions with bacon and sausage", "3. Add cabbage, tomato paste, and spices", "4. Add meat leftovers if available", "5. Simmer for 2-3 hours, stirring occasionally"]'::jsonb,
    steps_ru = '["1. Нарежьте квашеную и свежую капусту", "2. Обжарьте лук с беконом и колбасой", "3. Добавьте капусту, томатную пасту и специи", "4. Добавьте остатки мяса, если есть", "5. Тушите 2-3 часа, периодически помешивая"]'::jsonb
WHERE "canonicalName" = 'Polish Hunters Stew';

-- Polish Breaded Pork Chop (Kotlet schabowy)
UPDATE "Recipe"
SET 
    steps_en = '["1. Pound pork chops until thin", "2. Season with salt and pepper", "3. Coat in flour, beaten egg, and breadcrumbs", "4. Fry in hot oil until golden brown on both sides", "5. Drain on paper towels"]'::jsonb,
    steps_ru = '["1. Отбейте свиные отбивные до тонкости", "2. Приправьте солью и перцем", "3. Обваляйте в муке, взбитом яйце и панировочных сухарях", "4. Обжарьте в горячем масле до золотистой корочки с обеих сторон", "5. Выложите на бумажные полотенца"]'::jsonb
WHERE "canonicalName" = 'Polish Breaded Pork Chop';

-- Polish Potato Pancakes (Placki ziemniaczane)
UPDATE "Recipe"
SET 
    steps_en = '["1. Grate potatoes and onion", "2. Squeeze out excess liquid", "3. Mix with egg, flour, and salt", "4. Form small pancakes", "5. Fry in hot oil until crispy and golden"]'::jsonb,
    steps_ru = '["1. Натрите картофель и лук на терке", "2. Отожмите лишнюю жидкость", "3. Смешайте с яйцом, мукой и солью", "4. Сформируйте небольшие оладьи", "5. Обжарьте в горячем масле до хрустящей золотистой корочки"]'::jsonb
WHERE "canonicalName" = 'Polish Potato Pancakes';

-- Polish Chicken Soup (Rosół)
UPDATE "Recipe"
SET 
    steps_en = '["1. Boil chicken with vegetables (carrots, parsnip, celery)", "2. Add bay leaf, allspice, and peppercorns", "3. Simmer for 2-3 hours", "4. Strain broth", "5. Serve with noodles and chopped parsley"]'::jsonb,
    steps_ru = '["1. Отварите курицу с овощами (морковь, пастернак, сельдерей)", "2. Добавьте лавровый лист, душистый перец и горошины перца", "3. Варите на медленном огне 2-3 часа", "4. Процедите бульон", "5. Подавайте с лапшой и нарезанной петрушкой"]'::jsonb
WHERE "canonicalName" = 'Polish Chicken Soup';

SELECT COUNT(*) as "recipes_with_en_steps" FROM "Recipe" WHERE steps_en IS NOT NULL;
SELECT COUNT(*) as "recipes_with_ru_steps" FROM "Recipe" WHERE steps_ru IS NOT NULL;
