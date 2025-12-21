#!/bin/bash

# Add test products and run full recipe flow
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiI1NWViMmRiOC1hYWViLTQ1NWEtODFkMy1hZGNjMTEzNjI0ZWYiLCJlbWFpbCI6InJlY2lwZXRlc3RAZm9kaS5hcHAiLCJyb2xlIjoiaG9tZV9jaGVmIiwiZXhwIjoxNzY2NDA2MDE2LCJpYXQiOjE3NjYzMTk2MTZ9.EnB54La8LmmHG0BQ6G3UAjDHcXwUbr-YPTnCKG4nZYg"
BASE_URL="https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app"

echo "Step 1: Adding test products to fridge..."

# Wołowina
curl -s -X POST "$BASE_URL/api/fridge" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "Wołowina",
    "quantity": 400,
    "unit": "g",
    "expiresAt": "2025-12-22T12:00:00Z"
  }' | jq '.'

# Cebula
curl -s -X POST "$BASE_URL/api/fridge" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "Cebula",
    "quantity": 200,
    "unit": "g",
    "expiresAt": "2025-12-24T12:00:00Z"
  }' | jq '.'

# Mleko
curl -s -X POST "$BASE_URL/api/fridge" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "Mleko 3.2%",
    "quantity": 500,
    "unit": "ml",
    "expiresAt": "2025-12-26T12:00:00Z"
  }' | jq '.'

echo ""
echo "✅ Added 3 products to fridge"
echo ""

# Step 2: Generate recipe
echo "Step 2: Generating recipe from fridge..."
RECIPE_RESPONSE=$(curl -s -X POST "$BASE_URL/api/ai/create-recipe-from-fridge" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"language": "pl"}')

echo "📡 Full Recipe Response:"
echo $RECIPE_RESPONSE | jq '.'
echo ""

# Parse key data
RECIPE_NAME=$(echo $RECIPE_RESPONSE | jq -r '.data.recipe.name // .data.recipe.title')
echo "🍽️  RECIPE: $RECIPE_NAME"
echo ""

echo "✅ Ingredients USED (from fridge):"
echo $RECIPE_RESPONSE | jq '.data.recipe.ingredientsUsed[]' -c
echo ""

echo "🛒 Ingredients MISSING (to buy):"
MISSING_COUNT=$(echo $RECIPE_RESPONSE | jq '.data.recipe.ingredientsMissing | length')
echo "Count: $MISSING_COUNT"
echo $RECIPE_RESPONSE | jq '.data.recipe.ingredientsMissing[]' -c
echo ""

# Step 3: Add missing ingredients
if [ "$MISSING_COUNT" != "0" ]; then
  echo "Step 3: Adding missing ingredients to fridge..."
  MISSING_JSON=$(echo $RECIPE_RESPONSE | jq '.data.recipe.ingredientsMissing')
  
  ADD_RESPONSE=$(curl -s -X POST "$BASE_URL/api/ai/add-missing-ingredients" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d "{\"ingredients\": $MISSING_JSON}")
  
  echo "📡 Add Response:"
  echo $ADD_RESPONSE | jq '.'
  echo ""
fi

echo "✅ TEST COMPLETE!"
