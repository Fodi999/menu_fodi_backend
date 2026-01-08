#!/bin/bash

# Quick AI Recipe Test
BASE_URL="http://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api"

echo "=== Quick AI Recipe Test ==="
echo ""

# Login
echo "1️⃣ Logging in..."
LOGIN=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"admin_password_123"}')

echo "Response: $LOGIN"
TOKEN=$(echo "$LOGIN" | jq -r '.data.token // .token // empty')
echo "Token: ${TOKEN:0:50}..."
echo ""

if [ "$TOKEN" == "null" ] || [ -z "$TOKEN" ]; then
    echo "❌ Login failed. Response:"
    echo "$LOGIN"
    exit 1
fi

# Get some ingredient IDs
echo "2️⃣ Getting ingredient IDs..."
INGREDIENTS=$(curl -s "$BASE_URL/admin/ingredients?page=1&limit=50" \
  -H "Authorization: Bearer $TOKEN")
echo "Ingredients response (first 200 chars): ${INGREDIENTS:0:200}..."

# Try to find specific ingredients, or just use first one
TOMATO_ID=$(echo "$INGREDIENTS" | jq -r '.data[] | select(.name | ascii_downcase | contains("tomato") or contains("pomid")) | .id' | head -1)
ONION_ID=$(echo "$INGREDIENTS" | jq -r '.data[] | select(.name | ascii_downcase | contains("onion") or contains("cebul")) | .id' | head -1)

# Fallback to first two ingredients if not found
if [ -z "$TOMATO_ID" ]; then
    TOMATO_ID=$(echo "$INGREDIENTS" | jq -r '.data[0].id')
fi
if [ -z "$ONION_ID" ]; then
    ONION_ID=$(echo "$INGREDIENTS" | jq -r '.data[1].id')
fi

echo "Ingredient 1 ID: $TOMATO_ID"
echo "Ingredient 2 ID: $ONION_ID"
echo ""

if [ -z "$TOMATO_ID" ]; then
    echo "❌ Failed to get ingredient IDs"
    exit 1
fi

# Test preview endpoint
echo "3️⃣ Testing AI Recipe Preview..."
PREVIEW=$(curl -s -X POST "$BASE_URL/admin/recipes/preview-ai" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{
    \"title\": \"Simple Vegetable Dish\",
    \"language\": \"en\",
    \"ingredients\": [
      {\"ingredientId\": \"$TOMATO_ID\", \"quantity\": 200, \"unit\": \"g\"},
      {\"ingredientId\": \"$ONION_ID\", \"quantity\": 100, \"unit\": \"g\"}
    ],
    \"rawCookingText\": \"Chop vegetables. Cook for 10 minutes.\"
  }")

echo ""
echo "📋 AI Response:"
echo "$PREVIEW" | jq '.'
echo ""

# Check validation
TITLE=$(echo "$PREVIEW" | jq -r '.data.title // .title // "null"')
LANG=$(echo "$PREVIEW" | jq -r '.data.language // .language // "null"')
DESC=$(echo "$PREVIEW" | jq -r '.data.description // .description // "null"')

echo "✅ Validation:"
echo "   Title: $TITLE"
echo "   Language: $LANG"
echo "   Description: ${DESC:0:100}..."
echo ""

if [ "$TITLE" == "Simple Vegetable Dish" ] && [ "$LANG" == "en" ]; then
    echo "✅ SUCCESS! AI preserved all data correctly!"
else
    echo "❌ FAILED! Data was modified by AI"
fi
