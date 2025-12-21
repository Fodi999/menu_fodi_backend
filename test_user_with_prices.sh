#!/bin/bash

# Test recipe generation for users who HAVE prices in database
# Based on SQL query results showing existing prices

API_URL="https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app"

echo "=== Testing Users with Existing Prices ==="
echo ""
echo "From DB query we know these users have products WITH PRICES:"
echo "  - fodi85@gmail.ru: Wołowina (20.56), Ogórek (7.00), Cebula (3.45), Mleko (3.24)"
echo "  - maks@gmail.com: Cebula (5.00), Mleko kokosowe (3.00)"
echo "  - test_ai@fodi.app: Kurczak (15.00)"
echo ""

# Test with fodi85@gmail.ru (has most products with prices)
echo "=== Test 1: fodi85@gmail.ru (has 4 products with prices) ==="
echo "Login..."
LOGIN_RESPONSE=$(curl -s -X POST "$API_URL/api/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email": "fodi85@gmail.ru", "password": "12345678"}')

TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.data.token // empty')

if [ -z "$TOKEN" ]; then
  echo "❌ Login failed for fodi85@gmail.ru"
  echo "$LOGIN_RESPONSE" | jq '.'
  echo ""
  echo "Try different password or check user exists"
  exit 1
fi

echo "✅ Logged in successfully"
USER_ID=$(echo "$LOGIN_RESPONSE" | jq -r '.data.user.id')
echo "User ID: $USER_ID"

echo ""
echo "--- Checking fridge contents ---"
FRIDGE=$(curl -s "$API_URL/api/fridge" -H "Authorization: Bearer $TOKEN")
echo "$FRIDGE" | jq '.data.items[] | {name: .ingredient.name, quantity, unit, pricePerUnit: .currentPricePerUnit, currency: .currentPriceCurrency}'

echo ""
echo "--- Generating recipe from fridge ---"
RECIPE=$(curl -s -X POST "$API_URL/api/ai/create-recipe-from-fridge" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"language": "pl"}')

echo "$RECIPE" | jq '.'

echo ""
echo "=== ECONOMY CHECK ==="
ECONOMY=$(echo "$RECIPE" | jq '.data.recipe.economy')
echo "$ECONOMY" | jq '.'

USED_VALUE=$(echo "$ECONOMY" | jq -r '.usedValue // 0')
echo ""
if (( $(echo "$USED_VALUE > 0" | bc -l) )); then
  echo "✅✅✅ SUCCESS! Economy calculated with real prices!"
  echo "   usedValue: $USED_VALUE PLN"
  echo ""
  echo "🎉 Backend is working correctly!"
else
  echo "❌❌❌ FAILED! usedValue still = 0"
  echo ""
  echo "⚠️  CRITICAL: Prices exist in DB but not being used!"
  echo ""
  echo "🔍 Check Koyeb logs for these lines:"
  echo "   1. 'Loaded fridge items with prices' - how many items loaded?"
  echo "   2. 'Fridge item price' - does it show current_price_per_unit values?"
  echo "   3. 'Price data found for item' - are prices detected?"
  echo "   4. '[AI][ECONOMY] Used cost:' - final calculation result"
  echo ""
  echo "If logs show 'No price data' → GORM mapping issue"
  echo "If logs show prices → Service calculation issue"
fi

echo ""
echo "=== USED PRODUCTS ==="
echo "$RECIPE" | jq '.data.usedProducts[]? | {name, quantityUsed, pricePerUnit, usedCost, currency}'

echo ""
echo "=== Full Economy Object ==="
echo "$ECONOMY" | jq '.'

echo ""
echo "Token for manual testing: $TOKEN"
echo "User: fodi85@gmail.ru"
echo ""
echo "📋 Next: Check Koyeb logs at:"
echo "   https://app.koyeb.com/ → Your Service → Logs"
echo "   Search for: user_id=$USER_ID"
