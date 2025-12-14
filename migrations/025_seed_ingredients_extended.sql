-- Seed: Extended Polish ingredients catalog (дополнительные 77 продуктов до 200 total)
-- Author: AI Assistant
-- Date: 2025-12-14
-- Description: Расширение каталога популярными польскими продуктами

-- PROTEIN (дополнительно 8 продуктов = 20 total)
INSERT INTO "Ingredient" (id, name, unit, category, "defaultShelfLifeDays", "defaultPricePerUnit", "createdAt")
VALUES
    (gen_random_uuid(), 'Indyk (pierś)', 'g', 'protein', 3, 0.028, NOW()),
    (gen_random_uuid(), 'Kaczka', 'g', 'protein', 3, 0.035, NOW()),
    (gen_random_uuid(), 'Krewetki', 'g', 'protein', 1, 0.080, NOW()),
    (gen_random_uuid(), 'Ośmiornica', 'g', 'protein', 2, 0.090, NOW()),
    (gen_random_uuid(), 'Pstrąg', 'g', 'protein', 2, 0.040, NOW()),
    (gen_random_uuid(), 'Makrela', 'g', 'protein', 2, 0.025, NOW()),
    (gen_random_uuid(), 'Sardynki (puszka)', 'g', 'protein', 730, 0.018, NOW()),
    (gen_random_uuid(), 'Kalmary', 'g', 'protein', 2, 0.045, NOW())
ON CONFLICT DO NOTHING;

-- VEGETABLES (дополнительно 25 продуктов = 50 total)
INSERT INTO "Ingredient" (id, name, unit, category, "defaultShelfLifeDays", "defaultPricePerUnit", "createdAt")
VALUES
    (gen_random_uuid(), 'Awokado', 'g', 'vegetable', 5, 0.035, NOW()),
    (gen_random_uuid(), 'Batат', 'g', 'vegetable', 14, 0.009, NOW()),
    (gen_random_uuid(), 'Botwinka', 'g', 'vegetable', 3, 0.008, NOW()),
    (gen_random_uuid(), 'Buraki', 'g', 'vegetable', 21, 0.004, NOW()),
    (gen_random_uuid(), 'Cebula czerwona', 'g', 'vegetable', 30, 0.005, NOW()),
    (gen_random_uuid(), 'Chrzan', 'g', 'vegetable', 14, 0.015, NOW()),
    (gen_random_uuid(), 'Fasolka szparagowa', 'g', 'vegetable', 5, 0.012, NOW()),
    (gen_random_uuid(), 'Groszek zielony', 'g', 'vegetable', 3, 0.010, NOW()),
    (gen_random_uuid(), 'Jarmuż', 'g', 'vegetable', 5, 0.014, NOW()),
    (gen_random_uuid(), 'Kabaczek', 'g', 'vegetable', 7, 0.007, NOW()),
    (gen_random_uuid(), 'Kapusta kiszona', 'g', 'vegetable', 60, 0.005, NOW()),
    (gen_random_uuid(), 'Kapusta pekińska', 'g', 'vegetable', 7, 0.006, NOW()),
    (gen_random_uuid(), 'Kukurydza', 'g', 'vegetable', 3, 0.008, NOW()),
    (gen_random_uuid(), 'Mango', 'g', 'vegetable', 5, 0.018, NOW()),
    (gen_random_uuid(), 'Melon', 'g', 'vegetable', 7, 0.006, NOW()),
    (gen_random_uuid(), 'Pietruszka (korzeń)', 'g', 'vegetable', 14, 0.005, NOW()),
    (gen_random_uuid(), 'Pieczarki brązowe', 'g', 'vegetable', 5, 0.018, NOW()),
    (gen_random_uuid(), 'Pomidory koktajlowe', 'g', 'vegetable', 5, 0.012, NOW()),
    (gen_random_uuid(), 'Rzepa', 'g', 'vegetable', 21, 0.004, NOW()),
    (gen_random_uuid(), 'Rzodkiewka', 'g', 'vegetable', 7, 0.006, NOW()),
    (gen_random_uuid(), 'Seler (korzeń)', 'g', 'vegetable', 14, 0.005, NOW()),
    (gen_random_uuid(), 'Szparagi', 'g', 'vegetable', 3, 0.025, NOW()),
    (gen_random_uuid(), 'Winogrona', 'g', 'vegetable', 7, 0.015, NOW()),
    (gen_random_uuid(), 'Arbuz', 'g', 'vegetable', 7, 0.003, NOW()),
    (gen_random_uuid(), 'Gruszka', 'g', 'vegetable', 14, 0.006, NOW())
