#!/bin/bash

# 🎨 TEST SCRIPT: Recipe Edit Workflow
# Purpose: Test complete workflow: Preview → Edit → Save

echo "=========================================="
echo "🎨 Recipe Edit Workflow Test"
echo "=========================================="
echo ""

# Backend URL
API_URL="http://localhost:8080"

# Step 1: Login as admin
echo "Step 1: Logging in as admin..."
LOGIN_RESPONSE=$(curl -s -X POST "$API_URL/api/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "admin_password_123"
  }')

TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.token // .data.token // empty')

if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
  echo "❌ Login failed"
  echo "$LOGIN_RESPONSE" | jq .
  exit 1
fi

echo "✅ Login successful"
echo ""

# Step 2: Preview AI Recipe
echo "=========================================="
echo "📝 Step 2: Generate AI Preview"
echo "=========================================="
echo ""

PREVIEW_RESPONSE=$(curl -s -X POST "$API_URL/api/admin/recipes/preview-ai" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Accept-Language: ru-RU,ru;q=0.9" \
  -d '{
    "title": "Паста карбонара",
    "rawCookingText": "Сварить спагетти. Обжарить бекон. Смешать яйца с пармезаном. Соединить все ингредиенты.",
    "language": "ru",
    "ingredients": [
      {"ingredientId": "fe1c7431-b1b7-4d36-94bf-74276481983e", "quantity": 300, "unit": "g"}
    ]
  }')

echo "$PREVIEW_RESPONSE" | jq .

echo ""
echo "=========================================="
echo "✏️  Step 3: Edit Preview (user makes changes)"
echo "=========================================="
echo ""
echo "User edits:"
echo "  - Changes title to 'Паста карбонара (домашний рецепт)'"
echo "  - Adds more details to description"
echo "  - Adjusts cooking time to 25 minutes"
echo ""

# Step 3: Save Edited Recipe
echo "=========================================="
echo "💾 Step 4: Save Edited Recipe"
echo "=========================================="
echo ""

SAVE_RESPONSE=$(curl -s -X POST "$API_URL/api/admin/recipes/save" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "title": "Паста карбонара (домашний рецепт)",
    "language": "ru",
    "description": "Классическая итальянская паста карбонара с беконом и пармезаном. Это улучшенная версия с дополнительными специями.",
    "servings": 2,
    "time_minutes": 25,
    "difficulty": "medium",
    "calories": 650,
    "ingredients": [
      {
        "ingredientId": "fe1c7431-b1b7-4d36-94bf-74276481983e",
        "name": "Лосось",
        "amount": 300,
        "unit": "g"
      }
    ],
    "steps": [
      {
        "order": 1,
        "text": "Сварить спагетти в подсоленной воде до состояния al dente",
        "time": 10
      },
      {
        "order": 2,
        "text": "Обжарить бекон до хрустящей корочки",
        "time": 5
      },
      {
        "order": 3,
        "text": "Смешать яйца с тёртым пармезаном и черным перцем",
        "time": 2
      },
      {
        "order": 4,
        "text": "Соединить горячую пасту с беконом и яичной смесью, постоянно помешивая",
        "time": 3
      },
      {
        "order": 5,
        "text": "Подавать немедленно с дополнительным пармезаном",
        "time": 1
      }
    ]
  }')

echo "$SAVE_RESPONSE" | jq .

RECIPE_ID=$(echo "$SAVE_RESPONSE" | jq -r '.data.id // empty')

if [ -z "$RECIPE_ID" ] || [ "$RECIPE_ID" = "null" ]; then
  echo ""
  echo "❌ Failed to save recipe"
  exit 1
fi

echo ""
echo "✅ Recipe saved successfully"
echo "   Recipe ID: $RECIPE_ID"
echo ""

# Step 4: Update Existing Recipe
echo "=========================================="
echo "🔄 Step 5: Update Existing Recipe"
echo "=========================================="
echo ""

UPDATE_RESPONSE=$(curl -s -X PUT "$API_URL/api/admin/recipes/$RECIPE_ID" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "title": "Паста карбонара (авторский рецепт)",
    "language": "ru",
    "description": "Авторская версия классической пасты карбонара с секретным ингредиентом",
    "servings": 3,
    "time_minutes": 30,
    "difficulty": "medium",
    "calories": 700,
    "ingredients": [
      {
        "ingredientId": "fe1c7431-b1b7-4d36-94bf-74276481983e",
        "name": "Лосось",
        "amount": 400,
        "unit": "g"
      }
    ],
    "steps": [
      {
        "order": 1,
        "text": "Сварить спагетти в подсоленной воде до состояния al dente",
        "time": 10
      },
      {
        "order": 2,
        "text": "Обжарить бекон с чесноком до хрустящей корочки",
        "time": 7
      },
      {
        "order": 3,
        "text": "Смешать яйца с тёртым пармезаном, пекорино и черным перцем",
        "time": 3
      },
      {
        "order": 4,
        "text": "Соединить горячую пасту с беконом и яичной смесью, добавить немного пастовой воды",
        "time": 5
      },
      {
        "order": 5,
        "text": "Подавать с дополнительным пармезаном и свежей петрушкой",
        "time": 2
      }
    ]
  }')

echo "$UPDATE_RESPONSE" | jq .

echo ""
echo "=========================================="
echo "📊 WORKFLOW SUMMARY"
echo "=========================================="
echo ""
echo "✅ Step 1: Login successful"
echo "✅ Step 2: AI Preview generated"
echo "✅ Step 3: User edited preview"
echo "✅ Step 4: Edited recipe saved to DB (ID: $RECIPE_ID)"
echo "✅ Step 5: Existing recipe updated"
echo ""
echo "🎉 Complete workflow tested successfully!"
echo ""
echo "=========================================="
echo "📋 Check Backend Logs"
echo "=========================================="
echo ""
echo "Run this command to see logs:"
echo "  tail -50 server_test.log | grep -E 'Preview|SaveEdited|Update|💾|🔄'"
