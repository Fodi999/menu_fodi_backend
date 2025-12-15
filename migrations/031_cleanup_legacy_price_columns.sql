-- =====================================================
-- МИГРАЦИЯ 031: Удаление legacy price columns
-- =====================================================
-- Дата: 15 декабря 2025
-- Описание: Удаляем устаревшие поля price_per_unit и currency,
--           которые конфликтуют с новой архитектурой event sourcing

-- Проблема:
-- В user_fridge_items были поля price_per_unit и currency (migration 029)
-- Они конфликтуют с новой системой event sourcing (migration 030)
-- 
-- Источник истины теперь: user_fridge_price_history
-- Кэш для UI: current_price_per_unit, current_price_currency, price_updated_at

-- Удаляем legacy поля
ALTER TABLE user_fridge_items
DROP COLUMN IF EXISTS price_per_unit,
DROP COLUMN IF EXISTS currency;

-- Проверяем что осталось только правильные поля
COMMENT ON COLUMN user_fridge_items.current_price_per_unit 
IS 'Cache: latest price from history (denormalized for performance)';

COMMENT ON COLUMN user_fridge_items.current_price_currency 
IS 'Cache: currency from latest price event';

COMMENT ON COLUMN user_fridge_items.price_updated_at 
IS 'Cache: timestamp of last price update';

-- ✅ ПРОВЕРКА: Убедимся что legacy полей больше нет
SELECT column_name, data_type
FROM information_schema.columns
WHERE table_name = 'user_fridge_items'
  AND column_name LIKE '%price%'
ORDER BY ordinal_position;

-- Должно остаться только:
-- current_price_per_unit | double precision
-- current_price_currency | text
-- price_updated_at       | timestamp without time zone