ON CONFLICT DO NOTHING;

-- DAIRY (дополнительно 12 продуктов = 25 total)
INSERT INTO "Ingredient" (id, name, unit, category, "defaultShelfLifeDays", "defaultPricePerUnit", "createdAt")
VALUES
    (gen_random_uuid(), 'Mascarpone', 'g', 'dairy', 14, 0.035, NOW()),
    (gen_random_uuid(), 'Ricotta', 'g', 'dairy', 7, 0.020, NOW()),
    (gen_random_uuid(), 'Ser pleśniowy', 'g', 'dairy', 21, 0.040, NOW()),
    (gen_random_uuid(), 'Ser camembert', 'g', 'dairy', 21, 0.035, NOW()),
    (gen_random_uuid(), 'Ser kozi', 'g', 'dairy', 21, 0.045, NOW()),
    (gen_random_uuid(), 'Ser topiony', 'g', 'dairy', 60, 0.015, NOW()),
    (gen_random_uuid(), 'Mleko kokosowe', 'ml', 'dairy', 730, 0.008, NOW()),
    (gen_random_uuid(), 'Mleko migdałowe', 'ml', 'dairy', 7, 0.010, NOW()),
    (gen_random_uuid(), 'Jogurt grecki', 'g', 'dairy', 21, 0.008, NOW()),
    (gen_random_uuid(), 'Maślanka', 'ml', 'dairy', 7, 0.003, NOW()),
    (gen_random_uuid(), 'Serek wiejski', 'g', 'dairy', 7, 0.010, NOW()),
    (gen_random_uuid(), 'Śmietana kremówka', 'ml', 'dairy', 14, 0.015, NOW())
ON CONFLICT DO NOTHING;

-- GRAIN (дополнительно 12 продуктов = 25 total)
INSERT INTO "Ingredient" (id, name, unit, category, "defaultShelfLifeDays", "defaultPricePerUnit", "createdAt")
VALUES
    (gen_random_uuid(), 'Quinoa', 'g', 'grain', 365, 0.015, NOW()),
    (gen_random_uuid(), 'Bulgur', 'g', 'grain', 365, 0.006, NOW()),
    (gen_random_uuid(), 'Kuskus', 'g', 'grain', 365, 0.007, NOW()),
    (gen_random_uuid(), 'Ryż basmati', 'g', 'grain', 730, 0.008, NOW()),
    (gen_random_uuid(), 'Ryż jaśminowy', 'g', 'grain', 730, 0.009, NOW()),
    (gen_random_uuid(), 'Kasza orkiszowa', 'g', 'grain', 365, 0.008, NOW()),
    (gen_random_uuid(), 'Kasza manna', 'g', 'grain', 365, 0.004, NOW()),
    (gen_random_uuid(), 'Makaron pełnoziarnisty', 'g', 'grain', 365, 0.007, NOW()),
    (gen_random_uuid(), 'Makaron ryżowy', 'g', 'grain', 365, 0.010, NOW()),
    (gen_random_uuid(), 'Płatki kukurydziane', 'g', 'grain', 180, 0.006, NOW()),
    (gen_random_uuid(), 'Mąka kukurydziana', 'g', 'grain', 180, 0.004, NOW()),
    (gen_random_uuid(), 'Otręby', 'g', 'grain', 180, 0.005, NOW())
ON CONFLICT DO NOTHING;

