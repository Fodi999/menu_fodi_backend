-- Проверяем оба ID растительного масла
SELECT 
    id,
    name,
    name_pl,
    name_en, 
    name_ru,
    canonical_name,
    unit
FROM "Ingredient"
WHERE id IN (
    '1b7cea8e-b026-4329-9d2e-c94952e3fa6c',  -- В холодильнике
    '9ff773d2-a3ee-4f4b-bc45-4cfe0d7f680b'   -- В рецепте
)
ORDER BY name;
