#!/bin/bash

# Complete test: Create user, add products WITH PRICES, generate recipe
# This will verify the full price flow end-to-end

API_URL="https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app"

echo "=== STEP 1: Register new test user with prices ==="
REGISTER_RESPONSE=$(curl -s -X POST "$API_URL/api/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "pricetest@fodi.app",
    "password": "Test1234!",
    "name": "Price Test User"
  }')

echo "$REGISTER_RESPONSE" | jq '.'
TOKEN=$(echo "$REGISTER_RESPONSE" | jq -r '.data.token')

if [ "$TOKEN" == "null" ] || [ -z "$TOKEN" ]; then
  echo "❌ Registration failed!"
  exit 1
fi

echo ""
echo "✅ User registered, token: ${TOKEN:0:50}..."
USER_ID=$(echo "$REGISTER_RESPONSE" | jq -r '.data.user.id')
echo "✅ User ID: $USER_ID"

echo ""
echo "=== STEP 2: Get ingredient IDs from catalog ==="

# Search for common ingredients
BEEF_ID=$(curl -s "$API_URL/api/ingredients/search?query=wołowina" \
  -H "Authorization: Bearer $TOKEN" | jq -r '.data.items[0].id // empty')

MILK_ID=$(curl -s "$API_URL/api/ingredients/search?query=mleko" \
  -H "Authorization: Bearer $TOKEN" | jq -r '.data.items[0].id // empty')

SALT_ID=$(curl -s "$API_URL/api/ingredients/search?query=sól" \
  -H "Authorization: Bearer $TOKEN" | jq -r '.data.items[0].id // empty')

echo "Wołowina ID: $BEEF_ID"
echo "Mleko ID: $MILK_ID"
echo "Sól ID: $SALT_ID"

echo ""
echo "=== STEP 3: Add products WITH PRICES to fridge ==="

# Add Wołowina (beef) - 500g at 20.56 PLN/kg = 0.02056 PLN/g
if [ ! -z "$BEEF_ID" ]; then
  echo ""
  echo "Adding Wołowina (500g, 20.56 PLN/kg)..."
  curl -s -X POST "$API_URL/api/fridge" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
      \"ingredientId\": \"$BEEF_ID\",
      \"quantity\": 500,
      \"unit\": \"g\",
      \"expiresAt\": \"2025-12-25T00:00:00Z\",
      \"priceInput\": {
        \"value\": 20.56,
        \"per\": \"kg\"
      }
    }" | jq '{success, data: {message}}'
fi

# Add Mleko (milk) - 1000ml at 3.24 PLN/l = 0.00324 PLN/ml
if [ ! -z "$MILK_ID" ]; then
  echo ""
  echo "Adding Mleko (1000ml, 3.24 PLN/l)..."
  curl -s -X POST "$API_URL/api/fridge" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
      \"ingredientId\": \"$MILK_ID\",
      \"quantity\": 1000,
      \"unit\": \"ml\",
      \"expiresAt\": \"2025-12-23T00:00:00Z\",
      \"priceInput\": {
        \"value\": 3.24,
        \"per\": \"l\"
      }
    }" | jq '{success, data: {message}}'
fi

# Add Sól (salt) - 200g at 2.50 PLN/kg = 0.0025 PLN/g
if [ ! -z "$SALT_ID" ]; then
  echo ""
  echo "Adding Sól (200g, 2.50 PLN/kg)..."
  curl -s -X POST "$API_URL/api/fridge" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
      \"ingredientId\": \"$SALT_ID\",
      \"quantity\": 200,
      \"unit\": \"g\",
      \"expiresAt\": \"2026-12-31T00:00:00Z\",
      \"priceInput\": {
        \"value\": 2.50,
        \"per\": \"kg\"
      }
    }" | jq '{success, data: {message}}'
fi

echo ""
echo "=== STEP 4: Verify fridge contents ==="
curl -s "$API_URL/api/fridge" \
  -H "Authorization: Bearer $TOKEN" | jq '.data.items[] | {name: .ingredient.name, quantity, unit, currentPricePerUnit, currency}'

echo ""
echo "=== STEP 5: Generate recipe from fridge ==="
RECIPE_RESPONSE=$(curl -s -X POST "$API_URL/api/ai/create-recipe-from-fridge" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"language": "pl"}')

echo "$RECIPE_RESPONSE" | jq '.'

echo ""
echo "=== STEP 6: Check Economy Calculation ==="
ECONOMY=$(echo "$RECIPE_RESPONSE" | jq '.data.recipe.economy')
echo "$ECONOMY" | jq '.'

USED_VALUE=$(echo "$ECONOMY" | jq -r '.usedValue // 0')
echo ""
if (( $(echo "$USED_VALUE > 0" | bc -l) )); then
  echo "✅ SUCCESS! Economy calculated with real prices!"
  echo "   usedValue: $USED_VALUE PLN"
else
  echo "❌ FAILED! usedValue = 0"
  echo "   → Check debug logs in Koyeb"
  echo "   → Prices may not be loading from DB"
fi

echo ""
echo "=== STEP 7: Check Used Products ==="
echo "$RECIPE_RESPONSE" | jq '.data.usedProducts[] | {name, quantityUsed, pricePerUnit, usedCost, currency}'

echo ""
echo "Test completed!"
echo "Token for future tests: $TOKEN"
echo "User ID: $USER_ID"
