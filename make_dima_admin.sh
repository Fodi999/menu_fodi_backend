#!/bin/bash

# Make Dima's account admin
# Usage: ./make_dima_admin.sh EMAIL

EMAIL="${1:-fodi8521@gmail.com}"

echo "🔧 Making $EMAIL an admin..."

source .env

psql "$DATABASE_URL" << SQL
UPDATE "User"
SET role = 'admin'
WHERE email = '$EMAIL';

-- Verify
SELECT 
  name, 
  email, 
  role,
  status
FROM "User"
WHERE email = '$EMAIL';
SQL

echo ""
echo "✅ Done! $EMAIL is now an admin."
echo ""
echo "🔍 All admins now:"
psql "$DATABASE_URL" -c "SELECT name, email, role FROM \"User\" WHERE role = 'admin' ORDER BY \"createdAt\" DESC;"
