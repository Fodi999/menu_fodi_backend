#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BASE_URL="http://localhost:8080"
API_KEY="test-api-key"

echo -e "${BLUE}================================${NC}"
echo -e "${BLUE}  Fridge Chat Integration Tests${NC}"
echo -e "${BLUE}================================${NC}\n"

# Function to print section
print_section() {
    echo -e "\n${YELLOW}>>> $1${NC}"
}

# Function to make request
make_request() {
    local method=$1
    local endpoint=$2
    local data=$3
    local headers=$4
    
    echo -e "${BLUE}Request: $method $endpoint${NC}"
    
    if [ -z "$data" ]; then
        curl -s -X "$method" "$BASE_URL$endpoint" \
            -H "Content-Type: application/json" \
            $headers
    else
        echo -e "${BLUE}Body: $data${NC}"
        curl -s -X "$method" "$BASE_URL$endpoint" \
            -H "Content-Type: application/json" \
            -d "$data" \
            $headers
    fi
    
    echo -e "\n"
}

# ============================================================
# TEST 1: Health Check
# ============================================================
print_section "TEST 1: Health Check"
echo "Checking if server is running..."

response=$(curl -s -X GET "$BASE_URL/health")
if echo "$response" | grep -q "ok\|healthy"; then
    echo -e "${GREEN}✓ Server is running${NC}"
    echo "Response: $response"
else
    echo -e "${RED}✗ Server is not running${NC}"
    echo "Response: $response"
    exit 1
fi

# ============================================================
# TEST 2: Login (Get JWT Token)
# ============================================================
print_section "TEST 2: Login & Get JWT Token"
echo "Attempting to login with test credentials..."

login_response=$(curl -s -X POST "$BASE_URL/api/auth/login" \
    -H "Content-Type: application/json" \
    -d '{
        "email": "test@example.com",
        "password": "password123"
    }')

echo "Login Response:"
echo "$login_response" | jq '.' 2>/dev/null || echo "$login_response"

# Extract JWT token (adjust path based on actual response structure)
JWT_TOKEN=$(echo "$login_response" | jq -r '.data.token' 2>/dev/null || echo "")

if [ -z "$JWT_TOKEN" ] || [ "$JWT_TOKEN" = "null" ]; then
    echo -e "${YELLOW}! Could not extract JWT token. Using test token.${NC}"
    JWT_TOKEN="test-jwt-token-placeholder"
else
    echo -e "${GREEN}✓ Got JWT Token: ${JWT_TOKEN:0:20}...${NC}"
fi

# ============================================================
# TEST 3: Chef Mentor - Start Recipe
# ============================================================
print_section "TEST 3: Chef Mentor - Start Recipe Creation"
echo "Starting recipe creation conversation..."

chef_response=$(curl -s -X POST "$BASE_URL/api/ai/chef-mentor" \
    -H "Content-Type: application/json" \
    -d '{
        "message": "I want to make pasta carbonara",
        "language": "en",
        "currentRecipe": null,
        "conversationHistory": []
    }')

echo "Chef Mentor Response:"
echo "$chef_response" | jq '.' 2>/dev/null || echo "$chef_response"

# Try to extract recipe data
echo "$chef_response" | jq '.recipe' 2>/dev/null > /tmp/recipe.json || echo '{"title": "Pasta Carbonara", "ingredients": []}' > /tmp/recipe.json

# ============================================================
# TEST 4: Chef Mentor - Continue Conversation
# ============================================================
print_section "TEST 4: Chef Mentor - Continue Conversation"
echo "Continuing recipe development..."

chef_response=$(curl -s -X POST "$BASE_URL/api/ai/chef-mentor" \
    -H "Content-Type: application/json" \
    -d '{
        "message": "I have eggs, bacon, and pasta. What else do I need?",
        "language": "en",
        "currentRecipe": {
            "title": "Pasta Carbonara",
            "ingredients": [
                {"name": "Pasta", "amount": 400, "unit": "g"},
                {"name": "Eggs", "amount": 3, "unit": "pcs"},
                {"name": "Bacon", "amount": 200, "unit": "g"}
            ],
            "steps": []
        },
        "conversationHistory": [
            {"role": "user", "content": "I want to make pasta carbonara"},
            {"role": "assistant", "content": "Great choice! Pasta carbonara is a classic Italian dish..."}
        ]
    }')

