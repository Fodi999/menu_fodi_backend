#!/bin/bash

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

KOYEB_URL="https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app"

echo -e "${YELLOW}🔐 Koyeb JWT Authentication Test${NC}\n"

# Шаг 1: Логин
echo -e "${YELLOW}Step 1: Login to get fresh JWT token${NC}"
read -p "Enter email: " EMAIL
read -sp "Enter password: " PASSWORD
echo ""

LOGIN_RESPONSE=$(curl -s -X POST \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}" \
  "$KOYEB_URL/api/auth/login")

TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"token":"[^"]*' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
  echo -e "${RED}❌ Login failed!${NC}"
  echo "Response: $LOGIN_RESPONSE"
  exit 1
fi

echo -e "${GREEN}✅ Login successful!${NC}"
echo "Token length: ${#TOKEN} characters"
echo ""

# Шаг 2: Проверка protected endpoint
echo -e "${YELLOW}Step 2: Testing protected endpoint /api/user/recipes/saved${NC}"
SAVED_RESPONSE=$(curl -i -s \
  -H "Authorization: Bearer $TOKEN" \
  "$KOYEB_URL/api/user/recipes/saved")

HTTP_STATUS=$(echo "$SAVED_RESPONSE" | grep "HTTP/" | awk '{print $2}')

echo "HTTP Status: $HTTP_STATUS"
echo ""

if [ "$HTTP_STATUS" == "200" ]; then
  echo -e "${GREEN}✅ SUCCESS! Protected endpoint works!${NC}"
  echo "Response body:"
  echo "$SAVED_RESPONSE" | grep -A 50 "^\[" || echo "$SAVED_RESPONSE" | grep -A 50 "^{"
elif [ "$HTTP_STATUS" == "401" ]; then
  echo -e "${RED}❌ STILL 401 Unauthorized!${NC}"
  echo -e "${YELLOW}Possible issues:${NC}"
  echo "1. JWT_SECRET not set correctly on Koyeb"
  echo "2. Different secrets used for login vs validation"
  echo "3. Claims validation (issuer/audience) mismatch"
  echo "4. Service not redeployed after env var change"
  echo ""
  echo "Full response:"
  echo "$SAVED_RESPONSE"
else
  echo -e "${YELLOW}⚠️  Unexpected status: $HTTP_STATUS${NC}"
  echo "Full response:"
  echo "$SAVED_RESPONSE"
fi

echo ""
echo -e "${YELLOW}Step 3: Testing public endpoint /api/public/treasury${NC}"
TREASURY_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" "$KOYEB_URL/api/public/treasury")
TREASURY_STATUS=$(echo "$TREASURY_RESPONSE" | grep "HTTP_STATUS" | cut -d':' -f2)

if [ "$TREASURY_STATUS" == "200" ]; then
  echo -e "${GREEN}✅ Public treasury endpoint works!${NC}"
else
  echo -e "${RED}❌ Treasury endpoint failed with status: $TREASURY_STATUS${NC}"
fi

echo ""
echo -e "${YELLOW}📊 Test Summary:${NC}"
echo "Login: $([ -z "$TOKEN" ] && echo '❌ Failed' || echo '✅ Success')"
echo "Protected endpoint: $([ "$HTTP_STATUS" == "200" ] && echo '✅ Working' || echo '❌ Failed')"
echo "Public endpoint: $([ "$TREASURY_STATUS" == "200" ] && echo '✅ Working' || echo '❌ Failed')"
