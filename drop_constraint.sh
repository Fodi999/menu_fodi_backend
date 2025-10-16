#!/bin/bash

# Load DATABASE_URL from .env or environment
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

if [ -z "$DATABASE_URL" ]; then
    echo "❌ ERROR: DATABASE_URL not set"
    exit 1
fi

echo "🔧 Dropping foreign key constraint fk_Business_owner..."

psql "$DATABASE_URL" -c "ALTER TABLE \"Business\" DROP CONSTRAINT IF EXISTS \"fk_Business_owner\";"

if [ $? -eq 0 ]; then
    echo "✅ Constraint dropped successfully"
else
    echo "❌ Failed to drop constraint"
    exit 1
fi

echo "🔧 Making owner_id nullable..."

psql "$DATABASE_URL" -c "ALTER TABLE \"Business\" ALTER COLUMN \"owner_id\" DROP NOT NULL;"

if [ $? -eq 0 ]; then
    echo "✅ Column is now nullable"
else
    echo "❌ Failed to alter column"
    exit 1
fi

echo "✅ Migration completed successfully!"
