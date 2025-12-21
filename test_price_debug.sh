#!/bin/bash

# Test recipe generation with debug logging for prices
# Usage: ./test_price_debug.sh

API_URL="https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app"

# Use existing test user (has products with prices in fridge)
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiI5MjRhMjc4ZC1iZGI2LTRhZDUtOTE2Ny1kOWU5NDk1ZjZkYWIiLCJlbWFpbCI6ImZvZGk5OTlAZ21haWwuY29tIiwicm9sZSI6ImhvbWVfY2hlZiIsImV4cCI6MTc2NjQxMDk5NSwiaWF0IjoxNzY2MzI0NTk1fQ.MXzHdlXNLv2tFyVRmF_KaPFvVJT4M-jzGq04iKe2TlA"

echo "=== Testing Recipe Generation with Price Debug Logging ==="
echo ""
echo "User: fodi999@gmail.com (should have products with prices)"
echo "API: $API_URL"
echo ""

echo "--- Generating recipe from fridge ---"
RESPONSE=$(curl -s -X POST "$API_URL/api/ai/create-recipe-from-fridge" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"language": "pl"}')

echo "$RESPONSE" | jq '.'

echo ""
echo "--- Economy Data ---"
echo "$RESPONSE" | jq '.data.recipe.economy'

echo ""
echo "--- Used Products (first 3) ---"
echo "$RESPONSE" | jq '.data.usedProducts[:3]'

echo ""
echo "⚠️  Check Koyeb logs for debug output:"
echo "   - 'Loaded fridge items with prices'"
echo "   - 'Fridge item price' (for each product)"
echo "   - 'Price data found for item' OR 'No price data for item'"
echo "   - '[AI][ECONOMY] Used cost:' (final calculation)"
