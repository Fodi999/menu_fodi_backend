-- =====================================================
-- МИГРАЦИЯ 030: История цен для продуктов в холодильнике
-- =====================================================
-- Дата: 15 декабря 2025
-- Описание: Добавляем таблицу user_fridge_price_history для трекинга изменений цен
-- Это event sourcing подход: каждое обновление цены = новое событие

-- 1. Создаём таблицу истории цен
CREATE TABLE user_fridge_price_history (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    
    user_fridge_item_id TEXT NOT NULL
        REFERENCES user_fridge_items(id) ON DELETE CASCADE,
    
    price_per_unit DOUBLE PRECISION NOT NULL,
    currency TEXT NOT NULL DEFAULT 'PLN',
    
    -- Источник цены: manual (ручной ввод), receipt (чек), estimate (оценка), market (рынок), ai (AI)
    source TEXT NOT NULL DEFAULT 'manual',
    
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 2. Создаём индексы для быстрого поиска
CREATE INDEX idx_fridge_price_item_id
    ON user_fridge_price_history(user_fridge_item_id);

CREATE INDEX idx_fridge_price_created_at
    ON user_fridge_price_history(created_at DESC);

CREATE INDEX idx_fridge_price_source
    ON user_fridge_price_history(source);

-- 3. Добавляем кеш текущей цены в основную таблицу (для производительности)
ALTER TABLE user_fridge_items
ADD COLUMN current_price_per_unit DOUBLE PRECISION,
ADD COLUMN current_currency TEXT DEFAULT 'PLN',
ADD COLUMN price_updated_at TIMESTAMP;

-- 4. Создаём индекс для фильтрации по наличию цены
CREATE INDEX idx_fridge_items_has_price
    ON user_fridge_items(current_price_per_unit) WHERE current_price_per_unit IS NOT NULL;

-- 5. Добавляем комментарии
COMMENT ON TABLE user_fridge_price_history 
IS 'Price history tracking for fridge items (event sourcing pattern)';

COMMENT ON COLUMN user_fridge_price_history.source 
IS 'manual | receipt | estimate | market | ai';

COMMENT ON COLUMN user_fridge_items.current_price_per_unit 
IS 'Cached latest price from price_history (denormalized for performance)';

-- ✅ ПРОВЕРКА: Смотрим структуру новой таблицы
SELECT column_name, data_type, is_nullable, column_default
FROM information_schema.columns
WHERE table_name = 'user_fridge_price_history'
ORDER BY ordinal_position;
