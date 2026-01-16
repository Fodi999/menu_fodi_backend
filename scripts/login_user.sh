#!/bin/bash

# Логин как обычный пользователь (home_chef)

BASE_URL="${BASE_URL:-https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app}"

echo "🔐 Logging in as home_chef..."
echo ""

RESPONSE=$(curl -s -X POST "$BASE_URL/api/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "fodi85@gmail.ru",
    "password": "210185"
  }')

echo "$RESPONSE" | jq '.'

# Извлекаем токен (пробуем разные пути)
TOKEN=$(echo "$RESPONSE" | jq -r '.data.token // .data.accessToken // .token // empty')

if [ -n "$TOKEN" ]; then
  echo ""
  echo "✅ Login successful!"
  echo ""
  echo "📋 Export token:"
  echo "export USER_TOKEN=\"$TOKEN\""
  echo ""
  echo "Or copy this command:"
  echo "USER_TOKEN=\"$TOKEN\""
else
  echo ""
  echo "❌ Login failed!"
fi
