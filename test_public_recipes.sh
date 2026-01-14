#!/bin/bash

# 🎯 TEST: SEO-Ready Public Recipe Endpoints
# Tests public catalog and single recipe endpoints without authentication

BASE_URL="http://localhost:8080"
PUBLIC_API="${BASE_URL}/api/public/recipes"

echo "🚀 Testing SEO-Ready Public Endpoints"
echo "======================================"
echo ""

# Test 1: Public Catalog (no auth)
echo "📋 TEST 1: Public Recipe Catalog (no auth required)"
echo "GET ${PUBLIC_API}"
CATALOG_RESPONSE=$(curl -s "${PUBLIC_API}")
CATALOG_COUNT=$(echo "$CATALOG_RESPONSE" | jq -r '.meta.count // 0')

if [ "$CATALOG_COUNT" -gt 0 ]; then
  echo "✅ PASS: Got $CATALOG_COUNT recipes"
  echo "Meta: $(echo "$CATALOG_RESPONSE" | jq -c '.meta')"
else
  echo "❌ FAIL: No recipes returned"
  echo "Response: $CATALOG_RESPONSE"
fi
echo ""

# Test 2: Public Catalog with Filters
echo "📋 TEST 2: Public Catalog with Category Filter"
echo "GET ${PUBLIC_API}?category=main"
FILTERED_RESPONSE=$(curl -s "${PUBLIC_API}?category=main")
FILTERED_COUNT=$(echo "$FILTERED_RESPONSE" | jq -r '.meta.count // 0')

if [ "$FILTERED_COUNT" -gt 0 ]; then
  echo "✅ PASS: Got $FILTERED_COUNT main dishes"
else
  echo "❌ FAIL: Filter didn't work"
fi
echo ""

# Test 3: Pagination
echo "📋 TEST 3: Pagination (limit=5)"
echo "GET ${PUBLIC_API}?limit=5"
PAGINATED_RESPONSE=$(curl -s "${PUBLIC_API}?limit=5")
PAGINATED_COUNT=$(echo "$PAGINATED_RESPONSE" | jq -r '.meta.count // 0')

if [ "$PAGINATED_COUNT" -eq 5 ]; then
  echo "✅ PASS: Got exactly 5 recipes"
else
  echo "⚠️ WARNING: Expected 5, got $PAGINATED_COUNT"
fi
echo ""

# Test 4: Get first recipe canonical name
echo "🔍 TEST 4: Getting first recipe's canonical name..."
FIRST_CANONICAL=$(echo "$CATALOG_RESPONSE" | jq -r '.data[0].canonicalName // ""')

if [ -n "$FIRST_CANONICAL" ]; then
  echo "✅ Found recipe: $FIRST_CANONICAL"
  
  # Test 5: Single Recipe by Slug
  echo ""
  echo "📖 TEST 5: Get Recipe by Slug (SEO URL)"
  echo "GET ${PUBLIC_API}/${FIRST_CANONICAL}"
  RECIPE_RESPONSE=$(curl -s "${PUBLIC_API}/${FIRST_CANONICAL}")
  RECIPE_TITLE=$(echo "$RECIPE_RESPONSE" | jq -r '.data.title // ""')
  
  if [ -n "$RECIPE_TITLE" ]; then
    echo "✅ PASS: Recipe found"
    echo "Title: $RECIPE_TITLE"
    echo "Canonical: $(echo "$RECIPE_RESPONSE" | jq -r '.data.canonicalName')"
    echo "Category: $(echo "$RECIPE_RESPONSE" | jq -r '.data.category')"
    echo "Difficulty: $(echo "$RECIPE_RESPONSE" | jq -r '.data.difficulty')"
    echo "Time: $(echo "$RECIPE_RESPONSE" | jq -r '.data.timeMinutes')min"
  else
    echo "❌ FAIL: Recipe not found"
    echo "Response: $RECIPE_RESPONSE"
  fi
else
  echo "❌ FAIL: No recipes in catalog"
fi
echo ""

