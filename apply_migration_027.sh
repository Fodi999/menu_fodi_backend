#!/bin/bash

# 🚀 Применение миграции 027 в production БД
# Этот скрипт создаёт таблицу user_fridge_items с правильной структурой

set -e  # Остановить при ошибке

echo "🔥 ПРИМЕНЕНИЕ МИГРАЦИИ 027 В PRODUCTION"
echo "========================================"
echo ""

# Render PostgreSQL connection string
DB_HOST="dpg-csekngjtq21c73ctcqh0-a.oregon-postgres.render.com"
DB_PORT="5432"
DB_USER="menu_fodi_backend_user"
DB_PASS="u1YW6TRGKa0nC2ZZKGVY0OJZXfNUIomJ"
DB_NAME="menu_fodi_backend"

export PGPASSWORD="$DB_PASS"

echo "📡 Подключение к production БД..."
echo "   Host: $DB_HOST"
echo "   Database: $DB_NAME"
echo ""

# Проверяем подключение
if ! psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT 1" > /dev/null 2>&1; then
    echo "❌ Не удалось подключиться к БД"
    echo ""
    echo "Используй Render Dashboard вместо этого:"
    echo "1. https://dashboard.render.com/"
    echo "2. Выбери базу menu_fodi_backend"
    echo "3. Вкладка 'Query'"
    echo "4. Скопируй содержимое migrations/027_recreate_user_fridge_mvp.sql"
    exit 1
fi

echo "✅ Подключение успешно"
echo ""

# Применяем миграцию
echo "🔄 Применение миграции..."
psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
  -f migrations/027_recreate_user_fridge_mvp.sql

echo ""
echo "✅ Миграция применена!"
echo ""

# Проверка результата
echo "🔍 Проверка структуры таблицы:"
psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
  -c "SELECT column_name, data_type, is_nullable FROM information_schema.columns WHERE table_name = 'user_fridge_items' ORDER BY ordinal_position;"

echo ""
echo "🔍 Проверка записей:"
COUNT=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
  -t -c "SELECT COUNT(*) FROM user_fridge_items;" | xargs)

echo "   Записей в таблице: $COUNT"

if [ "$COUNT" = "0" ]; then
    echo "   ✅ Таблица пустая - это нормально для нового деплоя"
else
    echo "   ℹ️  Есть $COUNT записей"
fi

echo ""
echo "🎉 МИГРАЦИЯ ЗАВЕРШЕНА УСПЕШНО!"
echo ""
echo "📝 Следующие шаги:"
echo "   1. Зайди в Koyeb Dashboard"
echo "   2. Нажми 'Redeploy' на сервисе yeasty-madelaine-fodi999-671ccdf5"
echo "   3. Подожди 2-3 минуты"
echo "   4. Проверь /api/fridge/items"
