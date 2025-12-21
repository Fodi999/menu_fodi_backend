#!/bin/bash

# Full recipe flow test: Generate recipe → Check missing ingredients → Add to fridge
# User: testai@fodi.app

BASE_URL="https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app"

echo "================================"
echo "🧪 FULL RECIPE FLOW TEST"
echo "================================"
echo ""

# Step 1: Login
echo "Step 1: 🔑 Login as testai@fodi.app..."
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/api/auth/login" \
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

# Step 2: Check current fridge
echo "Step 2: 📦 Current fridge contents..."
FRIDGE_BEFORE=$(curl -s -X GET "$BASE_URL/api/fridge" \
  -H "Authorization: Bearer $TOKEN")

FRIDGE_COUNT=$(echo $FRIDGE_BEFORE | jq '.data.items | length')
echo "📊 Items in fridge: $FRIDGE_COUNT"
echo ""
echo "Current products:"
echo $FRIDGE_BEFORE | jq '.data.items[] | {name: .name, quantity: .quantity, unit: .unit, daysLeft: .daysLeft}' | head -15
echo ""

# Step 3: Generate recipe from fridge
echo "Step 3: 🍽️  Generate recipe from fridge..."
RECIPE_RESPONSE=$(curl -s -X POST "$BASE_URL/api/ai/create-recipe-from-fridge" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "language": "pl"
  }')

echo "📡 Recipe Response:"
echo $RECIPE_RESPONSE | jq '.'
echo ""

# Parse recipe details
SUCCESS=$(echo $RECIPE_RESPONSE | jq -r '.success')
RECIPE_NAME=$(echo $RECIPE_RESPONSE | jq -r '.data.recipe.name // .data.recipe.title')
INGREDIENTS_USED=$(echo $RECIPE_RESPONSE | jq -r '.data.recipe.ingredientsUsed | length')
INGREDIENTS_MISSING=$(echo $RECIPE_RESPONSE | jq -r '.data.recipe.ingredientsMissing | length')

echo "================================"
echo "📋 RECIPE SUMMARY:"
echo "================================"
echo "✅ Success: $SUCCESS"
echo "🍽️  Name: $RECIPE_NAME"
echo "✅ Ingredients from fridge: $INGREDIENTS_USED"
echo "🛒 Missing ingredients: $INGREDIENTS_MISSING"
echo ""

if [ "$INGREDIENTS_USED" != "null" ] && [ "$INGREDIENTS_USED" != "0" ]; then
  echo "✅ Ingredients USED (from fridge):"
  echo $RECIPE_RESPONSE | jq '.data.recipe.ingredientsUsed[] | "  - \(.name): \(.quantity) \(.unit)"' -r
  echo ""
fi

if [ "$INGREDIENTS_MISSING" != "null" ] && [ "$INGREDIENTS_MISSING" != "0" ]; then
  echo "🛒 Ingredients MISSING (to buy):"
  echo $RECIPE_RESPONSE | jq '.data.recipe.ingredientsMissing[] | "  - \(.name): \(.quantity) \(.unit)"' -r
  echo ""
else
  echo "✅ No missing ingredients! Recipe uses only fridge products."
  echo ""
fi

# Parse economy
ECONOMY=$(echo $RECIPE_RESPONSE | jq -r '.data.recipe.economy')
if [ "$ECONOMY" != "null" ]; then
  USED_VALUE=$(echo $RECIPE_RESPONSE | jq -r '.data.recipe.economy.usedValue')
  EXTRA_COST=$(echo $RECIPE_RESPONSE | jq -r '.data.recipe.economy.estimatedExtraCost')
  SAVED_MONEY=$(echo $RECIPE_RESPONSE | jq -r '.data.recipe.economy.savedMoney')
  CURRENCY=$(echo $RECIPE_RESPONSE | jq -r '.data.recipe.economy.currency')
  
  echo "💰 ECONOMY:"
  echo "  - Used from fridge: $USED_VALUE $CURRENCY"
  echo "  - Extra cost (pantry): $EXTRA_COST $CURRENCY"
  echo "  - Money saved: $SAVED_MONEY $CURRENCY"
  echo ""
