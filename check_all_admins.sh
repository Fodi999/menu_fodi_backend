#!/bin/bash
source .env
psql "$DATABASE_URL" << SQL
SELECT 
  name, 
  email, 
  role,
  status,
  "createdAt"
FROM "User"
WHERE role = 'admin'
ORDER BY "createdAt" DESC;
SQL
