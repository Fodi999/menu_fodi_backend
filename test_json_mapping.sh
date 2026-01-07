#!/bin/bash

# Final AI Classification Test
# Tests that all fields are properly returned in JSON response

TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiI3ZWM4YWJhNC04MTk1LTRiZTEtYTlhOC0wNjdjMzBhYWUzMDYiLCJlbWFpbCI6ImFkbWluQGV4YW1wbGUuY29tIiwicm9sZSI6InN1cGVyX2FkbWluIiwiZXhwIjoxNzY3ODY2MjUxLCJpYXQiOjE3Njc3Nzk4NTF9.5jw2OA_DZ5qJv-cSrTORbxZV_X_QK_J3VCEbjI4Fqx8"

API_BASE="http://localhost:8080"

# Unique ingredient name with timestamp
INGREDIENT="Avocado $(date +%s)"

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🧪 Testing AI Classification JSON Response"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "📝 Input: $INGREDIENT"
echo ""

RESPONSE=$(curl -s -X POST "$API_BASE/api/admin/ingredients" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"inputName\":\"$INGREDIENT\"}")

echo "$RESPONSE" | jq '.'

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ VALIDATION:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Extract data fields
NAME_EN=$(echo "$RESPONSE" | jq -r '.data.nameEn')
NAME_PL=$(echo "$RESPONSE" | jq -r '.data.namePl')
NAME_RU=$(echo "$RESPONSE" | jq -r '.data.nameRu')
NORMALIZED=$(echo "$RESPONSE" | jq -r '.data.normalizedValue')
CATEGORY=$(echo "$RESPONSE" | jq -r '.data.category')
UNIT=$(echo "$RESPONSE" | jq -r '.data.unit')

if [ "$NAME_EN" = "null" ]; then
  echo "❌ nameEn is null"
else
  echo "✅ nameEn: $NAME_EN"
fi

if [ "$NAME_PL" = "null" ]; then
  echo "❌ namePl is null"
else
  echo "✅ namePl: $NAME_PL"
fi

if [ "$NAME_RU" = "null" ]; then
  echo "❌ nameRu is null"
else
  echo "✅ nameRu: $NAME_RU"
fi

if [ "$NORMALIZED" = "null" ]; then
  echo "❌ normalizedValue is null - MAIN BUG!"
else
  echo "✅ normalizedValue: $NORMALIZED"
fi

if [ "$CATEGORY" = "null" ]; then
  echo "❌ category is null"
else
  echo "✅ category: $CATEGORY"
fi

if [ "$UNIT" = "null" ]; then
  echo "❌ unit is null"
else
  echo "✅ unit: $UNIT"
fi

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
