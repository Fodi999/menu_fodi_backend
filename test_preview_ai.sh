#!/bin/bash

# 🔍 TEST SCRIPT: Preview Recipe (AI generation WITHOUT saving to DB)
# Purpose: Test AI recipe preview with realistic quantities

echo "=========================================="
echo "🎨 Recipe Preview Test (No DB Save)"
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

# Step 2: Test Preview with Microscopic Quantities
echo "=========================================="
echo "🔬 Test 1: Preview with Microscopic Quantities (0.2g, 0.1ml)"
echo "=========================================="
echo ""

PREVIEW_MICRO=$(curl -s -X POST "$API_URL/api/admin/recipes/preview-ai" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Accept-Language: ru-RU,ru;q=0.9" \
  -d '{
    "title": "Жареный лосось (микро-превью)",
    "rawCookingText": "Обжарить лосось на оливковом масле до золотистой корочки",
    "language": "ru",
    "ingredients": [
      {"ingredientId": "fe1c7431-b1b7-4d36-94bf-74276481983e", "quantity": 0.2, "unit": "g"},
      {"ingredientId": "9ff773d2-a3ee-4f4b-bc45-4cfe0d7f680b", "quantity": 0.1, "unit": "ml"}
    ]
  }')

echo "$PREVIEW_MICRO" | jq .

# Extract key metrics
CALORIES_MICRO=$(echo "$PREVIEW_MICRO" | jq -r '.data.nutritionProfile.calories // .preview.nutritionProfile.calories // 0')
STEPS_COUNT_MICRO=$(echo "$PREVIEW_MICRO" | jq -r '.data.steps | length // .preview.steps | length // 0')

echo ""
echo "📊 Metrics (Microscopic):"
echo "   Calories: $CALORIES_MICRO"
echo "   Steps: $STEPS_COUNT_MICRO"
echo ""

sleep 2

# Step 3: Test Preview with Realistic Quantities
echo "=========================================="
echo "🍽️  Test 2: Preview with Realistic Quantities (150g, 20ml)"
echo "=========================================="
echo ""

PREVIEW_REAL=$(curl -s -X POST "$API_URL/api/admin/recipes/preview-ai" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Accept-Language: ru-RU,ru;q=0.9" \
  -d '{
    "title": "Жареный лосось (реальный-превью)",
    "rawCookingText": "Обжарить лосось на оливковом масле до золотистой корочки",
    "language": "ru",
    "ingredients": [
      {"ingredientId": "fe1c7431-b1b7-4d36-94bf-74276481983e", "quantity": 150, "unit": "g"},
      {"ingredientId": "9ff773d2-a3ee-4f4b-bc45-4cfe0d7f680b", "quantity": 20, "unit": "ml"}
    ]
  }')

echo "$PREVIEW_REAL" | jq .

# Extract key metrics
CALORIES_REAL=$(echo "$PREVIEW_REAL" | jq -r '.data.nutritionProfile.calories // .preview.nutritionProfile.calories // 0')
STEPS_COUNT_REAL=$(echo "$PREVIEW_REAL" | jq -r '.data.steps | length // .preview.steps | length // 0')

echo ""
echo "📊 Metrics (Realistic):"
echo "   Calories: $CALORIES_REAL"
echo "   Steps: $STEPS_COUNT_REAL"
echo ""

# Step 4: Test Preview with Complex Recipe (Multi-language)
echo "=========================================="
echo "🌍 Test 3: Complex Recipe (Polish Language)"
echo "=========================================="
echo ""

PREVIEW_COMPLEX=$(curl -s -X POST "$API_URL/api/admin/recipes/preview-ai" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Accept-Language: pl-PL,pl;q=0.9" \
  -d '{
    "title": "Łosoś z ryżem i sosem teriyaki",
    "rawCookingText": "Łosoś marynować w sosie teriyaki przez 30 minut. Ryż ugotować. Łosoś usmażyć na patelni. Podawać z ryżem i warzywami.",
    "language": "pl",
    "ingredients": [
      {"ingredientId": "fe1c7431-b1b7-4d36-94bf-74276481983e", "quantity": 200, "unit": "g"},
      {"ingredientId": "9ff773d2-a3ee-4f4b-bc45-4cfe0d7f680b", "quantity": 15, "unit": "ml"}
    ]
  }')

echo "$PREVIEW_COMPLEX" | jq .

CALORIES_COMPLEX=$(echo "$PREVIEW_COMPLEX" | jq -r '.data.nutritionProfile.calories // .preview.nutritionProfile.calories // 0')

echo ""
echo "📊 Metrics (Complex):"
echo "   Calories: $CALORIES_COMPLEX"
echo ""

# Summary
echo "=========================================="
echo "📈 SUMMARY"
echo "=========================================="
echo ""
echo "Test 1 (Microscopic 0.2g):  $CALORIES_MICRO kcal"
echo "Test 2 (Realistic 150g):    $CALORIES_REAL kcal"
echo "Test 3 (Complex 200g):      $CALORIES_COMPLEX kcal"
echo ""
echo "✅ Expected pattern: Calories increase with quantity"
echo "   0.2g < 150g < 200g"
echo ""

# Validation
if [ "$CALORIES_MICRO" -lt "$CALORIES_REAL" ] && [ "$CALORIES_REAL" -lt "$CALORIES_COMPLEX" ]; then
  echo "✅ PASS: Quantity affects AI calculations correctly"
else
  echo "⚠️  WARNING: Unexpected calorie pattern detected"
fi

echo ""
echo "=========================================="
echo "📋 Check Backend Logs for Debug Info"
echo "=========================================="
echo ""
echo "Run this command to see detailed logs:"
echo "  tail -100 server_test.log | grep -E '📥|📤|Preview|микро|реальный|Łosoś'"
