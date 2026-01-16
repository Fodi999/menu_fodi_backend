#!/bin/bash

# Quick check if User.settings column exists in production

echo "🔍 Checking if User.settings column exists..."
echo ""

# You need to replace these with actual Neon credentials
# Get from: https://console.neon.tech/ or Koyeb environment variables

if [ -z "$DATABASE_URL" ]; then
  echo "⚠️  DATABASE_URL not set"
  echo ""
  echo "Set it with:"
  echo '  export DATABASE_URL="postgresql://user:password@host/db?sslmode=require"'
  echo ""
  echo "Or run manually:"
  echo '  psql "YOUR_CONNECTION_STRING" -c "SELECT column_name FROM information_schema.columns WHERE table_name='"'"'User'"'"' AND column_name='"'"'settings'"'"';"'
  exit 1
fi

# Check if settings column exists
RESULT=$(psql "$DATABASE_URL" -t -c "SELECT column_name FROM information_schema.columns WHERE table_name='User' AND column_name='settings';" 2>&1)

if echo "$RESULT" | grep -q "settings"; then
  echo "✅ SUCCESS: User.settings column EXISTS"
  echo ""
  echo "Sample data:"
  psql "$DATABASE_URL" -c "SELECT id, email, settings FROM \"User\" LIMIT 3;"
  echo ""
  echo "🎉 Migration already applied!"
else
  echo "❌ MISSING: User.settings column DOES NOT EXIST"
  echo ""
  echo "📋 Apply migration with:"
  echo "   psql \"\$DATABASE_URL\" -f APPLY_USER_SETTINGS_MIGRATION.sql"
  echo ""
  echo "Or manually:"
  echo "   psql \"\$DATABASE_URL\" -c 'ALTER TABLE \"User\" ADD COLUMN settings JSONB DEFAULT '\"'{\"language\":\"pl\",\"timeFormat\":\"24h\",\"units\":\"metric\",\"aiStyle\":\"mentor\"}'\"'::jsonb NOT NULL;'"
fi
