#!/bin/bash

# Test Notification System
# This script manually triggers notification generation for testing

echo "🧪 Testing Notification System"
echo "================================"

USER_ID="407582be-59d5-4d21-873b-1a72d31b0d42"

# 1. Check expired items in database
echo ""
echo "📊 Step 1: Checking expired items in database..."
source <(grep DATABASE_URL .env | grep -v UNPOOLED)
psql "$DATABASE_URL" -c "
SELECT 
  i.name,
  f.quantity,
  f.unit,
  f.expires_at,
  EXTRACT(DAY FROM (f.expires_at - NOW())) as days_left
FROM user_fridge_items f
LEFT JOIN \"Ingredient\" i ON f.ingredient_id = i.id
WHERE f.user_id = '$USER_ID'
  AND f.expires_at IS NOT NULL
  AND f.expires_at < NOW()
ORDER BY f.expires_at ASC
LIMIT 10;
"

# 2. Check notifications table
echo ""
echo "📬 Step 2: Checking existing notifications..."
psql "$DATABASE_URL" -c "
SELECT 
  type,
  level,
  title,
  LEFT(message, 60) as message_preview,
  created_at
FROM notifications
WHERE user_id = '$USER_ID'
ORDER BY created_at DESC
LIMIT 5;
"

# 3. API Test - Get notifications
echo ""
echo "🌐 Step 3: Testing API endpoint..."
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiI0MDc1ODJiZS01OWQ1LTRkMjEtODczYi0xYTcyZDMxYjBkNDIiLCJlbWFpbCI6ImZvZGk4NUBnbWFpbC5ydSIsInJvbGUiOiJob21lX2NoZWYiLCJleHAiOjE3Njg1NjAyOTcsImlhdCI6MTc2ODQ3Mzg5N30.wWSasbP-1WVnVIst7_HMCpAXNRfwAHWQIUhqKEorHYY"

curl -s "https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/notifications" \
  -H "Authorization: Bearer $TOKEN" | jq '.'

echo ""
echo "📊 Step 4: Notification count..."
curl -s "https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/notifications/unread-count" \
  -H "Authorization: Bearer $TOKEN" | jq '.'

# 4. CRON explanation
echo ""
echo "⏰ CRON Schedule Info:"
echo "================================"
echo "Schedule: Daily at 08:00 UTC"
echo "Next run: Tomorrow at 08:00 UTC (09:00 Warsaw, 11:00 Moscow)"
echo ""
echo "What CRON does:"
echo "1. Finds all users with items in fridge"
echo "2. For each user:"
echo "   - Checks items with expires_at"
echo "   - Calculates daysLeft"
echo "   - Creates notifications:"
echo "     • daysLeft < 0  → CRITICAL (expired)"
echo "     • daysLeft = 0  → CRITICAL (expires today)"
echo "     • daysLeft = 1  → WARNING (expires tomorrow)"
echo "     • daysLeft 2-3  → INFO (expires soon)"
echo "3. Saves to notifications table"
echo "4. User sees in GET /api/notifications"
echo ""
echo "Expected notifications for user $USER_ID:"
echo "- 7 expired items → 7 CRITICAL notifications"
echo ""
echo "✅ Test completed!"
