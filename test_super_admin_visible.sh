#!/bin/bash

# Test admin filter - check if super admin is visible
# Usage: ./test_super_admin_visible.sh YOUR_JWT_TOKEN

TOKEN="${1}"

if [ -z "$TOKEN" ]; then
  echo "❌ Error: JWT token required"
  echo "Usage: ./test_super_admin_visible.sh YOUR_JWT_TOKEN"
  exit 1
fi

API_BASE="https://menu-fodi-backend.koyeb.app"

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🔍 Testing: GET /api/admin/users?role=admin"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

RESPONSE=$(curl -s -H "Authorization: Bearer $TOKEN" \
  "$API_BASE/api/admin/users?role=admin")

echo "$RESPONSE" | jq '.'

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📊 Summary:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

TOTAL=$(echo "$RESPONSE" | jq -r '.meta.total // "N/A"')
USERS_COUNT=$(echo "$RESPONSE" | jq -r '.users | length // "N/A"')

echo "Total admins (meta.total): $TOTAL"
echo "Users returned: $USERS_COUNT"
echo ""

# Check if admin@example.com is in the list
SUPER_ADMIN=$(echo "$RESPONSE" | jq -r '.users[] | select(.email == "admin@example.com") | .name // "NOT FOUND"')

if [ "$SUPER_ADMIN" == "NOT FOUND" ]; then
  echo "❌ PROBLEM: Super admin (admin@example.com) NOT in response!"
  echo ""
  echo "🔍 Emails in response:"
  echo "$RESPONSE" | jq -r '.users[].email'
else
  echo "✅ SUCCESS: Super admin found in response!"
  echo "   Name: $SUPER_ADMIN"
  echo "   Email: admin@example.com"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🗂️  All admins in database (for comparison):"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

source .env
psql "$DATABASE_URL" -c "SELECT name, email, role FROM \"User\" WHERE role = 'admin' ORDER BY \"createdAt\" DESC;"