-- CONDIMENTS (дополнительно 20 продуктów = 44 total)
INSERT INTO "Ingredient" (id, name, unit, category, "defaultShelfLifeDays", "defaultPricePerUnit", "createdAt")
VALUES
    (gen_random_uuid(), 'Wasabi', 'g', 'condiment', 730, 0.150, NOW()),
    (gen_random_uuid(), 'Sos teriyaki', 'ml', 'condiment', 365, 0.018, NOW()),
    (gen_random_uuid(), 'Sos rybny', 'ml', 'condiment', 730, 0.020, NOW()),
    (gen_random_uuid(), 'Tahini', 'g', 'condiment', 365, 0.025, NOW()),
    (gen_random_uuid(), 'Masło orzechowe', 'g', 'condiment', 180, 0.018, NOW()),
    (gen_random_uuid(), 'Pasta pomidorowa', 'g', 'condiment', 730, 0.008, NOW()),
    (gen_random_uuid(), 'Koncentrat pomidorowy', 'g', 'condiment', 730, 0.012, NOW()),
    (gen_random_uuid(), 'Kapary', 'g', 'condiment', 730, 0.040, NOW()),
    (gen_random_uuid(), 'Oliwki zielone', 'g', 'condiment', 180, 0.015, NOW()),
    (gen_random_uuid(), 'Oliwki czarne', 'g', 'condiment', 180, 0.015, NOW()),
    (gen_random_uuid(), 'Chili (świeże)', 'g', 'condiment', 7, 0.030, NOW()),
    (gen_random_uuid(), 'Chili (płatki)', 'g', 'condiment', 365, 0.050, NOW()),
    (gen_random_uuid(), 'Pieprz cayenne', 'g', 'condiment', 365, 0.045, NOW()),
    (gen_random_uuid(), 'Koper', 'g', 'condiment', 365, 0.040, NOW()),
    (gen_random_uuid(), 'Liść laurowy', 'g', 'condiment', 365, 0.035, NOW()),
    (gen_random_uuid(), 'Ziele angielskie', 'g', 'condiment', 730, 0.040, NOW()),
    (gen_random_uuid(), 'Goździki', 'g', 'condiment', 730, 0.070, NOW()),
    (gen_random_uuid(), 'Kardamon', 'g', 'condiment', 365, 0.120, NOW()),
    (gen_random_uuid(), 'Kolendra (nasiona)', 'g', 'condiment', 365, 0.035, NOW()),
    (gen_random_uuid(), 'Szafran', 'g', 'condiment', 730, 2.500, NOW())
ON CONFLICT DO NOTHING;

-- OTHER (дополнительно 10 продуктов = 46 total)
INSERT INTO "Ingredient" (id, name, unit, category, "defaultShelfLifeDays", "defaultPricePerUnit", "createdAt")
VALUES
    (gen_random_uuid(), 'Siemię lniane', 'g', 'other', 180, 0.008, NOW()),
    (gen_random_uuid(), 'Nasiona chia', 'g', 'other', 365, 0.020, NOW()),
    (gen_random_uuid(), 'Orzechy włoskie', 'g', 'other', 180, 0.040, NOW()),
    (gen_random_uuid(), 'Orzechy laskowe', 'g', 'other', 180, 0.038, NOW()),
    (gen_random_uuid(), 'Migdały', 'g', 'other', 180, 0.035, NOW()),
    (gen_random_uuid(), 'Rodzynki', 'g', 'other', 365, 0.015, NOW()),
    (gen_random_uuid(), 'Daktyle', 'g', 'other', 365, 0.025, NOW()),
    (gen_random_uuid(), 'Mleko skondensowane', 'g', 'other', 365, 0.010, NOW()),
    (gen_random_uuid(), 'Żelatyna', 'g', 'other', 730, 0.030, NOW()),
    (gen_random_uuid(), 'Ekstrakt waniliowy', 'ml', 'other', 730, 0.200, NOW())
ON CONFLICT DO NOTHING;
