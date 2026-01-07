#!/bin/bash

# Test ingredient suggest with Accept-Language localization
# Usage: ./test_localization.sh

TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiI3ZWM4YWJhNC04MTk1LTRiZTEtYTlhOC0wNjdjMzBhYWUzMDYiLCJlbWFpbCI6ImFkbWluQGV4YW1wbGUuY29tIiwicm9sZSI6InN1cGVyX2FkbWluIiwiZXhwIjoxNzY3ODY2MjUxLCJpYXQiOjE3Njc3Nzk4NTF9.5jw2OA_DZ5qJv-cSrTORbxZV_X_QK_J3VCEbjI4Fqx8"
BASE_URL="http://localhost:8080"

echo "🧪 Testing Accept-Language Localization"
echo "========================================"
echo ""

# Test 1: Polish (лосось → Łosoś)
echo "📋 Test 1: Accept-Language: pl"
echo "Query: 'лосось' (should return Polish name)"
echo ""
curl -s "$BASE_URL/api/admin/ingredients/suggest?q=лосось&limit=1" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Accept-Language: pl" | jq '.'
echo ""
echo "✅ Expected: name='Łosoś' (Polish)"
echo ""
echo "========================================"
echo ""

# Test 2: English (лосось → Salmon)
echo "📋 Test 2: Accept-Language: en"
echo "Query: 'лосось' (should return English name)"
echo ""
curl -s "$BASE_URL/api/admin/ingredients/suggest?q=лосось&limit=1" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Accept-Language: en" | jq '.'
echo ""
echo "✅ Expected: name='Salmon' (English)"
echo ""
echo "========================================"
echo ""

# Test 3: Russian (лосось → Лосось)
echo "📋 Test 3: Accept-Language: ru"
echo "Query: 'лосось' (should return Russian name)"
echo ""
curl -s "$BASE_URL/api/admin/ingredients/suggest?q=лосось&limit=1" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Accept-Language: ru" | jq '.'
echo ""
echo "✅ Expected: name='Лосось' (Russian)"
echo ""
echo "========================================"
echo ""

# Test 4: No header (default to English)
echo "📋 Test 4: No Accept-Language header (default)"
echo "Query: 'salmon'"
echo ""
curl -s "$BASE_URL/api/admin/ingredients/suggest?q=salmon&limit=1" \
  -H "Authorization: Bearer $TOKEN" | jq '.'
echo ""
echo "✅ Expected: name='Salmon' (English by default)"
echo ""
echo "========================================"
echo ""

# Test 5: Multi-language Accept-Language
echo "📋 Test 5: Accept-Language: pl-PL,en;q=0.9"
echo "Query: 'rice'"
echo ""
curl -s "$BASE_URL/api/admin/ingredients/suggest?q=rice&limit=1" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Accept-Language: pl-PL,en;q=0.9" | jq '.'
echo ""
echo "✅ Expected: name='Ryż' (Polish, first in list)"
echo ""
echo "========================================"

echo ""
echo "✅ All tests complete!"
echo ""
echo "📝 Summary:"
echo "  - Accept-Language: pl → Polish names (Łosoś, Ryż)"
echo "  - Accept-Language: en → English names (Salmon, Rice)"
echo "  - Accept-Language: ru → Russian names (Лосось, Рис)"
echo "  - No header → English by default"
