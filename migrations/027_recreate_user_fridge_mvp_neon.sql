-- =====================================================
-- МИГРАЦИЯ 027: Пересоздание user_fridge_items (MVP)
-- ВЕРСИЯ ДЛЯ NEON.TECH (TEXT вместо UUID)
-- =====================================================

-- Удаляем старые таблицы
DROP TABLE IF EXISTS "UserFridgeItem" CASCADE;
DROP TABLE IF EXISTS "user_fridge_items" CASCADE;
DROP TABLE IF EXISTS user_fridge_items CASCADE;

-- Создаем новую таблицу с правильной MVP структурой
-- ⚠️ ВАЖНО: используем TEXT вместо UUID (Neon.tech специфика)
CREATE TABLE user_fridge_items (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    
    user_id TEXT NOT NULL
        REFERENCES "User"(id) ON DELETE CASCADE,
    
    ingredient_id TEXT NOT NULL
        REFERENCES "Ingredient"(id) ON DELETE CASCADE,
    
    quantity DOUBLE PRECISION NOT NULL DEFAULT 0,
    unit VARCHAR(50) NOT NULL DEFAULT 'шт',
    expires_at TIMESTAMP,  -- ✅ NULLABLE - может быть NULL
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(user_id, ingredient_id)
);

-- Создаем индексы для оптимизации
CREATE INDEX idx_user_fridge_items_user_id 
    ON user_fridge_items(user_id);

CREATE INDEX idx_user_fridge_items_ingredient_id 
    ON user_fridge_items(ingredient_id);

CREATE INDEX idx_user_fridge_items_expires_at 
    ON user_fridge_items(expires_at);

CREATE INDEX idx_user_fridge_items_user_expires 
    ON user_fridge_items(user_id, expires_at);

-- Добавляем комментарий
COMMENT ON TABLE user_fridge_items 
IS 'MVP: User fridge items with simplified structure (Neon.tech version)';

-- ✅ ПРОВЕРКА: Смотрим структуру таблицы
SELECT column_name, data_type, is_nullable
FROM information_schema.columns
WHERE table_name = 'user_fridge_items'
ORDER BY ordinal_position;
