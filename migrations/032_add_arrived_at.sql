-- =====================================================
-- МИГРАЦИЯ 032: Добавление даты поступления продукта
-- =====================================================
-- Дата: 15 декабря 2025
-- Описание: Каждый продукт должен знать когда он попал в холодильник

-- Добавляем arrived_at (дата поступления)
ALTER TABLE user_fridge_items
ADD COLUMN IF NOT EXISTS arrived_at TIMESTAMP NOT NULL 
    DEFAULT CURRENT_TIMESTAMP;

-- Индекс для сортировки по дате поступления (новые сверху)
CREATE INDEX IF NOT EXISTS idx_user_fridge_arrived_at 
    ON user_fridge_items(arrived_at DESC);

-- Комментарий
COMMENT ON COLUMN user_fridge_items.arrived_at 
IS 'When product arrived to fridge (auto-set on insert, cannot be changed)';

-- ✅ ПРОВЕРКА: Смотрим все даты в таблице
SELECT 
    column_name, 
    data_type, 
    is_nullable,
    column_default
FROM information_schema.columns
WHERE table_name = 'user_fridge_items'
  AND (column_name LIKE '%_at' OR column_name = 'created_at')
ORDER BY ordinal_position;
