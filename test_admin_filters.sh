#!/bin/bash

# Test Admin Users API with Filters
# This script tests the new filtering functionality

BASE_URL="https://menu-fodi-backend.koyeb.app/api/admin/users"
# Замени на свой токен админа:
ADMIN_TOKEN="your_admin_token_here"

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🧪 Testing Admin Users API Filters"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Test 1: Get all users (no filters)
echo "1️⃣ Test: Get all users (default pagination)"
echo "   GET $BASE_URL?page=1&limit=20"
curl -s -H "Authorization: Bearer $ADMIN_TOKEN" \
  "$BASE_URL?page=1&limit=20" | jq '{
    total: .meta.total,
    page: .meta.page,
    limit: .meta.limit,
    totalPages: .meta.totalPages,
    returned: (.users | length)
  }'
echo ""

# Test 2: Filter by role=admin
echo "2️⃣ Test: Filter by role=admin"
echo "   GET $BASE_URL?role=admin"
curl -s -H "Authorization: Bearer $ADMIN_TOKEN" \
  "$BASE_URL?role=admin" | jq '{
    total: .meta.total,
    returned: (.users | length),
    roles: [.users[].role] | unique
  }'
echo ""

# Test 3: Filter by role=home_chef
echo "3️⃣ Test: Filter by role=home_chef"
echo "   GET $BASE_URL?role=home_chef&limit=5"
curl -s -H "Authorization: Bearer $ADMIN_TOKEN" \
  "$BASE_URL?role=home_chef&limit=5" | jq '{
    total: .meta.total,
    returned: (.users | length),
    roles: [.users[].role] | unique
  }'
echo ""

# Test 4: Filter by status=blocked
echo "4️⃣ Test: Filter by status=blocked"
echo "   GET $BASE_URL?status=blocked"
curl -s -H "Authorization: Bearer $ADMIN_TOKEN" \
  "$BASE_URL?status=blocked" | jq '{
    total: .meta.total,
    returned: (.users | length),
    statuses: [.users[].status] | unique
  }'
echo ""

# Test 5: Search by name/email
echo "5️⃣ Test: Search by email/name (search=admin)"
echo "   GET $BASE_URL?search=admin"
curl -s -H "Authorization: Bearer $ADMIN_TOKEN" \
  "$BASE_URL?search=admin" | jq '{
    total: .meta.total,
    returned: (.users | length),
    emails: [.users[].email]
  }'
echo ""

# Test 6: Combined filters
echo "6️⃣ Test: Combined filters (role=admin&status=active)"
echo "   GET $BASE_URL?role=admin&status=active"
curl -s -H "Authorization: Bearer $ADMIN_TOKEN" \
  "$BASE_URL?role=admin&status=active" | jq '{
    total: .meta.total,
    returned: (.users | length),
    roles: [.users[].role] | unique,
    statuses: [.users[].status] | unique
  }'
echo ""

# Test 7: Pagination
echo "7️⃣ Test: Pagination (page=2, limit=10)"
echo "   GET $BASE_URL?page=2&limit=10"
curl -s -H "Authorization: Bearer $ADMIN_TOKEN" \
  "$BASE_URL?page=2&limit=10" | jq '{
    meta: .meta,
    returned: (.users | length)
  }'
echo ""

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ Testing complete!"
echo ""
echo "Expected results:"
echo "  1. Total: 54 users, returned: 20 (default limit)"
echo "  2. Total: 4 admins, roles: [admin]"
echo "  3. Total: 49 home_chefs, roles: [home_chef]"
echo "  4. Total: 0 blocked, returned: 0"
echo "  5. Total: depends on search, emails contain 'admin'"
echo "  6. Total: 4 active admins"
echo "  7. Page 2, 10 items"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
