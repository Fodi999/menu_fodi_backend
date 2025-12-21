-- Migration: Add "Olej roślinny" (Vegetable oil) to ingredient catalog
-- Created: 2025-12-21
-- Purpose: Add generic vegetable oil for recipes that don't specify type

INSERT INTO "Ingredient" (id, name, unit, category, "defaultShelfLifeDays", "defaultPricePerUnit", "createdAt")
VALUES
    (gen_random_uuid(), 'Olej roślinny', 'ml', 'condiment', 365, 0.005, NOW())
ON CONFLICT (name) DO NOTHING;

-- Note: "Olej roślinny" = generic vegetable oil
-- If user has specific oil (rzepakowy, słonecznikowy), they should use that instead
-- Price: 0.005 PLN/ml (similar to sunflower oil)
-- Category: condiment (pantry item)
-- Shelf life: 365 days (1 year)
