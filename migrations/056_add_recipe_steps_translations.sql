-- Migration 056: Add steps translations to Recipe table
-- Adds steps_pl, steps_en, steps_ru columns for localized cooking instructions

-- Add translation columns for steps
ALTER TABLE "Recipe" ADD COLUMN IF NOT EXISTS steps_pl JSONB;
ALTER TABLE "Recipe" ADD COLUMN IF NOT EXISTS steps_en JSONB;
ALTER TABLE "Recipe" ADD COLUMN IF NOT EXISTS steps_ru JSONB;

-- Comment on columns
COMMENT ON COLUMN "Recipe".steps_pl IS 'Cooking instructions in Polish (JSONB array of strings)';
COMMENT ON COLUMN "Recipe".steps_en IS 'Cooking instructions in English (JSONB array of strings)';
COMMENT ON COLUMN "Recipe".steps_ru IS 'Cooking instructions in Russian (JSONB array of strings)';

-- Migrate existing steps data to steps_pl (assuming current steps are in Polish)
UPDATE "Recipe" 
SET steps_pl = steps 
WHERE steps IS NOT NULL AND steps_pl IS NULL;

-- Example: Greek Salad (id: 92691aae-c3af-427d-aaed-1408319f0a3c)
UPDATE "Recipe"
SET 
    steps_en = '["1. Cut tomatoes and cucumber into large chunks", "2. Add sliced red onion", "3. Add olives and feta cheese cubes", "4. Drizzle with olive oil and sprinkle oregano", "5. Toss gently before serving"]'::jsonb,
    steps_ru = '["1. Нарежьте помидоры и огурец крупными кусками", "2. Добавьте нарезанный красный лук кольцами", "3. Добавьте оливки и кубики феты", "4. Полейте оливковым маслом и посыпьте орегано", "5. Аккуратно перемешайте перед подачей"]'::jsonb
WHERE id = '92691aae-c3af-427d-aaed-1408319f0a3c';

-- Example: Scrambled Eggs (id: 5c6fdf8f-5df6-4d7f-99df-84ca93a1daaa)
UPDATE "Recipe"
SET 
    steps_en = '["1. Beat eggs in a bowl", "2. Melt butter in a pan over medium heat", "3. Pour eggs into pan", "4. Gently stir with spatula until creamy", "5. Season with salt and pepper"]'::jsonb,
    steps_ru = '["1. Взбейте яйца в миске", "2. Растопите сливочное масло на сковороде на среднем огне", "3. Вылейте яйца на сковороду", "4. Аккуратно помешивайте лопаткой до кремообразной консистенции", "5. Приправьте солью и перцем"]'::jsonb
WHERE id = '5c6fdf8f-5df6-4d7f-99df-84ca93a1daaa';

-- Example: Spaghetti Carbonara (id: 7d890f12-abcd-4321-9876-fedcba098765)
UPDATE "Recipe"
SET 
    steps_en = '["1. Cook spaghetti according to package directions", "2. Fry bacon until crispy", "3. Mix eggs with grated Parmesan", "4. Drain pasta, add to bacon pan", "5. Remove from heat, add egg mixture, toss quickly"]'::jsonb,
    steps_ru = '["1. Отварите спагетти согласно инструкции на упаковке", "2. Обжарьте бекон до хрустящей корочки", "3. Смешайте яйца с тертым пармезаном", "4. Слейте пасту, добавьте в сковороду с беконом", "5. Снимите с огня, добавьте яичную смесь, быстро перемешайте"]'::jsonb
WHERE id = '7d890f12-abcd-4321-9876-fedcba098765';

-- Example: Tomato Soup (id: 9a876543-fedc-ba09-8765-4321dcba9876)
UPDATE "Recipe"
SET 
    steps_en = '["1. Sauté chopped onion in olive oil", "2. Add crushed tomatoes and broth", "3. Season with basil and salt", "4. Simmer for 20 minutes", "5. Blend until smooth, garnish with cream"]'::jsonb,
    steps_ru = '["1. Обжарьте нарезанный лук на оливковом масле", "2. Добавьте измельченные помидоры и бульон", "3. Приправьте базиликом и солью", "4. Тушите 20 минут", "5. Взбейте блендером до однородности, украсьте сливками"]'::jsonb
WHERE id = '9a876543-fedc-ba09-8765-4321dcba9876';

-- Note: Remaining recipes need manual translation
-- Polish steps are already in the 'steps' column
-- English and Russian translations can be added later or generated via AI

COMMENT ON TABLE "Recipe" IS 'Recipe catalog with multilingual support (names, descriptions, steps in PL/EN/RU)';
