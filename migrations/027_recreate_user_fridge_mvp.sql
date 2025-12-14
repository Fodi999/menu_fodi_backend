-- =====================================================
-- ПЕРЕСОЗДАНИЕ ТАБЛИЦЫ user_fridge_items (MVP VERSION)
-- =====================================================
-- Удаляем старую таблицу и создаем новую с правильной структурой
-- MVP: id, user_id, ingredient_id, quantity, unit, expires_at, created_at
-- =====================================================

-- Удаляем старую таблицу если существует
DROP TABLE IF EXISTS "UserFridgeItem" CASCADE;
DROP TABLE IF EXISTS "user_fridge_items" CASCADE;
DROP TABLE IF EXISTS user_fridge_items CASCADE;

-- Создаем новую таблицу с MVP структурой
CREATE TABLE user_fridge_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    user_id UUID NOT NULL
        REFERENCES "User"(id) ON DELETE CASCADE,
    
    ingredient_id UUID NOT NULL
        REFERENCES "Ingredient"(id) ON DELETE CASCADE,
    
    quantity DOUBLE PRECISION NOT NULL DEFAULT 0,
    unit VARCHAR(50) NOT NULL DEFAULT 'шт',
    expires_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Ensure each user can have only one entry per ingredient
    UNIQUE(user_id, ingredient_id)
);

-- Создаем индексы для оптимизации запросов
CREATE INDEX idx_user_fridge_items_user_id 
    ON user_fridge_items(user_id);

CREATE INDEX idx_user_fridge_items_ingredient_id 
    ON user_fridge_items(ingredient_id);

CREATE INDEX idx_user_fridge_items_expires_at 
    ON user_fridge_items(expires_at);

CREATE INDEX idx_user_fridge_items_user_expires 
    ON user_fridge_items(user_id, expires_at);

-- Комментарии к таблице
COMMENT ON TABLE user_fridge_items 
IS 'MVP: User fridge items with simplified structure';

-- ✅ Verification queries (run these after migration)
-- Check table exists:
-- SELECT to_regclass('public.user_fridge_items');

-- Check structure:
-- SELECT column_name, data_type, is_nullable
-- FROM information_schema.columns
-- WHERE table_name = 'user_fridge_items'
-- ORDER BY ordinal_position;

-- Check foreign keys:
-- SELECT tc.constraint_name, kcu.column_name, ccu.table_name AS foreign_table
-- FROM information_schema.table_constraints tc
-- JOIN information_schema.key_column_usage kcu ON tc.constraint_name = kcu.constraint_name
-- JOIN information_schema.constraint_column_usage ccu ON ccu.constraint_name = tc.constraint_name
-- WHERE tc.table_name = 'user_fridge_items' AND tc.constraint_type = 'FOREIGN KEY';
