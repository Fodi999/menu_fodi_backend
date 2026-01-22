#!/bin/bash

# ============================================================================
# Тест Recipe Recommendation Engine (2025 Architecture)
# Rules Engine решает, AI объясняет
# ============================================================================

set -e

echo "🧪 Testing Recipe Recommendation Engine"
echo "======================================"
echo ""

# Configuration
API_URL="http://localhost:8080/api"
ADMIN_EMAIL="admin@example.com"
ADMIN_PASSWORD="securePassword123"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Step 1: Login as admin
echo "🔐 Step 1: Logging in as admin..."
LOGIN_RESPONSE=$(curl -s -X POST "${API_URL}/auth/login" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"${ADMIN_EMAIL}\",
    \"password\": \"${ADMIN_PASSWORD}\"
  }")

TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.token')

if [ "$TOKEN" == "null" ] || [ -z "$TOKEN" ]; then
  echo -e "${RED}❌ Login failed${NC}"
  echo "Response: $LOGIN_RESPONSE"
  exit 1
fi

echo -e "${GREEN}✅ Login successful${NC}"
USER_ID=$(echo $LOGIN_RESPONSE | jq -r '.user.id')
echo "User ID: $USER_ID"
echo ""

# Step 2: Check fridge
echo "📦 Step 2: Checking fridge contents..."
FRIDGE_RESPONSE=$(curl -s -X GET "${API_URL}/fridge/items?lang=ru" \
  -H "Authorization: Bearer ${TOKEN}")

FRIDGE_COUNT=$(echo $FRIDGE_RESPONSE | jq -r 'length')
echo "Fridge items count: $FRIDGE_COUNT"

if [ "$FRIDGE_COUNT" == "0" ]; then
  echo -e "${YELLOW}⚠️  Fridge is empty! Add some items first.${NC}"
  echo ""
  echo "💡 Run this to add test items:"
  echo "   ./test_fridge_units_autocomplete.sh"
  echo ""
fi

echo "Top 3 items:"
echo $FRIDGE_RESPONSE | jq -r '.[0:3] | .[] | "  - \(.name) (\(.quantity) \(.unit))"'
echo ""

# Step 3: Get Recipe Recommendations (NEW API)
echo "🎯 Step 3: Getting recipe recommendations (Rules Engine)..."
echo ""

# Test 1: Default (Polish, limit 10)
echo "Test 1: Default params (lang=pl, limit=10)"
echo "----------------------------------------"
RECOM_RESPONSE_1=$(curl -s -X GET "${API_URL}/recipes/recommendations" \
  -H "Authorization: Bearer ${TOKEN}")

echo "Response:"
echo $RECOM_RESPONSE_1 | jq '.'
echo ""

# Extract decision
DECISION=$(echo $RECOM_RESPONSE_1 | jq -r '.decision')
SUMMARY=$(echo $RECOM_RESPONSE_1 | jq -r '.summary')
TOTAL_MATCHES=$(echo $RECOM_RESPONSE_1 | jq -r '.total_matches')

echo -e "${GREEN}Decision: $DECISION${NC}"
echo "Summary: $SUMMARY"
echo "Total matches: $TOTAL_MATCHES"
echo ""

# Show top 3 recipes
if [ "$TOTAL_MATCHES" -gt "0" ]; then
  echo "Top 3 recommended recipes:"
  echo $RECOM_RESPONSE_1 | jq -r '.recipes[0:3] | .[] | 
    "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n" +
    "📌 \(.title)\n" +
    "   Match: \(.match_percent)% (\(.match_status))\n" +
    "   Available: \(.available_count)/\(.total_required)\n" +
    "   Missing: \(.missing_count)\n" +
    "   Cook time: \(.cook_time) min\n"'
fi

echo ""

# Test 2: Russian language
echo "Test 2: Russian language (lang=ru, limit=5)"
echo "----------------------------------------"
RECOM_RESPONSE_2=$(curl -s -X GET "${API_URL}/recipes/recommendations?lang=ru&limit=5" \
  -H "Authorization: Bearer ${TOKEN}")

SUMMARY_RU=$(echo $RECOM_RESPONSE_2 | jq -r '.summary')
TOTAL_MATCHES_RU=$(echo $RECOM_RESPONSE_2 | jq -r '.total_matches')

echo "Summary (RU): $SUMMARY_RU"
echo "Total matches: $TOTAL_MATCHES_RU"
echo ""

if [ "$TOTAL_MATCHES_RU" -gt "0" ]; then
  echo "Top recipe (Russian):"
  echo $RECOM_RESPONSE_2 | jq -r '.recipes[0] | 
    "  Title: \(.title)\n" +
    "  Match: \(.match_percent)%\n" +
    "  Status: \(.match_status)\n"'
fi

echo ""

# Test 3: English language
echo "Test 3: English language (lang=en, limit=3)"
echo "----------------------------------------"
RECOM_RESPONSE_3=$(curl -s -X GET "${API_URL}/recipes/recommendations?lang=en&limit=3" \
  -H "Authorization: Bearer ${TOKEN}")

SUMMARY_EN=$(echo $RECOM_RESPONSE_3 | jq -r '.summary')
TOTAL_MATCHES_EN=$(echo $RECOM_RESPONSE_3 | jq -r '.total_matches')

echo "Summary (EN): $SUMMARY_EN"
echo "Total matches: $TOTAL_MATCHES_EN"
echo ""

# Step 4: Performance test
echo "⏱️  Step 4: Performance test (10 requests)..."
echo "----------------------------------------"

start_time=$(date +%s%N)

for i in {1..10}; do
  curl -s -X GET "${API_URL}/recipes/recommendations?limit=5" \
    -H "Authorization: Bearer ${TOKEN}" > /dev/null
done

end_time=$(date +%s%N)
elapsed_ms=$(( (end_time - start_time) / 1000000 ))
avg_ms=$(( elapsed_ms / 10 ))

echo "Total time: ${elapsed_ms}ms"
echo "Average per request: ${avg_ms}ms"
echo ""

if [ "$avg_ms" -lt "200" ]; then
  echo -e "${GREEN}✅ Performance: EXCELLENT (<200ms)${NC}"
elif [ "$avg_ms" -lt "500" ]; then
  echo -e "${YELLOW}⚠️  Performance: GOOD (200-500ms)${NC}"
else
  echo -e "${RED}❌ Performance: SLOW (>500ms)${NC}"
fi

echo ""
echo "======================================"
echo -e "${GREEN}✅ Tests completed!${NC}"
echo "======================================"
echo ""
echo "📊 Summary:"
echo "  - Decision: $DECISION"
echo "  - Total matches: $TOTAL_MATCHES"
echo "  - Average response time: ${avg_ms}ms"
echo "  - Architecture: Rules Engine (NO AI)"
echo ""
