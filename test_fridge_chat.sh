#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

BASE_URL="http://localhost:3000/api"

# Test user token (from your system)
TOKEN="your-jwt-token-here"

echo -e "${BLUE}=== Testing Fridge Chat Integration ===${NC}\n"

# 1. First, let's test the chef-mentor endpoint to create a recipe
echo -e "${BLUE}1. Testing Chef Mentor (Create Recipe)${NC}"
curl -X POST "$BASE_URL/ai/chef-mentor" \
  -H "Content-Type: application/json" \
  -d '{
    "message": "I want to make a pasta dish",
    "language": "en",
    "currentRecipe": null,
    "history": []
  }' | jq .

echo -e "\n${BLUE}2. Testing Save Ingredients to Fridge${NC}"
# Now save some ingredients to fridge
curl -X POST "$BASE_URL/ai/save-ingredients" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "ingredients": [
      {
        "name": "Pasta",
        "amount": 500,
        "unit": "г"
      },
      {
        "name": "Tomato Sauce",
        "amount": 400,
        "unit": "мл"
      },
      {
        "name": "Garlic",
        "amount": 3,
        "unit": "шт"
      },
      {
        "name": "Olive Oil",
        "amount": 100,
        "unit": "мл"
      }
    ]
  }' | jq .

echo -e "\n${BLUE}3. Verify Ingredients Saved to Fridge${NC}"
curl -X GET "$BASE_URL/fridge/" \
  -H "Authorization: Bearer $TOKEN" | jq .

echo -e "\n${GREEN}✓ Test completed!${NC}"
