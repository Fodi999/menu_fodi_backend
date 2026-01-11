#!/bin/bash

echo "🧪 Testing Recipe DTO Response"
echo "================================"

# Login
TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"admin_password_123"}' | jq -r '.data.token')

if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
  echo "❌ Login failed"
  exit 1
fi

echo "✅ Login successful"
echo ""

# Get recipes
echo "📋 Fetching recipes..."
RESPONSE=$(curl -s http://localhost:8080/api/admin/recipes \
  -H "Authorization: Bearer $TOKEN")

echo "$RESPONSE" | jq '.'
echo ""

# Check first recipe
FIRST_RECIPE=$(echo "$RESPONSE" | jq '.data[0]')
echo "🔍 First recipe:"
echo "$FIRST_RECIPE" | jq '{id, title, namePl, nameEn, nameRu, createdAt, updatedAt, category}'
echo ""

# Count
TOTAL=$(echo "$RESPONSE" | jq '.meta.total')
echo "📊 Total recipes: $TOTAL"
