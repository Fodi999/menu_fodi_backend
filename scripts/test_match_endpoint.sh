#!/bin/bash

# Quick test for /api/recipes/match endpoint

BASE_URL="https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app"

echo "🔐 Logging in..."
USER_TOKEN=$(curl -s -X POST "$BASE_URL/api/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"fodi85@gmail.ru","password":"210185"}' | jq -r '.data.token')

if [ "$USER_TOKEN" == "null" ] || [ -z "$USER_TOKEN" ]; then
  echo "❌ Login failed"
  exit 1
fi

echo "✅ Token obtained"
echo ""
echo "📋 Testing /api/recipes/match..."
echo ""

curl -s "$BASE_URL/api/recipes/match?onlyCookable=true&minScore=100&limit=5" \
  -H "Authorization: Bearer $USER_TOKEN" | jq '{
    success,
    count: .data.count,
    recipes: .data.recipes | map({
      name: .localName,
      canCookNow,
      coverage,
      score,
      missing: (.missingIngredients | length),
      used: (.usedIngredients | length)
    })
  }'

echo ""
echo "✅ Test complete"
