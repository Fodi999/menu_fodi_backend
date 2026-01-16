#!/bin/bash

# JWT Sub Field - Quick Test Script
# Tests that new JWT tokens contain the 'sub' field

echo "🧪 Testing JWT Sub Field Fix..."
echo ""

# Test login endpoint
echo "📡 Testing POST /api/auth/login..."
RESPONSE=$(curl -s -X POST https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"fodi85@gmail.ru","password":"210185"}')

echo "$RESPONSE" | jq .

# Extract token
TOKEN=$(echo "$RESPONSE" | jq -r '.data.token')

if [ "$TOKEN" = "null" ] || [ -z "$TOKEN" ]; then
  echo "❌ Failed to get token"
  exit 1
fi

echo ""
echo "✅ Token received, length: ${#TOKEN}"
echo ""

# Decode JWT payload (middle part between dots)
PAYLOAD=$(echo "$TOKEN" | cut -d. -f2)

# Add padding if needed for base64
case $((${#PAYLOAD} % 4)) in
  2) PAYLOAD="${PAYLOAD}==" ;;
  3) PAYLOAD="${PAYLOAD}=" ;;
esac

echo "🔍 Decoding JWT payload..."
echo ""

# Decode with node (more reliable than base64 command)
if command -v node &> /dev/null; then
  DECODED=$(node -e "console.log(JSON.stringify(JSON.parse(Buffer.from('$PAYLOAD', 'base64').toString()), null, 2))")
  echo "$DECODED"
  
  # Check for 'sub' field
  SUB=$(echo "$DECODED" | jq -r '.sub')
  
  if [ "$SUB" = "null" ] || [ -z "$SUB" ]; then
    echo ""
    echo "❌ FAILED: 'sub' field is missing!"
    exit 1
  else
    echo ""
    echo "✅ SUCCESS: 'sub' field present: $SUB"
    echo ""
    echo "📋 Token structure:"
    echo "   - email: $(echo "$DECODED" | jq -r '.email')"
    echo "   - role: $(echo "$DECODED" | jq -r '.role')"
    echo "   - sub: $SUB"
    echo "   - hasRole: $(echo "$DECODED" | jq -r '.hasRole')"
    echo ""
    echo "🎉 JWT Sub Field Fix: VALIDATED ✅"
  fi
else
  echo "⚠️  Node.js not found, using base64 decode"
  echo "$PAYLOAD" | base64 -d 2>/dev/null | jq .
fi
