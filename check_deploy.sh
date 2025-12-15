#!/bin/bash

# 🔍 Скрипт проверки деплоя на Koyeb
# Проверяет, работает ли исправленная версия кода

API_URL="https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app"

echo "🔍 Проверка деплоя backend на Koyeb..."
echo ""

# 1. Health check
echo "1️⃣ Health check:"
HEALTH=$(curl -s "$API_URL/health" | jq -r '.data.status' 2>/dev/null)
if [ "$HEALTH" = "ok" ]; then
    echo "   ✅ Backend запущен и работает"
else
    echo "   ❌ Backend не отвечает"
    exit 1
fi

# 2. Регистрация тестового пользователя
echo ""
echo "2️⃣ Проверка регистрации (роль home_chef):"
TEST_EMAIL="deploy_test_$(date +%s)@example.com"
REGISTER_RESPONSE=$(curl -s -X POST "$API_URL/api/auth/register" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"Test123!\",\"name\":\"Deploy Test\"}")

ROLE=$(echo "$REGISTER_RESPONSE" | jq -r '.data.user.role' 2>/dev/null)
TOKEN=$(echo "$REGISTER_RESPONSE" | jq -r '.data.token' 2>/dev/null)

if [ "$ROLE" = "home_chef" ]; then
    echo "   ✅ Регистрация работает, роль: $ROLE"
else
    echo "   ❌ Роль неправильная: $ROLE"
    echo "   Response: $REGISTER_RESPONSE"
    exit 1
fi

# 3. Проверка fridge API (КЛЮЧЕВОЙ ТЕСТ)
echo ""
echo "3️⃣ Проверка Fridge API (nullable ExpiresAt):"
FRIDGE_RESPONSE=$(curl -s -X GET "$API_URL/api/fridge/items" \
  -H "Authorization: Bearer $TOKEN")

SUCCESS=$(echo "$FRIDGE_RESPONSE" | jq -r '.success' 2>/dev/null)

if [ "$SUCCESS" = "true" ]; then
    ITEMS=$(echo "$FRIDGE_RESPONSE" | jq -r '.data.items | length' 2>/dev/null)
    echo "   ✅ Fridge API работает! Холодильник пустой (items: $ITEMS)"
    echo ""
    echo "🎉 ВСЕ ПРОВЕРКИ ПРОЙДЕНЫ! Деплой успешен!"
    echo ""
    echo "📊 Детали:"
    echo "   Email: $TEST_EMAIL"
    echo "   Role: $ROLE"
    echo "   Fridge items: $ITEMS"
else
    echo "   ❌ Fridge API НЕ РАБОТАЕТ"
    echo "   Response: $FRIDGE_RESPONSE"
    echo ""
    echo "⚠️  Возможные причины:"
    echo "   1. Koyeb ещё не задеплоил коммит f902bf4"
    echo "   2. Есть ошибка деплоя"
    echo "   3. Нужно выполнить миграцию 027 в БД"
    exit 1
fi
