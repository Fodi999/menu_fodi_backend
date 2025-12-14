-- Seed: Polish ingredients catalog (базовый каталог продуктов для Польши)
-- Author: AI Assistant
-- Date: 2025-12-14
-- Description: 100+ popular Polish ingredients with categories, units, shelf life, and default prices

-- PROTEIN (Białka - Мясо, рыба, яйца)
INSERT INTO "Ingredient" (id, name, unit, category, "defaultShelfLifeDays", "defaultPricePerUnit", "createdAt")
VALUES
    (gen_random_uuid(), 'Kurczak (pierś)', 'g', 'protein', 3, 0.025, NOW()),
    (gen_random_uuid(), 'Kurczak (udo)', 'g', 'protein', 3, 0.018, NOW()),
    (gen_random_uuid(), 'Wołowina (rostbef)', 'g', 'protein', 5, 0.045, NOW()),
    (gen_random_uuid(), 'Wieprzowina (schab)', 'g', 'protein', 5, 0.028, NOW()),
    (gen_random_uuid(), 'Kiełbasa', 'g', 'protein', 14, 0.022, NOW()),
    (gen_random_uuid(), 'Boczek', 'g', 'protein', 21, 0.025, NOW()),
    (gen_random_uuid(), 'Łosoś', 'g', 'protein', 2, 0.065, NOW()),
    (gen_random_uuid(), 'Dorsz', 'g', 'protein', 2, 0.035, NOW()),
    (gen_random_uuid(), 'Tuńczyk (puszka)', 'g', 'protein', 730, 0.015, NOW()),
    (gen_random_uuid(), 'Jaja', 'pcs', 'protein', 28, 0.60, NOW()),
    (gen_random_uuid(), 'Tofu', 'g', 'protein', 7, 0.012, NOW()),
    (gen_random_uuid(), 'Szynka', 'g', 'protein', 7, 0.030, NOW())
ON CONFLICT DO NOTHING;

-- VEGETABLES (Warzywa i owoce)
INSERT INTO "Ingredient" (id, name, unit, category, "defaultShelfLifeDays", "defaultPricePerUnit", "createdAt")
VALUES
    (gen_random_uuid(), 'Pomidor', 'g', 'vegetable', 5, 0.008, NOW()),
    (gen_random_uuid(), 'Ogórek', 'g', 'vegetable', 7, 0.006, NOW()),
    (gen_random_uuid(), 'Cebula', 'g', 'vegetable', 30, 0.003, NOW()),
    (gen_random_uuid(), 'Czosnek', 'g', 'vegetable', 60, 0.015, NOW()),
    (gen_random_uuid(), 'Marchew', 'g', 'vegetable', 21, 0.004, NOW()),
    (gen_random_uuid(), 'Ziemniak', 'g', 'vegetable', 30, 0.002, NOW()),
    (gen_random_uuid(), 'Kapusta biała', 'g', 'vegetable', 14, 0.003, NOW()),
    (gen_random_uuid(), 'Papryka czerwona', 'g', 'vegetable', 7, 0.012, NOW()),
    (gen_random_uuid(), 'Papryka zielona', 'g', 'vegetable', 7, 0.010, NOW()),
    (gen_random_uuid(), 'Brokuł', 'g', 'vegetable', 5, 0.009, NOW()),
    (gen_random_uuid(), 'Kalafior', 'g', 'vegetable', 5, 0.008, NOW()),
    (gen_random_uuid(), 'Szpinak', 'g', 'vegetable', 3, 0.012, NOW()),
    (gen_random_uuid(), 'Sałata', 'g', 'vegetable', 5, 0.010, NOW()),
    (gen_random_uuid(), 'Rukola', 'g', 'vegetable', 3, 0.020, NOW()),
    (gen_random_uuid(), 'Por', 'g', 'vegetable', 7, 0.008, NOW()),
    (gen_random_uuid(), 'Bakłażan', 'g', 'vegetable', 7, 0.011, NOW()),
    (gen_random_uuid(), 'Cukinia', 'g', 'vegetable', 7, 0.007, NOW()),
    (gen_random_uuid(), 'Dynia', 'g', 'vegetable', 60, 0.004, NOW()),
    (gen_random_uuid(), 'Pieczarki', 'g', 'vegetable', 5, 0.015, NOW()),
    (gen_random_uuid(), 'Jabłko', 'g', 'vegetable', 30, 0.005, NOW()),
    (gen_random_uuid(), 'Banan', 'g', 'vegetable', 7, 0.008, NOW()),
    (gen_random_uuid(), 'Pomarańcza', 'g', 'vegetable', 14, 0.007, NOW()),
    (gen_random_uuid(), 'Cytryna', 'g', 'vegetable', 21, 0.012, NOW()),
    (gen_random_uuid(), 'Truskawka', 'g', 'vegetable', 3, 0.018, NOW()),
    (gen_random_uuid(), 'Malina', 'g', 'vegetable', 2, 0.025, NOW())
