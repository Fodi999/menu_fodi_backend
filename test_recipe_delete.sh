#!/bin/bash

# 🗑️ TEST: Recipe Deletion

BASE_URL="http://localhost:8080"
API="${BASE_URL}/api/admin/recipes"

echo "🗑️  Testing Recipe Deletion"
echo "==========================="
echo ""

# Сначала создадим тестовый рецепт для удаления
echo "📝 STEP 1: Creating test recipe..."
CREATE_RESPONSE=$(curl -s -X POST "${API}/save" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Тестовый Рецепт Для Удаления",
    "language": "ru",
    "description": "Этот рецепт будет удалён в тесте",
    "difficulty": "easy",
    "timeMinutes": 10,
    "servings": 1,
    "calories": 200,
    "ingredients": [
      {
        "ingredientId": "fe1c7431-b1b7-4d36-94bf-74276481983e",
        "quantity": 100
      }
    ],
    "steps": [
      {"text": "Тестовый шаг", "order": 1}
    ]
  }')

RECIPE_ID=$(echo "$CREATE_RESPONSE" | jq -r '.data.id // empty')

if [ -z "$RECIPE_ID" ]; then
  echo "❌ FAIL: Could not create test recipe"
  echo "Response: $CREATE_RESPONSE"
  exit 1
fi

echo "✅ Test recipe created: $RECIPE_ID"
echo ""

# Проверим что рецепт существует
echo "📋 STEP 2: Verifying recipe exists..."
VERIFY_RESPONSE=$(curl -s "${BASE_URL}/api/public/recipes" | jq --arg id "$RECIPE_ID" '.data[] | select(.id == $id)')

if [ -n "$VERIFY_RESPONSE" ]; then
  echo "✅ Recipe found in catalog"
else
  echo "⚠️  Recipe not found (may need admin auth)"
fi
echo ""

# Удаляем рецепт
echo "🗑️  STEP 3: Deleting recipe..."
DELETE_RESPONSE=$(curl -s -w "\nHTTP_CODE:%{http_code}" -X DELETE "${API}/${RECIPE_ID}")
DELETE_CODE=$(echo "$DELETE_RESPONSE" | grep "HTTP_CODE" | cut -d':' -f2)
DELETE_BODY=$(echo "$DELETE_RESPONSE" | sed '/HTTP_CODE/d')

echo "Response code: $DELETE_CODE"
echo "Response body: $DELETE_BODY"

if [ "$DELETE_CODE" == "200" ]; then
  SUCCESS=$(echo "$DELETE_BODY" | jq -r '.success')
  MESSAGE=$(echo "$DELETE_BODY" | jq -r '.message')
  
  if [ "$SUCCESS" == "true" ]; then
    echo "✅ PASS: Recipe deleted successfully"
    echo "Message: $MESSAGE"
  else
    echo "❌ FAIL: Success flag is false"
  fi
else
  echo "❌ FAIL: Expected 200, got $DELETE_CODE"
fi
echo ""

# Проверим что рецепт удалён
echo "🔍 STEP 4: Verifying recipe is deleted..."
CHECK_RESPONSE=$(curl -s "${BASE_URL}/api/public/recipes" | jq --arg id "$RECIPE_ID" '.data[] | select(.id == $id)')

if [ -z "$CHECK_RESPONSE" ]; then
  echo "✅ PASS: Recipe not found in catalog (deleted)"
else
  echo "❌ FAIL: Recipe still exists!"
  echo "$CHECK_RESPONSE"
fi
echo ""

# Test 2: Удаление несуществующего рецепта (404)
echo "🔍 TEST 2: Delete non-existent recipe"
FAKE_ID="00000000-0000-0000-0000-000000000000"
NOT_FOUND=$(curl -s -w "\nHTTP_CODE:%{http_code}" -X DELETE "${API}/${FAKE_ID}")
NOT_FOUND_CODE=$(echo "$NOT_FOUND" | grep "HTTP_CODE" | cut -d':' -f2)

if [ "$NOT_FOUND_CODE" == "404" ]; then
  echo "✅ PASS: 404 returned for non-existent recipe"
else
  echo "❌ FAIL: Expected 404, got $NOT_FOUND_CODE"
fi
echo ""

# Test 3: Удаление без ID (400)
echo "🔍 TEST 3: Delete without ID"
NO_ID=$(curl -s -w "\nHTTP_CODE:%{http_code}" -X DELETE "${API}/")
NO_ID_CODE=$(echo "$NO_ID" | grep "HTTP_CODE" | cut -d':' -f2)

if [ "$NO_ID_CODE" == "404" ] || [ "$NO_ID_CODE" == "400" ]; then
  echo "✅ PASS: Error returned when no ID provided ($NO_ID_CODE)"
else
  echo "⚠️  WARNING: Got code $NO_ID_CODE"
fi
echo ""

echo "==========================="
echo "✅ Recipe Deletion Tests Complete!"
echo "==========================="
