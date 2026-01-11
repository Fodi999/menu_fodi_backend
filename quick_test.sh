#!/bin/bash

# 🚀 Quick AI Language Test
# Purpose: Verify language detection works

echo "🧪 Testing AI Language Detection..."
echo ""

# Check if server is running
if ! curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo "❌ Server is not running!"
    echo "Start with: ./bin/server"
    exit 1
fi

echo "✅ Server is running"
echo ""

# Get ingredient IDs from suggest endpoint
echo "📋 Getting ingredient IDs..."
SALMON_RESPONSE=$(curl -s "http://localhost:8080/api/admin/ingredients/suggest?q=salmon&limit=1")
SALMON_ID=$(echo "$SALMON_RESPONSE" | jq -r '.[0].id' 2>/dev/null)

if [ "$SALMON_ID" = "null" ] || [ -z "$SALMON_ID" ]; then
    echo "❌ Could not find salmon ingredient"
    echo "Response: $SALMON_RESPONSE"
    exit 1
fi

echo "✅ Found Salmon ID: $SALMON_ID"
echo ""

# Note: This test requires authentication
echo "⚠️  To test fully, you need a JWT token"
echo ""
echo "Get token with:"
echo "  curl -X POST http://localhost:8080/api/auth/login \\"
echo "    -H 'Content-Type: application/json' \\"
echo "    -d '{\"email\": \"your_email\", \"password\": \"your_password\"}' | jq -r '.token'"
echo ""
echo "Then run manual test from TEST_AI_LANGUAGE_MANUAL.md"
echo ""
echo "Example test command:"
echo "  curl -X POST http://localhost:8080/api/admin/recipes/preview-ai \\"
echo "    -H 'Content-Type: application/json' \\"
echo "    -H 'Accept-Language: ru' \\"
echo "    -H 'Authorization: Bearer YOUR_TOKEN' \\"
echo "    -d '{"
echo "      \"title\": \"Жареный лосось\","
echo "      \"language\": \"ru\","
echo "      \"ingredients\": [{\"ingredientId\": \"$SALMON_ID\", \"quantity\": 150, \"unit\": \"g\"}],"
echo "      \"rawCookingText\": \"Обжарить лосось\""
echo "    }' | jq '.data.language'"
echo ""
echo "Expected: \"ru\""
echo ""
echo "📝 Full testing guide: TEST_AI_LANGUAGE_MANUAL.md"
