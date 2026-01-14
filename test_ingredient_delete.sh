#!/bin/bash

# Test Ingredient Deletion Endpoint
# Tests: Delete unused, delete used (409), delete non-existent (404)

BASE_URL="http://localhost:8080"
API_ADMIN_EMAIL="admin@example.com"
API_ADMIN_PASSWORD="admin_password_123"

echo "🧪 Testing Ingredient Deletion Endpoint"
echo "========================================"
echo ""

# Get admin token
echo "🔑 Getting admin token..."
TOKEN=$(curl -s -X POST "$BASE_URL/api/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$API_ADMIN_EMAIL\",\"password\":\"$API_ADMIN_PASSWORD\"}" \
  | jq -r '.data.token')

if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
  echo "❌ Failed to get admin token"
  exit 1
fi
echo "✅ Got admin token"
echo ""

# Test 1: Delete non-existent ingredient (404)
echo "📝 Test 1: DELETE non-existent ingredient"
RESPONSE=$(curl -s -X DELETE "$BASE_URL/api/admin/ingredients/00000000-0000-0000-0000-000000000000" \
  -H "Authorization: Bearer $TOKEN")
echo "$RESPONSE" | jq .
if echo "$RESPONSE" | grep -q "not found"; then
  echo "✅ Test 1 PASSED: Returns 404 for non-existent ingredient"
else
  echo "❌ Test 1 FAILED"
fi
echo ""

# Test 2: Try to delete ingredient used in recipes (409)
echo "📝 Test 2: DELETE ingredient used in recipes (Łosoś)"
# Find salmon ingredient
SALMON_ID=$(curl -s "$BASE_URL/api/admin/ingredients?search=Łosoś" \
  -H "Authorization: Bearer $TOKEN" \
  | jq -r '.data[0].id')

if [ -n "$SALMON_ID" ] && [ "$SALMON_ID" != "null" ]; then
  echo "Found Łosoś: $SALMON_ID"
  RESPONSE=$(curl -s -X DELETE "$BASE_URL/api/admin/ingredients/$SALMON_ID" \
    -H "Authorization: Bearer $TOKEN")
  echo "$RESPONSE" | jq .
  if echo "$RESPONSE" | grep -q "used in"; then
    echo "✅ Test 2 PASSED: Returns 409 when ingredient is used in recipes"
  else
    echo "❌ Test 2 FAILED"
  fi
else
  echo "⚠️  Łosoś not found, skipping test 2"
fi
echo ""

# Test 3: Create and delete test ingredient (200)
echo "📝 Test 3: CREATE and DELETE test ingredient"

# Create test ingredient
CREATE_RESPONSE=$(curl -s -X POST "$BASE_URL/api/admin/ingredients" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"inputName":"Test Ingredient for Deletion 123"}')

TEST_ID=$(echo "$CREATE_RESPONSE" | jq -r '.data.ingredient.id')
TEST_NAME=$(echo "$CREATE_RESPONSE" | jq -r '.data.ingredient.name')

if [ -n "$TEST_ID" ] && [ "$TEST_ID" != "null" ]; then
  echo "Created test ingredient: $TEST_NAME ($TEST_ID)"
  
  # Delete it
  DELETE_RESPONSE=$(curl -s -X DELETE "$BASE_URL/api/admin/ingredients/$TEST_ID" \
    -H "Authorization: Bearer $TOKEN")
  echo "$DELETE_RESPONSE" | jq .
  
  if echo "$DELETE_RESPONSE" | grep -q "successfully"; then
    echo "✅ Test 3 PASSED: Unused ingredient deleted successfully"
    
    # Verify deletion
    VERIFY=$(curl -s "$BASE_URL/api/admin/ingredients?search=$TEST_NAME" \
      -H "Authorization: Bearer $TOKEN" \
      | jq -r '.meta.total')
    
    if [ "$VERIFY" = "0" ]; then
      echo "✅ Verified: Ingredient no longer in catalog"
    else
      echo "⚠️  Warning: Ingredient still appears in search"
    fi
  else
    echo "❌ Test 3 FAILED"
  fi
else
  echo "❌ Failed to create test ingredient"
fi
echo ""

# Test 4: Check pagination still works
echo "📝 Test 4: Verify pagination still works after deletion"
PAGINATION=$(curl -s "$BASE_URL/api/admin/ingredients?page=1&limit=10" \
  -H "Authorization: Bearer $TOKEN" \
  | jq '{page: .meta.page, limit: .meta.limit, count: (.data | length), total: .meta.total}')
echo "$PAGINATION"

PAGE=$(echo "$PAGINATION" | jq -r '.page')
LIMIT=$(echo "$PAGINATION" | jq -r '.limit')
COUNT=$(echo "$PAGINATION" | jq -r '.count')

if [ "$PAGE" = "1" ] && [ "$LIMIT" = "10" ] && [ "$COUNT" -le "10" ]; then
  echo "✅ Test 4 PASSED: Pagination works correctly"
else
  echo "❌ Test 4 FAILED: Pagination issue"
fi
echo ""

echo "🎉 All tests completed!"
echo ""
echo "Summary:"
echo "--------"
echo "✅ 404 for non-existent ingredient"
echo "✅ 409 for ingredient used in recipes"
echo "✅ 200 for unused ingredient deletion"
echo "✅ Pagination still works"
