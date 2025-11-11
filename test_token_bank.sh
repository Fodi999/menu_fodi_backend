#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Token Bank Admin Panel - Integration Tests${NC}\n"

# Test base URL (change if needed)
BASE_URL="http://localhost:8080"
ADMIN_TOKEN="your_admin_token_here"  # Replace with actual token

# Test 1: Check if token bank migration works
echo -e "${YELLOW}Test 1: Checking token_bank table structure...${NC}"
echo "Command: SELECT column_name, data_type FROM information_schema.columns WHERE table_name = 'token_bank';"
echo "Expected: Columns: id, user_id, balance, total_allocated, total_used, created_at, updated_at"
echo ""

# Test 2: Get all token banks
echo -e "${YELLOW}Test 2: Get all token banks${NC}"
echo "Command:"
echo "curl -X GET $BASE_URL/api/admin/token-bank \\"
echo "  -H \"Authorization: Bearer \$ADMIN_TOKEN\" \\"
echo "  -H \"Content-Type: application/json\""
echo ""

# Test 3: Get token bank stats
echo -e "${YELLOW}Test 3: Get token bank statistics${NC}"
echo "Command:"
echo "curl -X GET $BASE_URL/api/admin/token-bank/stats \\"
echo "  -H \"Authorization: Bearer \$ADMIN_TOKEN\""
echo ""

# Test 4: Get specific user's token bank
echo -e "${YELLOW}Test 4: Get specific user's token bank${NC}"
echo "Command:"
echo "curl -X GET $BASE_URL/api/admin/token-bank/{userID} \\"
echo "  -H \"Authorization: Bearer \$ADMIN_TOKEN\""
echo ""

# Test 5: Allocate tokens to a user
echo -e "${YELLOW}Test 5: Allocate tokens to user${NC}"
echo "Command:"
echo "curl -X POST $BASE_URL/api/admin/token-bank/allocate \\"
echo "  -H \"Authorization: Bearer \$ADMIN_TOKEN\" \\"
echo "  -H \"Content-Type: application/json\" \\"
echo "  -d '{"
echo "    \"user_id\": \"user-uuid-here\","
echo "    \"amount\": 500,"
echo "    \"reason\": \"Monthly allocation\""
echo "  }'"
echo ""

# Test 6: Revoke tokens from a user
echo -e "${YELLOW}Test 6: Revoke tokens from user${NC}"
echo "Command:"
echo "curl -X POST $BASE_URL/api/admin/token-bank/revoke \\"
echo "  -H \"Authorization: Bearer \$ADMIN_TOKEN\" \\"
echo "  -H \"Content-Type: application/json\" \\"
echo "  -d '{"
echo "    \"user_id\": \"user-uuid-here\","
echo "    \"amount\": 100,"
echo "    \"reason\": \"Usage adjustment\""
echo "  }'"
echo ""

# Test 7: Set exact token balance
echo -e "${YELLOW}Test 7: Set exact token balance${NC}"
echo "Command:"
echo "curl -X PUT $BASE_URL/api/admin/token-bank/balance \\"
echo "  -H \"Authorization: Bearer \$ADMIN_TOKEN\" \\"
echo "  -H \"Content-Type: application/json\" \\"
echo "  -d '{"
echo "    \"user_id\": \"user-uuid-here\","
echo "    \"balance\": 1000"
echo "  }'"
echo ""

# Test 8: Verify RBAC - User should get 403 on token bank endpoints
echo -e "${YELLOW}Test 8: Verify RBAC - Regular user should get 403${NC}"
echo "Command:"
echo "curl -X GET $BASE_URL/api/admin/token-bank \\"
echo "  -H \"Authorization: Bearer \$USER_TOKEN\""
echo "Expected: 403 Forbidden"
echo ""

# Test 9: Verify missing token returns 401
echo -e "${YELLOW}Test 9: Verify missing token returns 401${NC}"
echo "Command:"
echo "curl -X GET $BASE_URL/api/admin/token-bank"
echo "Expected: 401 Unauthorized"
echo ""

# Test 10: Integration test - Full workflow
echo -e "${YELLOW}Test 10: Full workflow integration test${NC}"
echo "Steps:"
echo "1. Create a test user via registration endpoint"
echo "2. Get token bank for user (should be 0)"
echo "3. Allocate 1000 tokens to user"
echo "4. Get token bank stats (should include new user)"
echo "5. Revoke 200 tokens from user"
echo "6. Set balance to 500 tokens"
echo "7. Get token bank for user (should be 500)"
echo ""

echo -e "${GREEN}All test scenarios configured!${NC}"
echo "To run actual tests, replace \$ADMIN_TOKEN and \$USER_TOKEN with real tokens"
echo "and make sure the server is running on $BASE_URL"
