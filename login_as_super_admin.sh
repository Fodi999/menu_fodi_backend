#!/bin/bash

# Login as super admin and get JWT token
# Usage: ./login_as_super_admin.sh

API_BASE="https://menu-fodi-backend.koyeb.app"

echo "🔐 Logging in as super admin..."
echo ""

RESPONSE=$(curl -s -X POST "$API_BASE/api/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "admin_password_123"
  }')

echo "$RESPONSE" | jq '.'

TOKEN=$(echo "$RESPONSE" | jq -r '.token // empty')

if [ -z "$TOKEN" ]; then
  echo ""
  echo "❌ Login failed! Check credentials."
  exit 1
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ Login successful!"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "Your JWT token:"
echo "$TOKEN"
echo ""
echo "Use it to test admin API:"
echo "./test_super_admin_visible.sh \"$TOKEN\""
echo ""

# Auto-test if requested
if [ "$1" == "--test" ]; then
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo "🧪 Running automatic test..."
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  ./test_super_admin_visible.sh "$TOKEN"
fi
