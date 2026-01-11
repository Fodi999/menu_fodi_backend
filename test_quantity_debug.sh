#!/bin/bash

# 🔍 TEST SCRIPT: Debug Quantity Values
# Purpose: Send the EXACT frontend request to see what backend receives

echo "=========================================="
echo "🧪 Quantity Debug Test"
echo "=========================================="
echo ""

# Backend URL
API_URL="http://localhost:8080"

# Step 1: Login as admin
echo "Step 1: Logging in as admin..."
LOGIN_RESPONSE=$(curl -s -X POST "$API_URL/api/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "admin_password_123"
  }')

TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.token // .data.token // empty')

if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
  echo "❌ Login failed"
  echo "$LOGIN_RESPONSE" | jq .
  exit 1
fi

echo "✅ Login successful"
echo ""

# Step 2: Test with MICROSCOPIC quantities (like frontend sends)
echo "=========================================="
echo "🔬 Test 1: Microscopic Quantities (0.2g, 0.1ml)"
echo "=========================================="
echo ""

curl -X POST "$API_URL/api/admin/recipes/create-ai" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Accept-Language: ru-RU,ru;q=0.9" \
  -d '{
    "title": "Жареный лосось (микроскопический тест)",
    "rawCookingText": "Обжарить лосось на оливковом масле",
    "language": "ru",
    "ingredients": [
      {"ingredientId": "fe1c7431-b1b7-4d36-94bf-74276481983e", "quantity": 0.2, "unit": "g"},
      {"ingredientId": "9ff773d2-a3ee-4f4b-bc45-4cfe0d7f680b", "quantity": 0.1, "unit": "ml"}
    ]
  }' | jq .

echo ""
echo ""

# Step 3: Test with REALISTIC quantities
echo "=========================================="
echo "🍽️  Test 2: Realistic Quantities (150g, 20ml)"
echo "=========================================="
echo ""

curl -X POST "$API_URL/api/admin/recipes/create-ai" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Accept-Language: ru-RU,ru;q=0.9" \
  -d '{
    "title": "Жареный лосось (реалистичный тест)",
    "rawCookingText": "Обжарить лосось на оливковом масле",
    "language": "ru",
    "ingredients": [
      {"ingredientId": "fe1c7431-b1b7-4d36-94bf-74276481983e", "quantity": 150, "unit": "g"},
      {"ingredientId": "9ff773d2-a3ee-4f4b-bc45-4cfe0d7f680b", "quantity": 20, "unit": "ml"}
    ]
  }' | jq .

echo ""
echo ""
echo "=========================================="
echo "📋 Check Backend Logs"
echo "=========================================="
echo ""
echo "Run this command to see debug logs:"
echo "  tail -50 server_test.log | grep -E '📥|📤|Quantity|Amount'"
