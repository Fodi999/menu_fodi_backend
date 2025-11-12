#!/bin/bash

# User Endpoints Testing Script
# Koyeb: https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app

set -e

API_URL="https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app"

echo "🧪 Testing User Endpoints on Koyeb"
echo "API: $API_URL"
echo ""

# 1. Register a new user
echo "1️⃣ Registering new user..."
REGISTER_RESPONSE=$(curl -s -X POST "$API_URL/api/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "email":"testuser'$(date +%s)'@example.com",
    "password":"Password123!",
    "name":"Test User"
  }')

TOKEN=$(echo "$REGISTER_RESPONSE" | jq -r '.data.token')
USER_ID=$(echo "$REGISTER_RESPONSE" | jq -r '.data.user.id')

echo "✅ User registered"
echo "   User ID: $USER_ID"
echo "   Token: ${TOKEN:0:30}..."
echo ""

# 2. Get Profile
echo "2️⃣ Testing GET /api/user/profile"
curl -s -X GET "$API_URL/api/user/profile" \
  -H "Authorization: Bearer $TOKEN" | jq .
echo ""

# 3. Get Progress
echo "3️⃣ Testing GET /api/user/progress"
curl -s -X GET "$API_URL/api/user/progress" \
  -H "Authorization: Bearer $TOKEN" | jq .
echo ""

# 4. Get Dashboard (FIXED)
echo "4️⃣ Testing GET /api/user/dashboard (FIXED)"
curl -s -X GET "$API_URL/api/user/dashboard" \
  -H "Authorization: Bearer $TOKEN" | jq .
echo ""

# 5. Get Achievements (FIXED)
echo "5️⃣ Testing GET /api/user/achievements (FIXED)"
curl -s -X GET "$API_URL/api/user/achievements" \
  -H "Authorization: Bearer $TOKEN" | jq .
echo ""

# 6. Get Wallet (NEW)
echo "6️⃣ Testing GET /api/user/wallet (NEW)"
curl -s -X GET "$API_URL/api/user/wallet" \
  -H "Authorization: Bearer $TOKEN" | jq .
echo ""

# 7. Test Admin protection
echo "7️⃣ Testing Admin protection (should fail with user token)"
curl -s -X GET "$API_URL/api/admin/profile" \
  -H "Authorization: Bearer $TOKEN" | jq .
echo ""

echo "✅ All tests completed!"
