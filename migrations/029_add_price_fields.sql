-- =====================================================
-- МИГРАЦИЯ 029: Добавление полей цены
-- =====================================================
-- Дата: 15 декабря 2025
-- Описание: Добавляем price_per_unit и currency для расчёта стоимости продуктов

-- Добавляем поля цены
ALTER TABLE user_fridge_items
ADD COLUMN price_per_unit DOUBLE PRECISION,
ADD COLUMN currency TEXT DEFAULT 'PLN';

-- Создаём индекс для фильтрации по валюте (если понадобится)
CREATE INDEX idx_user_fridge_items_currency 
    ON user_fridge_items(currency);

-- Добавляем комментарии
COMMENT ON COLUMN user_fridge_items.price_per_unit 
IS 'Normalized price per unit (always per g/ml/szt, never kg/l)';

COMMENT ON COLUMN user_fridge_items.currency 
IS 'Currency code (PLN, EUR, USD, etc.)';

-- ✅ ПРОВЕРКА: Смотрим обновлённую структуру
SELECT column_name, data_type, is_nullable, column_default
FROM information_schema.columns
WHERE table_name = 'user_fridge_items'
  AND column_name IN ('price_per_unit', 'currency')
ORDER BY ordinal_position;
