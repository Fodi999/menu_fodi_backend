#!/bin/bash

# Test AI ingredient classification - single field creation
# Usage: ./test_ai_ingredient.sh YOUR_JWT_TOKEN "Соль каменная"

TOKEN="${1}"
INGREDIENT_NAME="${2:-Соль каменная}"

if [ -z "$TOKEN" ]; then
  echo "❌ Error: JWT token required"
  echo "Usage: ./test_ai_ingredient.sh YOUR_JWT_TOKEN [ingredient_name]"
  echo "Example: ./test_ai_ingredient.sh eyJhbGc... 'Соль каменная'"
  exit 1
fi

API_BASE="http://localhost:8080"

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🤖 Testing AI Ingredient Classification"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Input: $INGREDIENT_NAME"
echo ""

RESPONSE=$(curl -s -X POST "$API_BASE/api/admin/ingredients" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"inputName\":\"$INGREDIENT_NAME\"}")

echo "$RESPONSE" | jq '.'

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ AI должен был определить:"
echo "  - Язык исходного названия"
echo "  - Переводы на PL/EN/RU"
echo "  - Категорию (vegetable/fruit/protein/dairy/grain/condiment/other)"
echo "  - Единицу измерения (g/ml/pcs)"
echo "  - Normalized value (для проверки дубликатов)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
