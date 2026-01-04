#!/bin/bash
source .env
psql "$DATABASE_URL" << SQL
SELECT 
  name, 
  email, 
  role,
  status,
  "createdAt",
  last_login
FROM "User"
WHERE email LIKE '%dima%' OR name LIKE '%Dima%'
ORDER BY "createdAt" DESC;
SQL
