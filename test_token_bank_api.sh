#!/bin/bash

# TokenBank API Testing Script
# This script tests all token bank endpoints

set -e

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
API_URL="${API_URL:-http://localhost:8080}"
ADMIN_TOKEN="${ADMIN_TOKEN}"
TEST_USER_ID="${TEST_USER_ID}"

echo -e "${BLUE}🧪 TokenBank API Testing${NC}"
echo -e "${BLUE}========================${NC}\n"

if [ -z "$ADMIN_TOKEN" ]; then
    echo -e "${RED}❌ Error: ADMIN_TOKEN is not set${NC}"
    echo "Usage: ADMIN_TOKEN=<token> TEST_USER_ID=<user_id> $0"
    exit 1
fi

if [ -z "$TEST_USER_ID" ]; then
    echo -e "${RED}❌ Error: TEST_USER_ID is not set${NC}"
    echo "Usage: ADMIN_TOKEN=<token> TEST_USER_ID=<user_id> $0"
    exit 1
fi

test_endpoint() {
    local method=$1
    local endpoint=$2
    local data=$3
    local expected_code=$4
    
    echo -e "${YELLOW}Testing: ${method} ${endpoint}${NC}"
    
    if [ -z "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X "$method" \
            "${API_URL}${endpoint}" \
            -H "Authorization: Bearer $ADMIN_TOKEN" \
            -H "Content-Type: application/json")
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" \
            "${API_URL}${endpoint}" \
            -H "Authorization: Bearer $ADMIN_TOKEN" \
            -H "Content-Type: application/json" \
            -d "$data")
    fi
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" = "$expected_code" ]; then
        echo -e "${GREEN}✅ Status: $http_code (Expected: $expected_code)${NC}"
    else
        echo -e "${RED}❌ Status: $http_code (Expected: $expected_code)${NC}"
        echo -e "${RED}Response: $body${NC}"
        return 1
    fi
    
    echo "$body" | jq '.' 2>/dev/null || echo "$body"
    echo ""
}

echo -e "${BLUE}1. Get all token banks${NC}"
test_endpoint "GET" "/api/admin/token-bank" "" "200"

echo -e "${BLUE}2. Get token bank statistics${NC}"
test_endpoint "GET" "/api/admin/token-bank/stats" "" "200"

echo -e "${BLUE}3. Get specific user's token bank${NC}"
test_endpoint "GET" "/api/admin/token-bank/$TEST_USER_ID" "" "200"

echo -e "${BLUE}4. Allocate tokens to user${NC}"
test_endpoint "POST" "/api/admin/token-bank/allocate" \
    "{\"user_id\": \"$TEST_USER_ID\", \"amount\": 100, \"reason\": \"Test allocation\"}" \
    "200"

echo -e "${BLUE}5. Check updated token bank${NC}"
test_endpoint "GET" "/api/admin/token-bank/$TEST_USER_ID" "" "200"

echo -e "${BLUE}6. Set exact balance${NC}"
test_endpoint "PUT" "/api/admin/token-bank/balance" \
    "{\"user_id\": \"$TEST_USER_ID\", \"balance\": 500}" \
    "200"

echo -e "${BLUE}7. Revoke some tokens${NC}"
test_endpoint "POST" "/api/admin/token-bank/revoke" \
    "{\"user_id\": \"$TEST_USER_ID\", \"amount\": 50, \"reason\": \"Test revoke\"}" \
    "200"

echo -e "${BLUE}8. Check final token bank state${NC}"
test_endpoint "GET" "/api/admin/token-bank/$TEST_USER_ID" "" "200"

echo -e "${GREEN}✅ All tests completed!${NC}"
