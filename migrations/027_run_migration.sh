#!/bin/bash

# 🚀 Скрипт для запуска миграции 027
# Создаёт таблицу user_fridge_items в production database

DATABASE_URL="postgresql://menu_fodi_backend_user:u1YW6TRGKa0nC2ZZKGVY0OJZXfNUIomJ@dpg-csekngjtq21c73ctcqh0-a.oregon-postgres.render.com:5432/menu_fodi_backend"

echo "🔄 Запуск миграции 027: user_fridge_items..."
echo ""

# Запускаем SQL из файла
PGSSLMODE=require PGPASSWORD=u1YW6TRGKa0nC2ZZKGVY0OJZXfNUIomJ psql \
  -h dpg-csekngjtq21c73ctcqh0-a.oregon-postgres.render.com \
  -p 5432 \
  -U menu_fodi_backend_user \
  -d menu_fodi_backend \
  -v ON_ERROR_STOP=1 \
  << 'EOF'

-- Drop old tables
DROP TABLE IF EXISTS "UserFridgeItem" CASCADE;
DROP TABLE IF EXISTS "user_fridge_items" CASCADE;
DROP TABLE IF EXISTS user_fridge_items CASCADE;

-- Create new table
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
    
    UNIQUE(user_id, ingredient_id)
);

-- Create indexes
CREATE INDEX idx_user_fridge_items_user_id 
    ON user_fridge_items(user_id);

CREATE INDEX idx_user_fridge_items_ingredient_id 
    ON user_fridge_items(ingredient_id);

CREATE INDEX idx_user_fridge_items_expires_at 
    ON user_fridge_items(expires_at);

CREATE INDEX idx_user_fridge_items_user_expires 
    ON user_fridge_items(user_id, expires_at);

-- Add comment
COMMENT ON TABLE user_fridge_items 
IS 'MVP: User fridge items with simplified structure';

-- Verify table structure
SELECT column_name, data_type, is_nullable
FROM information_schema.columns
WHERE table_name = 'user_fridge_items'
ORDER BY ordinal_position;

EOF

if [ $? -eq 0 ]; then
    echo ""
    echo "✅ Миграция выполнена успешно!"
    echo ""
    echo "🔍 Проверка Foreign Keys:"
    
    PGSSLMODE=require PGPASSWORD=u1YW6TRGKa0nC2ZZKGVY0OJZXfNUIomJ psql \
      -h dpg-csekngjtq21c73ctcqh0-a.oregon-postgres.render.com \
      -p 5432 \
      -U menu_fodi_backend_user \
      -d menu_fodi_backend \
      -c "SELECT tc.constraint_name, kcu.column_name, ccu.table_name AS foreign_table 
          FROM information_schema.table_constraints tc
          JOIN information_schema.key_column_usage kcu ON tc.constraint_name = kcu.constraint_name
          JOIN information_schema.constraint_column_usage ccu ON ccu.constraint_name = tc.constraint_name
          WHERE tc.table_name = 'user_fridge_items' AND tc.constraint_type = 'FOREIGN KEY';"
    
    echo ""
    echo "🎉 Теперь можно тестировать API!"
    echo ""
    echo "📝 Следующие шаги:"
    echo "1. Обнови страницу /fridge"
    echo "2. Ошибка 500 должна исчезнуть"
else
    echo ""
    echo "❌ Ошибка при выполнении миграции"
    echo "Проверь логи выше"
fi
