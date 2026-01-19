#!/bin/bash
# Apply usage_count migration to production database

source .env

echo "🔧 Applying usage_count migration to production..."
psql "$DATABASE_URL" -f migrations/add_usage_count_to_ingredients.sql

echo "✅ Migration applied successfully!"
echo "📊 Checking usage_count values..."
psql "$DATABASE_URL" -c "SELECT name, usage_count FROM \"Ingredient\" ORDER BY usage_count DESC LIMIT 10;"
