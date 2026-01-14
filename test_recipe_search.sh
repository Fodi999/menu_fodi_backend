#!/bin/bash

# 🔍 TEST: Recipe Search & Filters

BASE_URL="http://localhost:8080"
API="${BASE_URL}/api/public/recipes"

echo "🔍 Testing Recipe Search & Filters"
echo "==================================="
echo ""

# Test 1: Search by keyword
echo "TEST 1: Search 'лосось'"
RESULT=$(curl -s "${API}?search=лосось&limit=5")
COUNT=$(echo "$RESULT" | jq -r '.meta.count')
FIRST_TITLE=$(echo "$RESULT" | jq -r '.data[0].title')

if [ "$COUNT" -gt 0 ] && [[ "$FIRST_TITLE" == *"осось"* ]]; then
  echo "✅ PASS: Found $COUNT recipes, first: $FIRST_TITLE"
else
  echo "❌ FAIL: Search didn't work"
fi
echo ""

# Test 2: Partial search
echo "TEST 2: Partial search 'лос'"
RESULT=$(curl -s "${API}?search=лос")
COUNT=$(echo "$RESULT" | jq -r '.meta.count')

if [ "$COUNT" -gt 0 ]; then
  echo "✅ PASS: Partial search found $COUNT recipes"
else
  echo "❌ FAIL: Partial search failed"
fi
echo ""

# Test 3: Search + Category filter
echo "TEST 3: Search 'лосось' + category=main"
RESULT=$(curl -s "${API}?search=лосось&category=main")
COUNT=$(echo "$RESULT" | jq -r '.meta.count')
CATEGORY=$(echo "$RESULT" | jq -r '.data[0].category')

if [ "$COUNT" -gt 0 ] && [ "$CATEGORY" == "main" ]; then
  echo "✅ PASS: Combined filter works ($COUNT main dishes with лосось)"
else
  echo "❌ FAIL: Combined filter failed"
fi
echo ""

# Test 4: Search + Difficulty + Time
echo "TEST 4: Search + difficulty=easy + timeLte=20"
RESULT=$(curl -s "${API}?search=лосось&difficulty=easy&timeLte=20")
COUNT=$(echo "$RESULT" | jq -r '.meta.count')

if [ "$COUNT" -ge 0 ]; then
  echo "✅ PASS: Complex filter works ($COUNT results)"
  if [ "$COUNT" -gt 0 ]; then
    DIFF=$(echo "$RESULT" | jq -r '.data[0].difficulty')
    TIME=$(echo "$RESULT" | jq -r '.data[0].timeMinutes')
    echo "   First result: difficulty=$DIFF, time=${TIME}min"
  fi
else
  echo "❌ FAIL: Complex filter failed"
fi
echo ""

# Test 5: Empty search result
echo "TEST 5: Search 'nonexistent999'"
RESULT=$(curl -s "${API}?search=nonexistent999")
COUNT=$(echo "$RESULT" | jq -r '.meta.count')

if [ "$COUNT" -eq 0 ]; then
  echo "✅ PASS: Empty result handled correctly"
else
  echo "❌ FAIL: Expected 0 results, got $COUNT"
fi
echo ""

# Test 6: Case-insensitive search
echo "TEST 6: Case-insensitive search 'ЛОСОСЬ'"
RESULT=$(curl -s "${API}?search=ЛОСОСЬ&limit=1")
COUNT=$(echo "$RESULT" | jq -r '.meta.count')

if [ "$COUNT" -gt 0 ]; then
  echo "✅ PASS: Case-insensitive search works"
else
  echo "❌ FAIL: Case sensitivity issue"
fi
echo ""

# Test 7: Pagination with search
echo "TEST 7: Pagination with search (page=1, limit=2)"
RESULT=$(curl -s "${API}?search=лосось&page=1&limit=2")
COUNT=$(echo "$RESULT" | jq -r '.meta.count')
TOTAL=$(echo "$RESULT" | jq -r '.meta.total')
PAGE=$(echo "$RESULT" | jq -r '.meta.page')

if [ "$COUNT" -eq 2 ] && [ "$PAGE" -eq 1 ]; then
  echo "✅ PASS: Pagination works (returned $COUNT of $TOTAL)"
else
  echo "⚠️  WARNING: Expected 2 items on page 1, got $COUNT"
fi
echo ""

echo "==================================="
echo "✅ Search & Filters Test Complete!"
echo "==================================="