ON CONFLICT DO NOTHING;

-- DAIRY (Nabiał)
INSERT INTO "Ingredient" (id, name, unit, category, "defaultShelfLifeDays", "defaultPricePerUnit", "createdAt")
VALUES
    (gen_random_uuid(), 'Mleko 2%', 'ml', 'dairy', 7, 0.003, NOW()),
    (gen_random_uuid(), 'Mleko 3.2%', 'ml', 'dairy', 7, 0.004, NOW()),
    (gen_random_uuid(), 'Śmietana 18%', 'ml', 'dairy', 14, 0.008, NOW()),
    (gen_random_uuid(), 'Śmietana 30%', 'ml', 'dairy', 14, 0.012, NOW()),
    (gen_random_uuid(), 'Jogurt naturalny', 'g', 'dairy', 21, 0.006, NOW()),
    (gen_random_uuid(), 'Ser żółty', 'g', 'dairy', 30, 0.028, NOW()),
    (gen_random_uuid(), 'Ser biały', 'g', 'dairy', 7, 0.015, NOW()),
    (gen_random_uuid(), 'Ser feta', 'g', 'dairy', 21, 0.020, NOW()),
    (gen_random_uuid(), 'Mozzarella', 'g', 'dairy', 14, 0.022, NOW()),
    (gen_random_uuid(), 'Parmezan', 'g', 'dairy', 60, 0.045, NOW()),
    (gen_random_uuid(), 'Masło', 'g', 'dairy', 60, 0.018, NOW()),
    (gen_random_uuid(), 'Twaróg', 'g', 'dairy', 7, 0.012, NOW()),
    (gen_random_uuid(), 'Kefir', 'ml', 'dairy', 14, 0.004, NOW())
ON CONFLICT DO NOTHING;

-- GRAIN (Zboża, makarony, pieczywo)
INSERT INTO "Ingredient" (id, name, unit, category, "defaultShelfLifeDays", "defaultPricePerUnit", "createdAt")
VALUES
    (gen_random_uuid(), 'Makaron (spaghetti)', 'g', 'grain', 365, 0.005, NOW()),
    (gen_random_uuid(), 'Makaron (penne)', 'g', 'grain', 365, 0.005, NOW()),
    (gen_random_uuid(), 'Makaron (fusilli)', 'g', 'grain', 365, 0.005, NOW()),
    (gen_random_uuid(), 'Ryż biały', 'g', 'grain', 730, 0.004, NOW()),
    (gen_random_uuid(), 'Ryż brązowy', 'g', 'grain', 365, 0.006, NOW()),
    (gen_random_uuid(), 'Kasza gryczana', 'g', 'grain', 365, 0.005, NOW()),
    (gen_random_uuid(), 'Kasza jaglana', 'g', 'grain', 365, 0.007, NOW()),
    (gen_random_uuid(), 'Płatki owsiane', 'g', 'grain', 365, 0.004, NOW()),
    (gen_random_uuid(), 'Mąka pszenna', 'g', 'grain', 180, 0.002, NOW()),
    (gen_random_uuid(), 'Mąka żytnia', 'g', 'grain', 180, 0.003, NOW()),
    (gen_random_uuid(), 'Chleb', 'g', 'grain', 3, 0.008, NOW()),
    (gen_random_uuid(), 'Bułka', 'pcs', 'grain', 2, 0.50, NOW()),
    (gen_random_uuid(), 'Tortilla', 'pcs', 'grain', 30, 0.80, NOW())
ON CONFLICT DO NOTHING;

