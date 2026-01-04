#!/bin/bash

# Quick User Activity Check Script
# Usage: ./check_activity.sh

DB_URL="postgresql://neondb_owner:npg_dz4Gl8ZhPLbX@ep-soft-mud-agon8wu3-pooler.c-2.eu-central-1.aws.neon.tech/neondb?sslmode=require"

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📊 USER ACTIVITY STATISTICS"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Overall stats
psql "$DB_URL" -c "
SELECT
  COUNT(*) AS \"📈 Total Users\",
  COUNT(*) FILTER (WHERE last_login >= DATE_TRUNC('day', NOW())) AS \"🟢 Active Today\",
  COUNT(*) FILTER (WHERE last_login >= NOW() - INTERVAL '7 days') AS \"🟡 Active This Week\",
  COUNT(*) FILTER (WHERE last_login >= NOW() - INTERVAL '30 days') AS \"🟠 Active This Month\",
  COUNT(*) FILTER (WHERE last_login < NOW() - INTERVAL '30 days') AS \"🔴 Inactive 30+ days\",
  COUNT(*) FILTER (WHERE status = 'blocked') AS \"⛔ Blocked\"
FROM \"User\";
"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "👥 ACTIVITY BY ROLE"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

psql "$DB_URL" -c "
SELECT
  role AS \"Role\",
  COUNT(*) AS \"Total\",
  COUNT(*) FILTER (WHERE last_login >= DATE_TRUNC('day', NOW())) AS \"Active 24h\",
  COUNT(*) FILTER (WHERE last_login >= NOW() - INTERVAL '7 days') AS \"Active 7d\",
  COUNT(*) FILTER (WHERE last_login >= NOW() - INTERVAL '30 days') AS \"Active 30d\",
  ROUND(COUNT(*) FILTER (WHERE last_login >= NOW() - INTERVAL '30 days') * 100.0 / NULLIF(COUNT(*), 0), 1) || '%' AS \"Active %\"
FROM \"User\"
GROUP BY role
ORDER BY COUNT(*) DESC;
"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🔴 CHURN RISK (Inactive 14+ days)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

psql "$DB_URL" -c "
SELECT
  COUNT(*) AS \"At Risk Users\",
  ROUND(COUNT(*) * 100.0 / (SELECT COUNT(*) FROM \"User\" WHERE status = 'active'), 1) || '%' AS \"Churn Risk %\"
FROM \"User\"
WHERE last_login < NOW() - INTERVAL '14 days'
  AND status = 'active';
"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🔝 TOP 10 MOST ACTIVE USERS (Recent Logins)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

psql "$DB_URL" -c "
SELECT
  name AS \"Name\",
  email AS \"Email\",
  role AS \"Role\",
  TO_CHAR(last_login, 'YYYY-MM-DD HH24:MI') AS \"Last Login\",
  CASE
    WHEN last_login >= NOW() - INTERVAL '1 hour' THEN '🟢 Just now'
    WHEN last_login >= NOW() - INTERVAL '24 hours' THEN '🟢 Today'
    WHEN last_login >= NOW() - INTERVAL '7 days' THEN '🟡 This week'
    WHEN last_login >= NOW() - INTERVAL '30 days' THEN '🟠 This month'
    ELSE '🔴 Long ago'
  END AS \"Status\"
FROM \"User\"
ORDER BY last_login DESC NULLS LAST
LIMIT 10;
"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ Done! For detailed analysis, run:"
echo "   psql \"$DB_URL\" -f sql/check_user_activity.sql"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
