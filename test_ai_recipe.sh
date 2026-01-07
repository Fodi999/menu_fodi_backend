#!/bin/bash

# Test AI Recipe Creation Endpoints
# Usage: ./test_ai_recipe.sh

TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiI3ZWM4YWJhNC04MTk1LTRiZTEtYTlhOC0wNjdjMzBhYWUzMDYiLCJlbWFpbCI6ImFkbWluQGV4YW1wbGUuY29tIiwicm9sZSI6InN1cGVyX2FkbWluIiwiZXhwIjoxNzY3ODY2MjUxLCJpYXQiOjE3Njc3Nzk4NTF9.5jw2OA_DZ5qJv-cSrTORbxZV_X_QK_J3VCEbjI4Fqx8"
BASE_URL="http://localhost:8080"

echo "🧪 Testing AI Recipe Endpoints"
echo "================================"
echo ""

# Get salmon and rice ingredient IDs
echo "📋 Step 1: Get ingredient IDs..."
SALMON_ID=$(curl -s "$BASE_URL/api/admin/ingredients/suggest?q=salmon&limit=1" \
  -H "Authorization: Bearer $TOKEN" | jq -r '.data[0].id')

RICE_ID=$(curl -s "$BASE_URL/api/admin/ingredients/suggest?q=rice&limit=1" \
  -H "Authorization: Bearer $TOKEN" | jq -r '.data[0].id')

echo "✅ Salmon ID: $SALMON_ID"
echo "✅ Rice ID: $RICE_ID"
echo ""

# Test Preview (no save)
echo "🔍 Step 2: Test Preview Endpoint (no save)"
echo "POST /api/admin/recipes/preview-ai"
echo ""

curl -X POST "$BASE_URL/api/admin/recipes/preview-ai" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Salmon with Rice and Teriyaki Sauce",
    "ingredients": [
      {"ingredientId": "'$SALMON_ID'", "quantity": 150, "unit": "g"},
      {"ingredientId": "'$RICE_ID'", "quantity": 100, "unit": "g"}
    ],
    "rawCookingText": "Marinate salmon in teriyaki sauce for 10 minutes. Grill salmon for 7 minutes. Boil rice until tender."
  }' | jq '.'

echo ""
echo "================================"
echo ""

# Test Create (with save)
echo "💾 Step 3: Test Create Endpoint (saves to DB)"
echo "POST /api/admin/recipes/create-ai"
echo ""

curl -X POST "$BASE_URL/api/admin/recipes/create-ai" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Grilled Salmon with Jasmine Rice",
    "ingredients": [
      {"ingredientId": "'$SALMON_ID'", "quantity": 200, "unit": "g"},
      {"ingredientId": "'$RICE_ID'", "quantity": 150, "unit": "g"}
    ],
    "rawCookingText": "Season salmon with salt and pepper. Grill for 8 minutes. Cook rice according to package instructions."
  }' | jq '.'

echo ""
echo "================================"
echo "✅ Tests complete!"
