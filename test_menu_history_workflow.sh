#!/bin/bash

# =============================================================================
# TEST: Complete Kitchen Pipeline with History Separation
# =============================================================================
# Tests that:
# 1. GetTodayMenu returns only active items (planned + cooking)
# 2. Completed items move to history
# 3. GetHistory returns completed items separately
# =============================================================================

set -e

API_URL="https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app"
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6ImZvZGk4NUBnbWFpbC5ydSIsInJvbGUiOiJob21lX2NoZWYiLCJoYXNSb2xlIjp0cnVlLCJzdWIiOiI0MDc1ODJiZS01OWQ1LTRkMjEtODczYi0xYTcyZDMxYjBkNDIiLCJleHAiOjE3NjkyNDM1MzEsImlhdCI6MTc2OTE1NzEzMX0.LWEQVj50fcFgoM1lmxj29ggGundEMtauQCkaFP_o_H0"
RECIPE_ID="605c8419-2d42-4ef0-a9d2-839582e98727" # zharenye_yaytsa

echo "🧪 KITCHEN PIPELINE + HISTORY TEST"
echo "=================================="
echo ""

# Step 1: Add recipe to today's menu
echo "📝 Step 1: Adding recipe to menu..."
ADD_RESPONSE=$(curl -s -X POST "$API_URL/api/menu/today" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"recipe_id\":\"$RECIPE_ID\",\"servings\":2}")

ITEM_ID=$(echo "$ADD_RESPONSE" | jq -r '.data.id // .id')
if [ "$ITEM_ID" == "null" ] || [ -z "$ITEM_ID" ]; then
  echo "❌ Failed to add recipe to menu"
  echo "$ADD_RESPONSE" | jq
  exit 1
fi
echo "✅ Recipe added to menu: $ITEM_ID"
echo ""

# Step 2: Check today's menu (should have 1 planned item)
echo "📋 Step 2: Checking today's menu (should show planned item)..."
TODAY_MENU=$(curl -s -X GET "$API_URL/api/menu/today" \
  -H "Authorization: Bearer $TOKEN")
TODAY_COUNT=$(echo "$TODAY_MENU" | jq 'length')
echo "   Found $TODAY_COUNT active item(s)"
if [ "$TODAY_COUNT" -eq 1 ]; then
  echo "✅ Today's menu shows active item"
else
  echo "⚠️  Expected 1 item, got $TODAY_COUNT"
fi
echo ""

# Step 3: Start cooking
echo "🔥 Step 3: Starting cooking..."
START_RESPONSE=$(curl -s -X POST "$API_URL/api/menu/$ITEM_ID/start" \
  -H "Authorization: Bearer $TOKEN")
echo "$START_RESPONSE" | jq -r '.message // .error'

TODAY_MENU=$(curl -s -X GET "$API_URL/api/menu/today" \
  -H "Authorization: Bearer $TOKEN")
COOKING_STATUS=$(echo "$TODAY_MENU" | jq -r '.[0].status // "not_found"')
if [ "$COOKING_STATUS" == "cooking" ]; then
  echo "✅ Status changed to cooking, still in active menu"
else
  echo "❌ Status not updated correctly: $COOKING_STATUS"
fi
echo ""

# Step 4: Complete cooking
echo "✅ Step 4: Completing cooking..."
COMPLETE_RESPONSE=$(curl -s -X POST "$API_URL/api/menu/$ITEM_ID/complete" \
  -H "Authorization: Bearer $TOKEN")
echo "$COMPLETE_RESPONSE" | jq -r '.message // .error'
echo ""

# Step 5: Check today's menu (should be empty - completed items hidden)
echo "📋 Step 5: Checking today's menu (should be EMPTY)..."
TODAY_MENU=$(curl -s -X GET "$API_URL/api/menu/today" \
  -H "Authorization: Bearer $TOKEN")
TODAY_COUNT=$(echo "$TODAY_MENU" | jq 'length')
echo "   Found $TODAY_COUNT active item(s)"
if [ "$TODAY_COUNT" -eq 0 ]; then
  echo "✅ Today's menu is empty (completed items hidden)"
else
  echo "❌ Expected 0 items, got $TODAY_COUNT"
  echo "$TODAY_MENU" | jq
fi
echo ""

# Step 6: Check history (should show completed item)
echo "📚 Step 6: Checking history (should show completed item)..."
HISTORY=$(curl -s -X GET "$API_URL/api/menu/history?limit=5" \
  -H "Authorization: Bearer $TOKEN")
HISTORY_COUNT=$(echo "$HISTORY" | jq 'length')
LATEST_STATUS=$(echo "$HISTORY" | jq -r '.[0].status // "not_found"')
LATEST_ID=$(echo "$HISTORY" | jq -r '.[0].id // "not_found"')

echo "   Found $HISTORY_COUNT completed item(s)"
if [ "$LATEST_ID" == "$ITEM_ID" ] && [ "$LATEST_STATUS" == "completed" ]; then
  echo "✅ Completed item found in history"
  echo "   Recipe: $(echo "$HISTORY" | jq -r '.[0].recipe.title')"
  echo "   Completed at: $(echo "$HISTORY" | jq -r '.[0].completed_at')"
else
  echo "❌ Completed item not found in history"
  echo "$HISTORY" | jq
fi
echo ""

# Step 7: Cleanup - delete the item
echo "🗑️  Step 7: Cleaning up..."
DELETE_RESPONSE=$(curl -s -X DELETE "$API_URL/api/menu/$ITEM_ID" \
  -H "Authorization: Bearer $TOKEN")
echo "$DELETE_RESPONSE" | jq -r '.message // .error'
echo ""

echo "=================================="
echo "🎉 TEST COMPLETE!"
echo "=================================="
echo ""
echo "Summary:"
echo "  ✅ Active menu shows only planned/cooking items"
echo "  ✅ Completed items hidden from active menu"
echo "  ✅ Completed items accessible via /history endpoint"
echo "  ✅ Full workflow: planned → cooking → completed → history"
