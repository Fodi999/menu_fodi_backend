-- =====================================================
-- ПЕРЕСОЗДАНИЕ ТАБЛИЦЫ user_fridge_items (MVP VERSION)
-- =====================================================
-- Удаляем старую таблицу и создаем новую с правильной структурой
-- MVP: id, user_id, ingredient_id, quantity, unit, expires_at, created_at
-- =====================================================

-- Удаляем старую таблицу если существует
DROP TABLE IF EXISTS "UserFridgeItem" CASCADE;
DROP TABLE IF EXISTS "user_fridge_items" CASCADE;

-- Создаем новую таблицу с MVP структурой
CREATE TABLE IF NOT EXISTS user_fridge_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES "User"(id) ON DELETE CASCADE,
    ingredient_id UUID NOT NULL REFERENCES "Ingredient"(id) ON DELETE CASCADE,
    quantity DOUBLE PRECISION NOT NULL CHECK (quantity > 0),
    unit VARCHAR(10) NOT NULL, -- 'g', 'ml', 'pcs'
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Создаем индексы для оптимизации запросов
CREATE INDEX idx_user_fridge_items_user_id ON user_fridge_items(user_id);
CREATE INDEX idx_user_fridge_items_ingredient_id ON user_fridge_items(ingredient_id);
CREATE INDEX idx_user_fridge_items_expires_at ON user_fridge_items(expires_at);
CREATE INDEX idx_user_fridge_items_user_expires ON user_fridge_items(user_id, expires_at);

-- Комментарии к таблице и колонкам
COMMENT ON TABLE user_fridge_items IS 'Холодильник домашнего повара - MVP версия';
COMMENT ON COLUMN user_fridge_items.id IS 'Уникальный идентификатор записи';
COMMENT ON COLUMN user_fridge_items.user_id IS 'ID пользователя (владелец холодильника)';
COMMENT ON COLUMN user_fridge_items.ingredient_id IS 'ID ингредиента из каталога';
COMMENT ON COLUMN user_fridge_items.quantity IS 'Количество продукта (числовое значение)';
COMMENT ON COLUMN user_fridge_items.unit IS 'Единица измерения (g, ml, pcs) - копия из каталога';
COMMENT ON COLUMN user_fridge_items.expires_at IS 'Дата истечения срока годности';
COMMENT ON COLUMN user_fridge_items.created_at IS 'Дата добавления продукта в холодильник';

-- Проверка структуры
SELECT 
    column_name, 
    data_type, 
    is_nullable,
    column_default
FROM information_schema.columns
WHERE table_name = 'user_fridge_items'
ORDER BY ordinal_position;
