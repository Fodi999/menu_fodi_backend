#!/bin/bash

# Test AI Recipe Creation with Polish Language (LOCAL)
BASE_URL="http://localhost:8080/api"

echo "=== Testing AI Recipe Creation (Polish - LOCAL) ==="
echo ""

# Step 1: Login as admin
echo "1️⃣ Logging in as admin..."
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"dima_fomin@icloud.com","password":"Smirnof23"}')

echo "Login response: $LOGIN_RESPONSE"
TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.token')
echo "✅ Token: ${TOKEN:0:50}..."
echo ""

# Step 2: Get ingredient IDs (salmon, rice, soy sauce)
echo "2️⃣ Getting ingredient IDs..."
SALMON_RESPONSE=$(curl -s "$BASE_URL/ingredients/suggest?query=salmon" \
  -H "Authorization: Bearer $TOKEN")
echo "Salmon response: $SALMON_RESPONSE"
SALMON_ID=$(echo $SALMON_RESPONSE | jq -r '.[0].id')

RICE_RESPONSE=$(curl -s "$BASE_URL/ingredients/suggest?query=rice" \
  -H "Authorization: Bearer $TOKEN")
RICE_ID=$(echo $RICE_RESPONSE | jq -r '.[0].id')

SOY_RESPONSE=$(curl -s "$BASE_URL/ingredients/suggest?query=soy" \
  -H "Authorization: Bearer $TOKEN")
SOY_ID=$(echo $SOY_RESPONSE | jq -r '.[0].id')

echo "🐟 Salmon ID: $SALMON_ID"
echo "🍚 Rice ID: $RICE_ID"
echo "🥫 Soy Sauce ID: $SOY_ID"
echo ""

# Step 3: Preview AI Recipe (Polish)
echo "3️⃣ Creating AI recipe preview (Polish)..."
PREVIEW=$(curl -s -X POST "$BASE_URL/admin/recipes/preview-ai" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{
    \"title\": \"Łosoś z ryżem i sosem sojowym\",
    \"language\": \"pl\",
    \"ingredients\": [
      {\"ingredientId\": \"$SALMON_ID\", \"quantity\": 150, \"unit\": \"g\"},
      {\"ingredientId\": \"$RICE_ID\", \"quantity\": 100, \"unit\": \"g\"},
      {\"ingredientId\": \"$SOY_ID\", \"quantity\": 20, \"unit\": \"ml\"}
    ],
    \"rawCookingText\": \"Rybę opłukać i osuszyć. Grillować 5 minut z każdej strony. Ryż ugotować w osolonej wodzie przez 12 minut. Podawać z sosem sojowym.\"
  }")

echo "📋 Preview Response:"
echo "$PREVIEW" | jq '.'
echo ""

# Step 4: Check validation
TITLE_CHECK=$(echo "$PREVIEW" | jq -r '.title')
LANGUAGE_CHECK=$(echo "$PREVIEW" | jq -r '.language')
DESCRIPTION=$(echo "$PREVIEW" | jq -r '.description')
INGREDIENTS_COUNT=$(echo "$PREVIEW" | jq '.ingredients | length')

echo "✅ Validation Results:"
echo "   Title preserved: $TITLE_CHECK"
echo "   Language: $LANGUAGE_CHECK"
echo "   Description (PL): ${DESCRIPTION:0:100}..."
echo "   Ingredients count: $INGREDIENTS_COUNT"
echo ""

# Step 5: Check ingredients have IDs
echo "🔍 Checking ingredient IDs..."
echo "$PREVIEW" | jq '.ingredients[] | {id: .ingredientId, name: .name, amount: .amount, unit: .unit}'
echo ""

echo "✅ Test completed!"
