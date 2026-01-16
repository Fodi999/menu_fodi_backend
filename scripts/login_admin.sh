#!/bin/bash

# Логин как администратор (super_admin)

BASE_URL="${BASE_URL:-https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app}"

echo "🔐 Logging in as super_admin..."
echo ""

RESPONSE=$(curl -s -X POST "$BASE_URL/api/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "admin_password_123"
  }')

echo "$RESPONSE" | jq '.'

# Извлекаем токен (пробуем разные пути)
TOKEN=$(echo "$RESPONSE" | jq -r '.data.token // .data.accessToken // .token // empty')

if [ -n "$TOKEN" ]; then
  echo ""
  echo "✅ Login successful!"
  echo ""
  echo "📋 Export token:"
  echo "export ADMIN_TOKEN=\"$TOKEN\""
  echo ""
  echo "Or copy this command:"
  echo "ADMIN_TOKEN=\"$TOKEN\""
else
  echo ""
  echo "❌ Login failed!"
fi
