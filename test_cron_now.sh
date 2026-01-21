#!/bin/bash

# 🧪 CRON TEST SCRIPT
# Запускает проверку истекающих продуктов НЕМЕДЛЕННО (не ждёт 08:00)
# Используется для отладки и проверки работы уведомлений

set -e

echo "🧪 Testing CRON expiry checker..."
echo "================================"
echo ""

# Компилируем утилиту
echo "📦 Building test utility..."
go build -o bin/test_cron ./cmd/test_cron

# Запускаем
echo ""
echo "🚀 Running expiry check NOW..."
./bin/test_cron

echo ""
echo "✅ Test completed!"
echo ""
echo "📊 Next steps:"
echo "  1. Check logs above for processed users"
echo "  2. Verify notifications: GET /api/notifications"
echo "  3. Check unread count: GET /api/notifications/unread-count"
