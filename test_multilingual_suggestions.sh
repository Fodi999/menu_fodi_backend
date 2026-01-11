#!/bin/bash

# 🌍 TEST: Multilingual Conflict Resolution

echo "=========================================="
echo "🌍 Multilingual Suggestions Test"
echo "=========================================="
echo ""

API_URL="http://localhost:8080"

# Login
echo "Step 1: Logging in..."
LOGIN_RESPONSE=$(curl -s -X POST "$API_URL/api/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "admin_password_123"
  }')

TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.token // .data.token // empty')

if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
  echo "❌ Login failed"
  exit 1
fi

echo "✅ Login successful"
echo ""

# Test conflict to get multilingual suggestions
echo "=========================================="
echo "🌍 Trigger conflict for multilingual suggestions"
echo "=========================================="
echo ""

RESPONSE=$(curl -s -X POST "$API_URL/api/admin/recipes/save" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "title": "жареный лосось",
    "language": "ru",
    "description": "Test",
    "servings": 1,
    "time_minutes": 10,
    "difficulty": "easy",
    "calories": 300,
    "ingredients": [
      {
        "ingredientId": "fe1c7431-b1b7-4d36-94bf-74276481983e",
        "name": "Лосось",
        "amount": 150,
        "unit": "g"
      }
    ],
    "steps": [
      {
        "order": 1,
        "text": "Обжарить",
        "time": 10
      }
    ]
  }')

echo "$RESPONSE" | jq .
echo ""

# Check if multilingual suggestions present
HAS_RU=$(echo "$RESPONSE" | jq -e '.suggestions.ru' >/dev/null 2>&1 && echo "yes" || echo "no")
HAS_EN=$(echo "$RESPONSE" | jq -e '.suggestions.en' >/dev/null 2>&1 && echo "yes" || echo "no")
HAS_PL=$(echo "$RESPONSE" | jq -e '.suggestions.pl' >/dev/null 2>&1 && echo "yes" || echo "no")

echo "=========================================="
echo "📊 Multilingual Check"
echo "=========================================="
echo ""

if [ "$HAS_RU" = "yes" ] && [ "$HAS_EN" = "yes" ] && [ "$HAS_PL" = "yes" ]; then
  echo "✅ All languages present!"
  echo ""
  
  echo "🇷🇺 Russian suggestions:"
  echo "$RESPONSE" | jq -r '.suggestions.ru[]' | nl
  echo ""
  
  echo "🇬🇧 English suggestions:"
  echo "$RESPONSE" | jq -r '.suggestions.en[]' | nl
  echo ""
  
  echo "🇵🇱 Polish suggestions:"
  echo "$RESPONSE" | jq -r '.suggestions.pl[]' | nl
  echo ""
  
  # Count suggestions
  RU_COUNT=$(echo "$RESPONSE" | jq -r '.suggestions.ru | length')
  EN_COUNT=$(echo "$RESPONSE" | jq -r '.suggestions.en | length')
  PL_COUNT=$(echo "$RESPONSE" | jq -r '.suggestions.pl | length')
  
  echo "📈 Counts:"
  echo "   RU: $RU_COUNT suggestions"
  echo "   EN: $EN_COUNT suggestions"
  echo "   PL: $PL_COUNT suggestions"
  echo ""
  
  if [ "$RU_COUNT" -ge 3 ] && [ "$EN_COUNT" -ge 3 ] && [ "$PL_COUNT" -ge 3 ]; then
    echo "✅ Sufficient suggestions in all languages!"
  else
    echo "⚠️  Some languages have too few suggestions"
  fi
  
else
  echo "❌ Missing languages:"
  [ "$HAS_RU" = "no" ] && echo "   - Russian"
  [ "$HAS_EN" = "no" ] && echo "   - English"
  [ "$HAS_PL" = "no" ] && echo "   - Polish"
fi

echo ""
echo "=========================================="
echo "📋 SUMMARY"
echo "=========================================="
echo ""

if [ "$HAS_RU" = "yes" ] && [ "$HAS_EN" = "yes" ] && [ "$HAS_PL" = "yes" ]; then
  echo "🎉 Multilingual conflict resolution working!"
  echo ""
  echo "✅ Test 1: Russian suggestions - PASS"
  echo "✅ Test 2: English suggestions - PASS"
  echo "✅ Test 3: Polish suggestions - PASS"
else
  echo "❌ Multilingual suggestions not working properly"
fi

echo ""
echo "Check backend logs:"
echo "  tail -50 server_test.log | grep -E 'multilingual|🌍|Generated multilingual'"
