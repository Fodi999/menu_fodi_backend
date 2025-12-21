#!/bin/bash

# Quick check if "Olej roślinny" exists in catalog via API
# This checks the actual backend, not just database

API_URL="${API_URL:-https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app}"
INGREDIENT="Olej roślinny"
TOKEN="${JWT_TOKEN:-}"

echo "=== Checking Ingredient in Catalog via API ==="
echo ""
echo "🔍 Searching for: $INGREDIENT"
echo "🌐 API: $API_URL"
echo ""

if [ -z "$TOKEN" ]; then
    echo "⚠️  JWT_TOKEN not set. Need authentication."
    echo ""
    echo "To test with auth:"
    echo "  export JWT_TOKEN='your_jwt_token_here'"
    echo "  ./check_olej_in_api.sh"
    echo ""
    echo "📋 For now, check via SQL on Neon.tech:"
    echo ""
    echo "SELECT id, name, category, unit"
    echo "FROM \"Ingredient\""
    echo "WHERE LOWER(name) LIKE '%olej%'"
    echo "ORDER BY name;"
    echo ""
    exit 0
fi

# Search ingredient via API
echo "📡 Calling: GET /catalog/ingredients/search?query=olej"
echo ""

curl -s -H "Authorization: Bearer $TOKEN" "$API_URL/catalog/ingredients/search?query=olej" | jq '.'

echo ""
echo "---"
echo ""
echo "Alternative: Search for 'roślinny'"
echo ""

curl -s -H "Authorization: Bearer $TOKEN" "$API_URL/catalog/ingredients/search?query=roślinny" | jq '.'

echo ""
echo "---"
echo ""
echo "💡 If NOT found:"
echo "1. Go to Neon.tech dashboard"
echo "2. Run migration: ./apply_migration_034.sh"
echo "3. Copy SQL and execute"
echo "4. Re-run this script"
echo ""
echo "✅ If found:"
echo "- User can add 'Olej roślinny' to fridge"
echo "- AI recipes can use it in ingredientsMissing"
echo "- AddMissingIngredients will work"