echo "Chef Mentor Response:"
echo "$chef_response" | jq '.' 2>/dev/null || echo "$chef_response"

# ============================================================
# TEST 5: Save Ingredients to Fridge
# ============================================================
print_section "TEST 5: Save Ingredients to Fridge"
echo "Saving recipe ingredients to fridge..."

save_response=$(curl -s -X POST "$BASE_URL/api/ai/save-ingredients" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $JWT_TOKEN" \
    -d '{
        "ingredients": [
            {"name": "Pasta", "amount": 400, "unit": "g"},
            {"name": "Eggs", "amount": 3, "unit": "pcs"},
            {"name": "Bacon", "amount": 200, "unit": "g"},
            {"name": "Parmesan Cheese", "amount": 100, "unit": "g"}
        ]
    }')

echo "Save Ingredients Response:"
echo "$save_response" | jq '.' 2>/dev/null || echo "$save_response"

# Check if successful
if echo "$save_response" | grep -q "success.*true\|successfully"; then
    echo -e "${GREEN}✓ Ingredients saved successfully${NC}"
else
    echo -e "${YELLOW}! Check if server requires authentication${NC}"
fi

# ============================================================
# TEST 6: Get Fridge Items
# ============================================================
print_section "TEST 6: Get Fridge Items"
echo "Retrieving fridge contents..."

fridge_response=$(curl -s -X GET "$BASE_URL/api/fridge/" \
    -H "Authorization: Bearer $JWT_TOKEN")

echo "Fridge Response:"
echo "$fridge_response" | jq '.' 2>/dev/null || echo "$fridge_response"

# ============================================================
# TEST 7: Get Fridge Recommendations
# ============================================================
print_section "TEST 7: Get Fridge Recommendations"
echo "Getting recipe recommendations based on fridge items..."

recommendations_response=$(curl -s -X POST "$BASE_URL/api/ai/fridge-recommendations" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $JWT_TOKEN" \
    -d '{
        "cuisine": "italian",
        "maxTime": 30
    }')

echo "Fridge Recommendations Response:"
echo "$recommendations_response" | jq '.' 2>/dev/null || echo "$recommendations_response"

# ============================================================
# TEST 8: Generate Meal Plan
# ============================================================
print_section "TEST 8: Generate Meal Plan"
echo "Generating 3-day meal plan..."

meal_plan_response=$(curl -s -X POST "$BASE_URL/api/ai/meal-plan" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $JWT_TOKEN" \
    -d '{
        "days": 3,
        "targetCalories": 2000,
        "language": "en"
    }')

echo "Meal Plan Response:"
echo "$meal_plan_response" | jq '.' 2>/dev/null || echo "$meal_plan_response"

# ============================================================
# TEST 9: Generate Recipe from Title
# ============================================================
print_section "TEST 9: Generate Recipe from Title"
echo "Generating complete recipe for 'Margherita Pizza'..."

recipe_gen_response=$(curl -s -X POST "$BASE_URL/api/ai/recipe-generator" \
    -H "Content-Type: application/json" \
    -d '{
        "title": "Margherita Pizza",
        "language": "en"
    }')

echo "Recipe Generation Response:"
echo "$recipe_gen_response" | jq '.' 2>/dev/null || echo "$recipe_gen_response"

# ============================================================
# TEST SUMMARY
# ============================================================
print_section "TEST SUMMARY"

echo -e "${BLUE}Tests Completed!${NC}\n"

echo "Endpoints tested:"
echo "  ✓ GET  /health"
echo "  ✓ POST /api/auth/login"
echo "  ✓ POST /api/ai/chef-mentor"
echo "  ✓ POST /api/ai/save-ingredients"
echo "  ✓ GET  /api/fridge/"
echo "  ✓ POST /api/ai/fridge-recommendations"
echo "  ✓ POST /api/ai/meal-plan"
echo "  ✓ POST /api/ai/recipe-generator"

echo -e "\n${YELLOW}Notes:${NC}"
echo "  - JWT Token extracted: ${JWT_TOKEN:0:30}${JWT_TOKEN:0:0:+...}"
echo "  - Server base URL: $BASE_URL"
echo "  - All responses logged above"

echo -e "\n${BLUE}To run more tests:${NC}"
echo "  1. Start the server: ./bin/server"
echo "  2. Run this script: bash TEST_MANUAL.sh"
echo "  3. Check responses in output"
