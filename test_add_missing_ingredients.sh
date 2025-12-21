#!/bin/bash

# Test adding missing ingredients from recipe to fridge
# User: testai@fodi.app

API_URL="https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/ai/add-missing-ingredients"

# Login first to get token
echo "🔑 Logging in as testai@fodi.app..."
LOGIN_RESPONSE=$(curl -s -X POST \
  https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "testai@fodi.app",
    "password": "test123456"
  }')

TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.data.token')

if [ "$TOKEN" == "null" ] || [ -z "$TOKEN" ]; then
  echo "❌ Login failed!"
  echo "Response: $LOGIN_RESPONSE"
  exit 1
fi

echo "✅ Token obtained: ${TOKEN:0:20}..."
echo ""

# Test 1: Add missing ingredients (Olej roślinny, Sól)
echo "📦 Test 1: Adding missing ingredients from recipe..."
echo "Ingredients: Olej roślinny (15ml), Sól (3g)"
echo ""

RESPONSE=$(curl -s -X POST $API_URL \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "ingredients": [
      {
        "name": "Olej roślinny",
        "quantity": 15,
        "unit": "ml"
      },
      {
        "name": "Sól",
        "quantity": 3,
        "unit": "g"
      }
    ]
  }')

echo "📡 Response:"
echo $RESPONSE | jq '.'
echo ""

# Parse response
SUCCESS=$(echo $RESPONSE | jq -r '.success')
ADDED=$(echo $RESPONSE | jq -r '.data.added')
SKIPPED=$(echo $RESPONSE | jq -r '.data.skipped')
MESSAGE=$(echo $RESPONSE | jq -r '.data.message')

echo "✅ Success: $SUCCESS"
echo "📊 Added: $ADDED"
echo "⚠️  Skipped: $SKIPPED"
echo "💬 Message: $MESSAGE"
echo ""

# Test 2: Try adding non-existent ingredient
echo "📦 Test 2: Adding non-existent ingredient (should skip)..."
echo "Ingredient: Unicorn Magic Powder (fictional)"
echo ""

RESPONSE2=$(curl -s -X POST $API_URL \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "ingredients": [
      {
        "name": "Unicorn Magic Powder",
        "quantity": 100,
        "unit": "g"
      }
    ]
  }')

echo "📡 Response:"
echo $RESPONSE2 | jq '.'
echo ""

# Test 3: Check fridge to verify items were added
echo "🔍 Test 3: Checking fridge contents..."
FRIDGE_RESPONSE=$(curl -s -X GET \
  https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/fridge \
  -H "Authorization: Bearer $TOKEN")

echo "📦 Current fridge items:"
echo $FRIDGE_RESPONSE | jq '.data.items[] | {name: .name, quantity: .quantity, unit: .unit}' | head -20
echo ""

echo "✅ Test complete!"
