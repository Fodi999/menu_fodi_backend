-- =====================================================
-- МИГРАЦИЯ 030: История цен (Event Sourcing)
-- =====================================================
-- Дата: 15 декабря 2025
-- Описание: Система отслеживания истории изменения цен продуктов

-- 1. Создаём таблицу истории цен (event sourcing)
CREATE TABLE user_fridge_price_history (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    
    user_fridge_item_id TEXT NOT NULL
        REFERENCES user_fridge_items(id) ON DELETE CASCADE,
    
    price_per_unit DOUBLE PRECISION NOT NULL,
    currency TEXT NOT NULL DEFAULT 'PLN',
    
    -- Источник данных о цене
    source TEXT NOT NULL DEFAULT 'manual',
    -- Возможные значения:
    -- 'manual'   - ручной ввод пользователем
    -- 'receipt'  - из чека (OCR)
    -- 'estimate' - оценка AI
    -- 'market'   - рыночные данные
    -- 'ai'       - рекомендация AI
    
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Индексы для быстрого поиска
CREATE INDEX idx_fridge_price_item_id 
    ON user_fridge_price_history(user_fridge_item_id);

CREATE INDEX idx_fridge_price_created_at 
    ON user_fridge_price_history(created_at DESC);

CREATE INDEX idx_fridge_price_source 
    ON user_fridge_price_history(source);

-- 2. Добавляем кэш текущей цены в основную таблицу (денормализация для производительности)
ALTER TABLE user_fridge_items
ADD COLUMN current_price_per_unit DOUBLE PRECISION,
ADD COLUMN current_price_currency TEXT DEFAULT 'PLN',
ADD COLUMN price_updated_at TIMESTAMP;

-- Индекс для поиска по текущей цене
CREATE INDEX idx_user_fridge_current_price 
    ON user_fridge_items(current_price_per_unit) 
    WHERE current_price_per_unit IS NOT NULL;

-- Комментарии
COMMENT ON TABLE user_fridge_price_history 
IS 'Event sourcing: complete price change history with source tracking';

COMMENT ON COLUMN user_fridge_price_history.source 
IS 'Price data source: manual | receipt | estimate | market | ai';

COMMENT ON COLUMN user_fridge_items.current_price_per_unit 
IS 'Cache: current price (denormalized from history for performance)';

-- ✅ ПРОВЕРКА: Смотрим структуру таблиц
SELECT 
    table_name,
    column_name, 
    data_type, 
    is_nullable
FROM information_schema.columns
WHERE table_name IN ('user_fridge_price_history', 'user_fridge_items')
  AND column_name LIKE '%price%'
ORDER BY table_name, ordinal_position;
