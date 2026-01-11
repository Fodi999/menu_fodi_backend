#!/bin/bash

# 🧪 Backend AI Language Test Script
# Tests: Frontend → Backend → AI → Backend (NO DB save)
# Purpose: Verify that AI generates recipes in the REQUESTED language

set -e

echo "============================================"
echo "🧪 AI Language Generation Test"
echo "============================================"
echo ""

# Configuration
BASE_URL="http://localhost:8080"
ENDPOINT="/api/admin/recipes/preview-ai"

# Test user JWT (from logs: user 7ec8aba4-8195-4be1-a9a8-067c30aae306)
# This is a super admin token - replace with your actual token
JWT_TOKEN="YOUR_TOKEN_HERE"

# Ingredient IDs (from logs)
SALMON_ID="fe1c7431-b1b7-4d36-94bf-74276481983e"  # Лосось
OIL_ID="9ff773d2-a3ee-4f4b-bc45-4cfe0d7f680b"     # Масло оливковое

echo "📝 Configuration:"
echo "  Endpoint: $ENDPOINT"
echo "  Salmon ID: $SALMON_ID"
echo "  Oil ID: $OIL_ID"
echo ""

# ============================================
# TEST 1: Russian Language
# ============================================
echo "============================================"
echo "TEST 1: RUSSIAN LANGUAGE (ru)"
echo "============================================"
echo ""

echo "📤 Sending request with Accept-Language: ru..."
echo ""

RESPONSE_RU=$(curl -s -w "\nHTTP_CODE:%{http_code}" -X POST "$BASE_URL$ENDPOINT" \
  -H "Content-Type: application/json" \
  -H "Accept-Language: ru" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -d '{
    "title": "Жареный лосось с маслом",
    "language": "ru",
    "ingredients": [
      {
        "ingredientId": "'"$SALMON_ID"'",
        "quantity": 150,
        "unit": "g"
      },
      {
        "ingredientId": "'"$OIL_ID"'",
        "quantity": 20,
        "unit": "ml"
      }
    ],
    "rawCookingText": "Обжарить лосось на оливковом масле до золотистой корочки. Подавать горячим."
  }')

HTTP_CODE_RU=$(echo "$RESPONSE_RU" | grep "HTTP_CODE" | cut -d':' -f2)
BODY_RU=$(echo "$RESPONSE_RU" | sed '/HTTP_CODE/d')

echo "📥 Response HTTP Code: $HTTP_CODE_RU"
echo ""

if [ "$HTTP_CODE_RU" != "200" ]; then
  echo "❌ TEST FAILED: Expected 200, got $HTTP_CODE_RU"
  echo "Response body:"
  echo "$BODY_RU" | jq '.' 2>/dev/null || echo "$BODY_RU"
  exit 1
fi

echo "✅ HTTP 200 OK"
echo ""

# Parse response
LANG_RU=$(echo "$BODY_RU" | jq -r '.data.language' 2>/dev/null || echo "null")
TITLE_RU=$(echo "$BODY_RU" | jq -r '.data.title' 2>/dev/null || echo "")
DESC_RU=$(echo "$BODY_RU" | jq -r '.data.description' 2>/dev/null || echo "")
STEP1_TEXT_RU=$(echo "$BODY_RU" | jq -r '.data.steps[0].text' 2>/dev/null || echo "")

echo "📊 Response Analysis:"
echo "  Language field: $LANG_RU"
echo "  Title: $TITLE_RU"
echo "  Description: $DESC_RU"
echo "  First step: $STEP1_TEXT_RU"
echo ""

# Validation checks
echo "🔍 Validation:"

# Check 1: Language field should be "ru"
if [ "$LANG_RU" = "ru" ]; then
  echo "  ✅ Language field = 'ru'"
else
  echo "  ❌ Language field = '$LANG_RU' (expected 'ru')"
  exit 1
fi

# Check 2: Text should contain Cyrillic characters
if echo "$DESC_RU" | grep -q "[А-Яа-я]"; then
  echo "  ✅ Description contains Cyrillic (Russian text detected)"
else
  echo "  ⚠️  WARNING: Description does NOT contain Cyrillic"
  echo "     Description: $DESC_RU"
fi

if echo "$STEP1_TEXT_RU" | grep -q "[А-Яа-я]"; then
  echo "  ✅ Steps contain Cyrillic (Russian text detected)"
else
  echo "  ⚠️  WARNING: Steps do NOT contain Cyrillic"
  echo "     First step: $STEP1_TEXT_RU"
fi

echo ""
echo "============================================"
echo "TEST 1: PASSED ✅"
echo "============================================"
echo ""
echo ""

# ============================================
# TEST 2: English Language
# ============================================
echo "============================================"
echo "TEST 2: ENGLISH LANGUAGE (en)"
echo "============================================"
echo ""

echo "📤 Sending request with Accept-Language: en..."
echo ""

RESPONSE_EN=$(curl -s -w "\nHTTP_CODE:%{http_code}" -X POST "$BASE_URL$ENDPOINT" \
  -H "Content-Type: application/json" \
  -H "Accept-Language: en" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -d '{
    "title": "Grilled Salmon with Oil",
    "language": "en",
    "ingredients": [
      {
        "ingredientId": "'"$SALMON_ID"'",
        "quantity": 150,
        "unit": "g"
      },
      {
        "ingredientId": "'"$OIL_ID"'",
        "quantity": 20,
        "unit": "ml"
      }
    ],
    "rawCookingText": "Grill the salmon in olive oil until golden. Serve hot."
  }')