# Test 6: Non-existent Recipe (404 handling)
echo "📖 TEST 6: Non-existent Recipe (should return 404)"
echo "GET ${PUBLIC_API}/nonexistent-recipe-xyz"
NOT_FOUND_RESPONSE=$(curl -s -w "\nHTTP_CODE:%{http_code}" "${PUBLIC_API}/nonexistent-recipe-xyz")
NOT_FOUND_CODE=$(echo "$NOT_FOUND_RESPONSE" | grep "HTTP_CODE" | cut -d':' -f2)

if [ "$NOT_FOUND_CODE" == "404" ]; then
  echo "✅ PASS: Correctly returns 404"
else
  echo "❌ FAIL: Expected 404, got $NOT_FOUND_CODE"
fi
echo ""

# Test 7: Public DTO (no internal fields)
echo "🔒 TEST 7: Public DTO (should not expose internal fields)"
FIRST_RECIPE=$(echo "$CATALOG_RESPONSE" | jq -r '.data[0]')
HAS_AUTHOR=$(echo "$FIRST_RECIPE" | jq 'has("authorId")')
HAS_SOURCE=$(echo "$FIRST_RECIPE" | jq 'has("source")')
HAS_UPDATED=$(echo "$FIRST_RECIPE" | jq 'has("updatedAt")')

if [ "$HAS_AUTHOR" == "false" ] && [ "$HAS_SOURCE" == "false" ]; then
  echo "✅ PASS: Internal fields hidden (authorId, source)"
else
  echo "⚠️ WARNING: Internal fields exposed"
  echo "Has authorId: $HAS_AUTHOR"
  echo "Has source: $HAS_SOURCE"
fi
echo ""

# Test 8: Multilingual Fields
echo "🌍 TEST 8: Multilingual Fields Present"
HAS_NAME_PL=$(echo "$FIRST_RECIPE" | jq 'has("namePl")')
HAS_NAME_EN=$(echo "$FIRST_RECIPE" | jq 'has("nameEn")')
HAS_NAME_RU=$(echo "$FIRST_RECIPE" | jq 'has("nameRu")')

if [ "$HAS_NAME_PL" == "true" ] && [ "$HAS_NAME_EN" == "true" ] && [ "$HAS_NAME_RU" == "true" ]; then
  echo "✅ PASS: All language fields present"
  echo "PL: $(echo "$FIRST_RECIPE" | jq -r '.namePl // "N/A"')"
  echo "EN: $(echo "$FIRST_RECIPE" | jq -r '.nameEn // "N/A"')"
  echo "RU: $(echo "$FIRST_RECIPE" | jq -r '.nameRu // "N/A"')"
else
  echo "⚠️ WARNING: Some language fields missing"
fi
echo ""

# Test 9: Cache Headers (SEO optimization)
echo "⚡ TEST 9: Cache-Control Headers (for SEO)"
CACHE_HEADER=$(curl -s -I "${PUBLIC_API}" | grep -i "cache-control")

if [ -n "$CACHE_HEADER" ]; then
  echo "✅ PASS: Cache headers present"
  echo "$CACHE_HEADER"
else
  echo "⚠️ WARNING: No cache headers found"
fi
echo ""

# Test 10: Response Time (performance)
echo "⏱️ TEST 10: Response Time Test"
TIME_START=$(date +%s%3N)
curl -s "${PUBLIC_API}?limit=20" > /dev/null
TIME_END=$(date +%s%3N)
TIME_DIFF=$((TIME_END - TIME_START))

if [ $TIME_DIFF -lt 300 ]; then
  echo "✅ PASS: Response time: ${TIME_DIFF}ms (fast)"
elif [ $TIME_DIFF -lt 1000 ]; then
  echo "⚠️ ACCEPTABLE: Response time: ${TIME_DIFF}ms"
else
  echo "❌ SLOW: Response time: ${TIME_DIFF}ms (needs optimization)"
fi
echo ""

echo "======================================"
echo "✅ Public SEO Endpoints Test Complete!"
echo "======================================"
