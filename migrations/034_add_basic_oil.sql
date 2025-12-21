-- Migration 034: Add basic oil (Olej roślinny)
-- Created: 2025-12-21
-- Purpose: Add generic vegetable oil for AI recipes

INSERT INTO "Ingredient" (
    id,
    name,
    category,
    unit,
    "defaultPricePerUnit",
    "createdAt"
)
VALUES (
    gen_random_uuid(),
    'Olej roślinny',
    'condiment',
    'ml',
    0.005,
    NOW()
)
ON CONFLICT (name) DO NOTHING;

-- 📌 0.005 PLN/ml — нормальная стартовая цена
-- 📌 ON CONFLICT — защита от повторного запуска
-- 📌 condiment — pantry категория (не скоропортящийся продукт)
