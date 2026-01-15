#!/bin/bash

# Test instant notification when adding product to fridge

echo "🧪 Testing Instant Notification Feature"
echo "========================================"

# 1. Login and get token
echo ""
echo "Step 1: Getting auth token..."
TOKEN=$(curl -s "https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"fodi85@gmail.ru","password":"dima"}' | jq -r '.data.token')

if [ "$TOKEN" = "null" ] || [ -z "$TOKEN" ]; then
  echo "❌ Failed to get token"
  exit 1
fi

echo "✅ Token received: ${TOKEN:0:20}..."

# 2. Check current notification count
echo ""
echo "Step 2: Checking current notification count..."
BEFORE_COUNT=$(curl -s "https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/notifications/unread-count" \
  -H "Authorization: Bearer $TOKEN" | jq -r '.count')

echo "Current notifications: $BEFORE_COUNT"

# 3. Find an ingredient to add
echo ""
echo "Step 3: Finding ingredient (Cebula czerwona - red onion)..."
source <(grep DATABASE_URL .env | grep -v UNPOOLED)
INGREDIENT_ID=$(psql "$DATABASE_URL" -t -c "SELECT id FROM \"Ingredient\" WHERE name ILIKE '%cebul%czerwon%' LIMIT 1;" | tr -d ' ')

if [ -z "$INGREDIENT_ID" ]; then
  echo "❌ Ingredient not found, using Czosnek (garlic)..."
  INGREDIENT_ID=$(psql "$DATABASE_URL" -t -c "SELECT id FROM \"Ingredient\" WHERE name = 'Czosnek' LIMIT 1;" | tr -d ' ')
fi

echo "Using ingredient ID: $INGREDIENT_ID"

# 4. Add product to fridge
echo ""
echo "Step 4: Adding product to fridge..."
ADD_RESPONSE=$(curl -s -X POST "https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/fridge/items" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"ingredientId\": \"$INGREDIENT_ID\",
    \"quantity\": 1.3
  }")

echo "$ADD_RESPONSE" | jq '.'

if [ "$(echo $ADD_RESPONSE | jq -r '.success')" = "true" ]; then
  echo "✅ Product added successfully!"
else
  echo "❌ Failed to add product"
  exit 1
fi

# 5. Wait a moment for notification to be created
echo ""
echo "Step 5: Waiting 2 seconds for notification creation..."
sleep 2

# 6. Check notification count again
echo ""
echo "Step 6: Checking notification count after adding..."
AFTER_COUNT=$(curl -s "https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/notifications/unread-count" \
  -H "Authorization: Bearer $TOKEN" | jq -r '.count')

echo "New notifications: $AFTER_COUNT"
echo "Difference: +$(($AFTER_COUNT - $BEFORE_COUNT))"

# 7. Get latest notifications
echo ""
echo "Step 7: Fetching latest notifications..."
curl -s "https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/notifications" \
  -H "Authorization: Bearer $TOKEN" | jq '.data[0:3] | .[] | {type, level, title, message, createdAt}'

# 8. Result
echo ""
echo "========================================"
if [ $AFTER_COUNT -gt $BEFORE_COUNT ]; then
  echo "✅ TEST PASSED: Notification created instantly!"
  echo "   Badge will show: 🔔($AFTER_COUNT)"
else
  echo "⚠️  TEST INCONCLUSIVE: No new notification detected"
  echo "   Check logs on Koyeb for errors"
fi
echo "========================================"