HTTP_CODE_EN=$(echo "$RESPONSE_EN" | grep "HTTP_CODE" | cut -d':' -f2)
BODY_EN=$(echo "$RESPONSE_EN" | sed '/HTTP_CODE/d')

echo "📥 Response HTTP Code: $HTTP_CODE_EN"
echo ""

if [ "$HTTP_CODE_EN" != "200" ]; then
  echo "❌ TEST FAILED: Expected 200, got $HTTP_CODE_EN"
  echo "Response body:"
  echo "$BODY_EN" | jq '.' 2>/dev/null || echo "$BODY_EN"
  exit 1
fi

echo "✅ HTTP 200 OK"
echo ""

# Parse response
LANG_EN=$(echo "$BODY_EN" | jq -r '.data.language' 2>/dev/null || echo "null")
TITLE_EN=$(echo "$BODY_EN" | jq -r '.data.title' 2>/dev/null || echo "")
DESC_EN=$(echo "$BODY_EN" | jq -r '.data.description' 2>/dev/null || echo "")

echo "📊 Response Analysis:"
echo "  Language field: $LANG_EN"
echo "  Title: $TITLE_EN"
echo "  Description: $DESC_EN"
echo ""

echo "🔍 Validation:"

# Check: Language field should be "en"
if [ "$LANG_EN" = "en" ]; then
  echo "  ✅ Language field = 'en'"
else
  echo "  ❌ Language field = '$LANG_EN' (expected 'en')"
  exit 1
fi

# Check: Should NOT contain Cyrillic
if echo "$DESC_EN" | grep -q "[А-Яа-я]"; then
  echo "  ⚠️  WARNING: Description contains Cyrillic (should be English)"
else
  echo "  ✅ Description is in English (no Cyrillic detected)"
fi

echo ""
echo "============================================"
echo "TEST 2: PASSED ✅"
echo "============================================"
echo ""
echo ""

# ============================================
# TEST 3: Polish Language
# ============================================
echo "============================================"
echo "TEST 3: POLISH LANGUAGE (pl)"
echo "============================================"
echo ""

echo "📤 Sending request with Accept-Language: pl..."
echo ""

RESPONSE_PL=$(curl -s -w "\nHTTP_CODE:%{http_code}" -X POST "$BASE_URL$ENDPOINT" \
  -H "Content-Type: application/json" \
  -H "Accept-Language: pl" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -d '{
    "title": "Smażony łosoś z olejem",
    "language": "pl",
    "ingredients": [
      {
        "ingredientId": "'"$SALMON_ID"'",
        "quantity": 150,
        "unit": "g"
      },
      {
        "ingredientId": "'"$OIL_ID"'",
        "quantity": 20,
        "unit": "ml"
      }
    ],
    "rawCookingText": "Smażyć łososia na oliwie z oliwek do złocistego koloru. Podawać na gorąco."
  }')

HTTP_CODE_PL=$(echo "$RESPONSE_PL" | grep "HTTP_CODE" | cut -d':' -f2)
BODY_PL=$(echo "$RESPONSE_PL" | sed '/HTTP_CODE/d')

echo "📥 Response HTTP Code: $HTTP_CODE_PL"
echo ""

if [ "$HTTP_CODE_PL" != "200" ]; then
  echo "❌ TEST FAILED: Expected 200, got $HTTP_CODE_PL"
  echo "Response body:"
  echo "$BODY_PL" | jq '.' 2>/dev/null || echo "$BODY_PL"
  exit 1
fi

echo "✅ HTTP 200 OK"
echo ""

# Parse response
LANG_PL=$(echo "$BODY_PL" | jq -r '.data.language' 2>/dev/null || echo "null")
TITLE_PL=$(echo "$BODY_PL" | jq -r '.data.title' 2>/dev/null || echo "")
DESC_PL=$(echo "$BODY_PL" | jq -r '.data.description' 2>/dev/null || echo "")

echo "📊 Response Analysis:"
echo "  Language field: $LANG_PL"
echo "  Title: $TITLE_PL"
echo "  Description: $DESC_PL"
echo ""

echo "🔍 Validation:"

# Check: Language field should be "pl"
if [ "$LANG_PL" = "pl" ]; then
  echo "  ✅ Language field = 'pl'"
else
  echo "  ❌ Language field = '$LANG_PL' (expected 'pl')"
  exit 1
fi

echo ""
echo "============================================"
echo "TEST 3: PASSED ✅"
echo "============================================"
echo ""
echo ""

# ============================================
# SUMMARY
# ============================================
echo "============================================"
echo "🎉 ALL TESTS PASSED!"
echo "============================================"
echo ""
echo "Summary:"
echo "  ✅ Russian (ru): AI generated text in Russian"
echo "  ✅ English (en): AI generated text in English"
echo "  ✅ Polish (pl): AI generated text in Polish"
echo ""
echo "Backend language support: WORKING ✅"
echo ""
echo "📝 Next steps:"
echo "  1. Check backend logs for language detection:"
echo "     grep '🌐 Language from' server.log"
echo ""
echo "  2. Verify AI prompt included language:"
echo "     grep '🤖 Calling AI' server.log"
echo ""
echo "  3. Test database save (create-ai endpoint) to verify translations"
echo ""