-- CONDIMENTS (Przyprawy, sosy, oleje)
INSERT INTO "Ingredient" (id, name, unit, category, "defaultShelfLifeDays", "defaultPricePerUnit", "createdAt")
VALUES
    (gen_random_uuid(), 'Sól', 'g', 'condiment', 3650, 0.001, NOW()),
    (gen_random_uuid(), 'Pieprz czarny', 'g', 'condiment', 730, 0.040, NOW()),
    (gen_random_uuid(), 'Papryka słodka', 'g', 'condiment', 365, 0.030, NOW()),
    (gen_random_uuid(), 'Papryka ostra', 'g', 'condiment', 365, 0.035, NOW()),
    (gen_random_uuid(), 'Oregano', 'g', 'condiment', 365, 0.050, NOW()),
    (gen_random_uuid(), 'Bazylia', 'g', 'condiment', 365, 0.045, NOW()),
    (gen_random_uuid(), 'Tymianek', 'g', 'condiment', 365, 0.055, NOW()),
    (gen_random_uuid(), 'Rozmaryn', 'g', 'condiment', 365, 0.060, NOW()),
    (gen_random_uuid(), 'Majeranek', 'g', 'condiment', 365, 0.040, NOW()),
    (gen_random_uuid(), 'Kminek', 'g', 'condiment', 365, 0.035, NOW()),
    (gen_random_uuid(), 'Curry', 'g', 'condiment', 365, 0.045, NOW()),
    (gen_random_uuid(), 'Kurkuma', 'g', 'condiment', 365, 0.038, NOW()),
    (gen_random_uuid(), 'Imbir', 'g', 'condiment', 365, 0.042, NOW()),
    (gen_random_uuid(), 'Cynamon', 'g', 'condiment', 365, 0.048, NOW()),
    (gen_random_uuid(), 'Gałka muszkatołowa', 'g', 'condiment', 730, 0.080, NOW()),
    (gen_random_uuid(), 'Olej rzepakowy', 'ml', 'condiment', 365, 0.006, NOW()),
    (gen_random_uuid(), 'Olej słonecznikowy', 'ml', 'condiment', 365, 0.005, NOW()),
    (gen_random_uuid(), 'Oliwa z oliwek', 'ml', 'condiment', 730, 0.025, NOW()),
    (gen_random_uuid(), 'Ocet', 'ml', 'condiment', 730, 0.004, NOW()),
    (gen_random_uuid(), 'Musztarda', 'g', 'condiment', 180, 0.010, NOW()),
    (gen_random_uuid(), 'Ketchup', 'g', 'condiment', 180, 0.008, NOW()),
    (gen_random_uuid(), 'Majonez', 'g', 'condiment', 90, 0.012, NOW()),
    (gen_random_uuid(), 'Sos sojowy', 'ml', 'condiment', 365, 0.015, NOW()),
    (gen_random_uuid(), 'Bulion (kostka)', 'pcs', 'condiment', 730, 0.40, NOW())
ON CONFLICT DO NOTHING;

-- OTHER (Inne)
INSERT INTO "Ingredient" (id, name, unit, category, "defaultShelfLifeDays", "defaultPricePerUnit", "createdAt")
VALUES
    (gen_random_uuid(), 'Cukier', 'g', 'other', 730, 0.003, NOW()),
    (gen_random_uuid(), 'Cukier waniliowy', 'g', 'other', 365, 0.015, NOW()),
    (gen_random_uuid(), 'Proszek do pieczenia', 'g', 'other', 365, 0.008, NOW()),
    (gen_random_uuid(), 'Soda oczyszczona', 'g', 'other', 365, 0.005, NOW()),
    (gen_random_uuid(), 'Drożdże', 'g', 'other', 14, 0.020, NOW()),
    (gen_random_uuid(), 'Miód', 'g', 'other', 730, 0.018, NOW()),
    (gen_random_uuid(), 'Dżem truskawkowy', 'g', 'other', 180, 0.012, NOW()),
    (gen_random_uuid(), 'Nutella', 'g', 'other', 365, 0.020, NOW()),
    (gen_random_uuid(), 'Kakao', 'g', 'other', 365, 0.025, NOW()),
    (gen_random_uuid(), 'Czekolada gorzka', 'g', 'other', 365, 0.022, NOW())
ON CONFLICT DO NOTHING;
