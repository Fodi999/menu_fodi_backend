#!/bin/bash
# ШАГ 1: Тест минимального payload для POST /api/admin/recipes

BASE_URL="http://localhost:8080"
# BASE_URL="https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app"

echo "=== 🔐 Login as admin ==="
LOGIN_RESPONSE=$(curl -s -X POST $BASE_URL/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"admin_password_123"}')

TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.token')

if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
  echo "❌ Login failed"
  echo "Response: $LOGIN_RESPONSE"
  exit 1
fi

echo "✅ Token received: ${TOKEN:0:30}..."

echo -e "\n=== 📝 Create MINIMAL draft (канонический payload) ==="
DRAFT_RESPONSE=$(curl -s -X POST $BASE_URL/api/admin/recipes \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "localName": "Pierogi ruskie",
    "canonicalName": "pierogi-ruskie",
    "description": "Черновик рецепта",
    "category": "main",
    "difficulty": "easy"
  }')

echo "Response:"
echo $DRAFT_RESPONSE | jq '.'

# Check status
STATUS=$(echo $DRAFT_RESPONSE | jq -r '.data.status')
RECIPE_ID=$(echo $DRAFT_RESPONSE | jq -r '.data.id')

if [ "$STATUS" = "draft" ]; then
  echo -e "\n✅ SUCCESS! Draft created with ID: $RECIPE_ID"
  echo "✅ Status: $STATUS"
  echo "✅ LocalName: $(echo $DRAFT_RESPONSE | jq -r '.data.localName')"
  echo "✅ Category: $(echo $DRAFT_RESPONSE | jq -r '.data.category')"
  echo "✅ Difficulty: $(echo $DRAFT_RESPONSE | jq -r '.data.difficulty')"
else
  echo -e "\n❌ FAILED! Expected status=draft, got: $STATUS"
  exit 1
fi

echo -e "\n=== 🔍 Verify in database ==="
echo "Run this SQL to verify:"
echo "SELECT id, \"localName\", status, source, category, difficulty, \"authorId\" FROM \"Recipe\" WHERE id = '$RECIPE_ID';"
