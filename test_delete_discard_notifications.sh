#!/bin/bash

# Тест уведомлений при удалении и выбросе продуктов

BASE_URL="https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app"
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiI0MDc1ODJiZS01OWQ1LTRkMjEtODczYi0xYTcyZDMxYjBkNDIiLCJlbWFpbCI6ImZvZGk4NUBnbWFpbC5ydSIsInJvbGUiOiJob21lX2NoZWYiLCJleHAiOjE3Njg1NjAyOTcsImlhdCI6MTc2ODQ3Mzg5N30.wWSasbP-1WVnVIst7_HMCpAXNRfwAHWQIUhqKEorHYY"

echo "=========================================="
echo "🧪 Тест уведомлений при удалении/выбросе"
echo "=========================================="
echo ""

# 1. Проверяем начальное количество уведомлений
echo "1️⃣  Начальное количество уведомлений:"
INITIAL_COUNT=$(curl -s "$BASE_URL/api/notifications/unread-count" \
  -H "Authorization: Bearer $TOKEN" | jq -r '.count')
echo "   Уведомлений: $INITIAL_COUNT"
echo ""

# 2. Добавляем продукт для теста удаления
echo "2️⃣  Добавляем Czosnek (чеснок) для теста DELETE..."
ITEM_1=$(curl -s -X POST "$BASE_URL/api/fridge/items" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"ingredientId":"2c3405e0-60cf-4e5f-9872-0bb8d1f91b83","quantity":5.0}' | jq -r '.data.id')
echo "   ✅ Добавлен: $ITEM_1"
echo ""

sleep 2

# 3. Удаляем продукт (DELETE)
echo "3️⃣  Удаляем продукт (DELETE /api/fridge/items/$ITEM_1)..."
curl -s -X DELETE "$BASE_URL/api/fridge/items/$ITEM_1" \
  -H "Authorization: Bearer $TOKEN" | jq '.'
echo ""

sleep 1

# 4. Проверяем количество уведомлений после DELETE
echo "4️⃣  Количество уведомлений после DELETE:"
COUNT_AFTER_DELETE=$(curl -s "$BASE_URL/api/notifications/unread-count" \
  -H "Authorization: Bearer $TOKEN" | jq -r '.count')
echo "   Уведомлений: $COUNT_AFTER_DELETE (было: $((INITIAL_COUNT + 1)))"
echo ""

# 5. Добавляем продукт с ценой для теста DISCARD
echo "5️⃣  Добавляем Jajka (яйца) с ценой для теста DISCARD..."
ITEM_2=$(curl -s -X POST "$BASE_URL/api/fridge/items" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "ingredientId":"b4e0c3d2-6f9a-4e5b-8c7d-1a2b3c4d5e6f",
    "quantity":12,
    "unit":"szt",
    "priceTotal":15.50
  }')
ITEM_2_ID=$(echo "$ITEM_2" | jq -r '.data.id')
echo "   ✅ Добавлено: $ITEM_2_ID (15.50 PLN)"
echo ""

sleep 2

# 6. Выбрасываем продукт (DISCARD)
echo "6️⃣  Выбрасываем продукт (POST /api/fridge/items/$ITEM_2_ID/discard)..."
curl -s -X POST "$BASE_URL/api/fridge/items/$ITEM_2_ID/discard" \
  -H "Authorization: Bearer $TOKEN" | jq '.'
echo ""

sleep 1

# 7. Финальная проверка количества уведомлений
echo "7️⃣  Финальное количество уведомлений:"
FINAL_COUNT=$(curl -s "$BASE_URL/api/notifications/unread-count" \
  -H "Authorization: Bearer $TOKEN" | jq -r '.count')
echo "   Уведомлений: $FINAL_COUNT (было: $INITIAL_COUNT)"
echo ""

# 8. Проверяем последние 3 уведомления
echo "8️⃣  Последние уведомления:"
curl -s "$BASE_URL/api/notifications" \
  -H "Authorization: Bearer $TOKEN" | jq '.data[0:3] | .[] | {
    type: .type,
    level: .level,
    title: .title,
    message: .message,
    meta: .meta
  }'
echo ""

# 9. Итоги
echo "=========================================="
echo "📊 Итоги теста:"
echo "=========================================="
echo "Начальное количество:  $INITIAL_COUNT"
echo "После добавления #1:   $((INITIAL_COUNT + 1))"
echo "После DELETE:          $COUNT_AFTER_DELETE (ожидалось: $((INITIAL_COUNT + 2)))"
echo "После добавления #2:   $((COUNT_AFTER_DELETE + 1))"
echo "После DISCARD:         $FINAL_COUNT (ожидалось: $((INITIAL_COUNT + 4)))"
echo ""

EXPECTED=$((INITIAL_COUNT + 4))
if [ "$FINAL_COUNT" -eq "$EXPECTED" ]; then
  echo "✅ ТЕСТ ПРОЙДЕН! Все 4 уведомления созданы!"
else
  echo "❌ ТЕСТ НЕ ПРОЙДЕН. Ожидалось $EXPECTED, получено $FINAL_COUNT"
fi
echo ""
