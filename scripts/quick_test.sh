#!/bin/bash

# Quick local test script - все токены и ID уже захардкожены

BASE_URL="https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app"

# Токены (обнови если истекли - run: ./scripts/login_user.sh)
USER_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiI0MDc1ODJiZS01OWQ1LTRkMjEtODczYi0xYTcyZDMxYjBkNDIiLCJlbWFpbCI6ImZvZGk4NUBnbWFpbC5ydSIsInJvbGUiOiJob21lX2NoZWYiLCJleHAiOjE3Njg2NDcxMDYsImlhdCI6MTc2ODU2MDcwNn0.qFlVLraolHiLd_H3CPetYLt6PssD41KWTxW9yoteOxs"

# Известные ID ингредиентов
EGGS_ID="b4e0c3d2-6f9a-4e5b-8c7d-1a2b3c4d5e6f"  # Jajka
GARLIC_ID="2c3405e0-60cf-4e5f-9872-0bb8d1f91b83"  # Czosnek

echo "=========================================="
echo "⚡ Quick Test Commands"
echo "=========================================="
echo ""

case "${1:-help}" in
  
  # ============================================
  # FRIDGE
  # ============================================
  
  fridge-list)
    echo "📦 Getting fridge items..."
    curl -s "$BASE_URL/api/fridge/items" \
      -H "Authorization: Bearer $USER_TOKEN" | jq '.data | length'
    ;;
  
  fridge-add)
    echo "➕ Adding item to fridge..."
    INGREDIENT_ID="${2:-$GARLIC_ID}"
    QUANTITY="${3:-5.0}"
    curl -s -X POST "$BASE_URL/api/fridge/items" \
      -H "Authorization: Bearer $USER_TOKEN" \
      -H "Content-Type: application/json" \
      -d "{\"ingredientId\":\"$INGREDIENT_ID\",\"quantity\":$QUANTITY}" | jq '.'
    ;;
  
  fridge-delete)
    if [ -z "$2" ]; then
      echo "Usage: $0 fridge-delete <item_id>"
      exit 1
    fi
    echo "🗑️  Deleting item from fridge..."
    curl -s -X DELETE "$BASE_URL/api/fridge/items/$2" \
      -H "Authorization: Bearer $USER_TOKEN" | jq '.'
    ;;
  
  # ============================================
  # NOTIFICATIONS
  # ============================================
  
  notif-count)
    echo "🔔 Notification count..."
    curl -s "$BASE_URL/api/notifications/unread-count" \
      -H "Authorization: Bearer $USER_TOKEN" | jq '.'
    ;;
  
  notif-list)
    echo "📋 Notification list..."
    curl -s "$BASE_URL/api/notifications" \
      -H "Authorization: Bearer $USER_TOKEN" | jq '.data[0:3]'
    ;;
  
  notif-read)
    if [ -z "$2" ]; then
      echo "Usage: $0 notif-read <notification_id>"
      exit 1
    fi
    echo "✅ Marking notification as read..."
    curl -s -X PATCH "$BASE_URL/api/notifications/$2/read" \
      -H "Authorization: Bearer $USER_TOKEN" | jq '.'
    ;;
  
  notif-read-all)
    echo "✅ Marking all notifications as read..."
    curl -s -X POST "$BASE_URL/api/notifications/read-all" \
      -H "Authorization: Bearer $USER_TOKEN" | jq '.'
    ;;
  
  # ============================================
  # AI
  # ============================================
  
  ai-cook)
    echo "🤖 AI: What can I cook now?"
    curl -s -X POST "$BASE_URL/api/ai/recommendation" \
      -H "Authorization: Bearer $USER_TOKEN" \
      -H "Content-Type: application/json" \
      -d '{"scenario":"cook_now"}' | jq '.data.recipes[0] | {title, coverage, source}'
    ;;
  
  ai-fridge)
    echo "🤖 AI: Smart fridge insights..."
    curl -s -X POST "$BASE_URL/api/ai/fridge-insights" \
      -H "Authorization: Bearer $USER_TOKEN" | jq '.'
    ;;
  
  # ============================================
  # INGREDIENTS
  # ============================================
  
  ingredient-search)
    QUERY="${2:-czosnek}"
    echo "🔍 Searching ingredients: $QUERY"
    curl -s "$BASE_URL/api/admin/ingredients/suggest?q=$QUERY&lang=pl" \
      -H "Authorization: Bearer $USER_TOKEN" | jq '.data[0:3] | .[] | {id, name, namePL, category}'
    ;;
  
  # ============================================
  # FLOW TESTS
  # ============================================
  
  flow-add-delete)
    echo "=========================================="
    echo "🔄 Flow Test: Add → Delete → Notification"
    echo "=========================================="
    echo ""
    
    echo "1. Initial notification count:"
    INITIAL=$(curl -s "$BASE_URL/api/notifications/unread-count" \
      -H "Authorization: Bearer $USER_TOKEN" | jq -r '.count')
    echo "   Count: $INITIAL"
    echo ""
    
    echo "2. Adding garlic..."
    ITEM_ID=$(curl -s -X POST "$BASE_URL/api/fridge/items" \
      -H "Authorization: Bearer $USER_TOKEN" \
      -H "Content-Type: application/json" \
      -d "{\"ingredientId\":\"$GARLIC_ID\",\"quantity\":3.0}" | jq -r '.data.id')
    echo "   Item ID: $ITEM_ID"
    echo ""
    
    sleep 1
    
    echo "3. Count after add:"
    COUNT_AFTER_ADD=$(curl -s "$BASE_URL/api/notifications/unread-count" \
      -H "Authorization: Bearer $USER_TOKEN" | jq -r '.count')
    echo "   Count: $COUNT_AFTER_ADD (expected: $((INITIAL + 1)))"
    echo ""
    
    echo "4. Deleting item..."
    curl -s -X DELETE "$BASE_URL/api/fridge/items/$ITEM_ID" \
      -H "Authorization: Bearer $USER_TOKEN" | jq '.'
    echo ""
    
    sleep 1
    
    echo "5. Count after delete:"
    COUNT_AFTER_DELETE=$(curl -s "$BASE_URL/api/notifications/unread-count" \
      -H "Authorization: Bearer $USER_TOKEN" | jq -r '.count')
    echo "   Count: $COUNT_AFTER_DELETE (expected: $((INITIAL + 2)))"
    echo ""
    
    echo "6. Latest notifications:"
    curl -s "$BASE_URL/api/notifications" \
      -H "Authorization: Bearer $USER_TOKEN" | jq '.data[0:2] | .[] | {title, message}'
    echo ""
    
    if [ "$COUNT_AFTER_DELETE" -eq "$((INITIAL + 2))" ]; then
      echo "✅ FLOW TEST PASSED!"
    else
      echo "❌ FLOW TEST FAILED!"
    fi
    ;;
  
  # ============================================
  # HELP
  # ============================================
  
  help|*)
    echo "Usage: $0 <command> [args]"
    echo ""
    echo "Fridge Commands:"
    echo "  fridge-list                    - List all fridge items"
    echo "  fridge-add [id] [quantity]     - Add item (default: garlic, 5g)"
    echo "  fridge-delete <item_id>        - Delete item"
    echo ""
    echo "Notification Commands:"
    echo "  notif-count                    - Get unread count"
    echo "  notif-list                     - List notifications"
    echo "  notif-read <id>                - Mark as read"
    echo "  notif-read-all                 - Mark all as read"
    echo ""
    echo "AI Commands:"
    echo "  ai-cook                        - What can I cook now?"
    echo "  ai-fridge                      - Smart fridge insights"
    echo ""
    echo "Ingredient Commands:"
    echo "  ingredient-search [query]      - Search ingredients"
    echo ""
    echo "Flow Tests:"
    echo "  flow-add-delete                - Test add→delete→notification flow"
    echo ""
    echo "Examples:"
    echo "  $0 fridge-add"
    echo "  $0 notif-count"
    echo "  $0 flow-add-delete"
    echo "  $0 ingredient-search eggs"
    ;;
esac