fi

# Step 4: Add missing ingredients to fridge (if any)
if [ "$INGREDIENTS_MISSING" != "null" ] && [ "$INGREDIENTS_MISSING" != "0" ]; then
  echo "Step 4: 🛒 Adding missing ingredients to fridge..."
  
  # Extract ingredientsMissing array
  MISSING_JSON=$(echo $RECIPE_RESPONSE | jq '.data.recipe.ingredientsMissing')
  
  ADD_RESPONSE=$(curl -s -X POST "$BASE_URL/api/ai/add-missing-ingredients" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d "{\"ingredients\": $MISSING_JSON}")
  
  echo "📡 Add Response:"
  echo $ADD_RESPONSE | jq '.'
  echo ""
  
  ADDED=$(echo $ADD_RESPONSE | jq -r '.data.added')
  SKIPPED=$(echo $ADD_RESPONSE | jq -r '.data.skipped')
  MESSAGE=$(echo $ADD_RESPONSE | jq -r '.data.message')
  
  echo "📊 RESULT:"
  echo "  ✅ Added: $ADDED"
  echo "  ⚠️  Skipped: $SKIPPED"
  echo "  💬 Message: $MESSAGE"
  echo ""
  
  if [ "$SKIPPED" != "null" ] && [ "$SKIPPED" != "[]" ]; then
    echo "⚠️  Skipped ingredients (not found in catalog):"
    echo $ADD_RESPONSE | jq '.data.skipped[]' -r | while read item; do
      echo "  - $item"
    done
    echo ""
  fi
else
  echo "Step 4: ⏭️  Skipped - no missing ingredients to add"
  echo ""
fi

# Step 5: Check fridge after adding
echo "Step 5: 📦 Fridge contents AFTER adding ingredients..."
FRIDGE_AFTER=$(curl -s -X GET "$BASE_URL/api/fridge" \
  -H "Authorization: Bearer $TOKEN")

FRIDGE_COUNT_AFTER=$(echo $FRIDGE_AFTER | jq '.data.items | length')
echo "📊 Items in fridge: $FRIDGE_COUNT_AFTER (was: $FRIDGE_COUNT, added: $(($FRIDGE_COUNT_AFTER - $FRIDGE_COUNT)))"
echo ""

if [ "$INGREDIENTS_MISSING" != "null" ] && [ "$INGREDIENTS_MISSING" != "0" ]; then
  echo "🔍 Checking if missing ingredients were added..."
  echo $RECIPE_RESPONSE | jq '.data.recipe.ingredientsMissing[].name' -r | while read ingredient; do
    FOUND=$(echo $FRIDGE_AFTER | jq ".data.items[] | select(.name == \"$ingredient\") | .name" -r)
    if [ ! -z "$FOUND" ]; then
      QUANTITY=$(echo $FRIDGE_AFTER | jq ".data.items[] | select(.name == \"$ingredient\") | .quantity" -r)
      UNIT=$(echo $FRIDGE_AFTER | jq ".data.items[] | select(.name == \"$ingredient\") | .unit" -r)
      echo "  ✅ $ingredient: $QUANTITY $UNIT (FOUND in fridge)"
    else
      echo "  ❌ $ingredient: NOT FOUND (may be skipped or different name)"
    fi
  done
  echo ""
fi

echo "================================"
echo "✅ FULL FLOW TEST COMPLETE!"
echo "================================"
echo ""
echo "Summary:"
echo "  1. ✅ Login successful"
echo "  2. ✅ Recipe generated: $RECIPE_NAME"
echo "  3. ✅ Used $INGREDIENTS_USED products from fridge"
echo "  4. ✅ Found $INGREDIENTS_MISSING missing ingredients"
if [ "$INGREDIENTS_MISSING" != "null" ] && [ "$INGREDIENTS_MISSING" != "0" ]; then
  echo "  5. ✅ Added missing ingredients to fridge"
fi
echo ""
