-- ========================================
-- User Activity Analysis
-- ========================================

-- 1️⃣ ОБЩАЯ СТАТИСТИКА (как в API)
SELECT
  COUNT(*)                                        AS total_users,
  COUNT(*) FILTER (
    WHERE last_login >= DATE_TRUNC('day', NOW())
  )                                               AS active_today,
  COUNT(*) FILTER (
    WHERE last_login >= NOW() - INTERVAL '7 days'
  )                                               AS active_this_week,
  COUNT(*) FILTER (
    WHERE status = 'blocked'
  )                                               AS blocked_users,
  COUNT(*) FILTER (
    WHERE last_login IS NULL
  )                                               AS never_logged_in
FROM "User";


-- 2️⃣ ТОП-10 АКТИВНЫХ ПОЛЬЗОВАТЕЛЕЙ (недавние логины)
SELECT
  id,
  name,
  email,
  role,
  status,
  last_login,
  NOW() - last_login AS time_since_login
FROM "User"
WHERE last_login IS NOT NULL
ORDER BY last_login DESC
LIMIT 10;


-- 3️⃣ НЕАКТИВНЫЕ ПОЛЬЗОВАТЕЛИ (не заходили больше 7 дней)
SELECT
  id,
  name,
  email,
  role,
  status,
  last_login,
  NOW() - last_login AS inactive_for,
  "createdAt" AS registered_at
FROM "User"
WHERE last_login < NOW() - INTERVAL '7 days'
  AND status = 'active'
ORDER BY last_login ASC;


-- 4️⃣ ПОЛЬЗОВАТЕЛИ, КОТОРЫЕ НИКОГДА НЕ ЗАХОДИЛИ
SELECT
  id,
  name,
  email,
  role,
  status,
  "createdAt" AS registered_at,
  NOW() - "createdAt" AS registered_ago
FROM "User"
WHERE last_login IS NULL
ORDER BY "createdAt" DESC;


-- 5️⃣ АКТИВНОСТЬ ПО ДНЯМ (за последние 7 дней)
SELECT
  DATE(last_login) AS login_date,
  COUNT(*) AS unique_users,
  COUNT(DISTINCT role) AS different_roles
FROM "User"
WHERE last_login >= NOW() - INTERVAL '7 days'
GROUP BY DATE(last_login)
ORDER BY login_date DESC;


-- 6️⃣ РАСПРЕДЕЛЕНИЕ ПО СТАТУСАМ
SELECT
  status,
  COUNT(*) AS count,
  ROUND(COUNT(*) * 100.0 / SUM(COUNT(*)) OVER (), 2) AS percentage
FROM "User"
GROUP BY status
ORDER BY count DESC;


-- 7️⃣ АКТИВНОСТЬ ПО РОЛЯМ
SELECT
  role,
  COUNT(*) AS total,
  COUNT(*) FILTER (
    WHERE last_login >= DATE_TRUNC('day', NOW())
  ) AS active_today,
  COUNT(*) FILTER (
    WHERE last_login >= NOW() - INTERVAL '7 days'
  ) AS active_week,
  ROUND(
    COUNT(*) FILTER (WHERE last_login >= NOW() - INTERVAL '7 days') * 100.0 / 
    NULLIF(COUNT(*), 0), 
    2
  ) AS active_week_percentage
FROM "User"
GROUP BY role
ORDER BY total DESC;


-- 8️⃣ ПОЛНЫЙ СПИСОК ВСЕХ ПОЛЬЗОВАТЕЛЕЙ С АКТИВНОСТЬЮ
SELECT
  id,
  name,
  email,
  role,
  status,
  last_login,
  CASE
    WHEN last_login IS NULL THEN '❌ Ніколи'
    WHEN last_login >= NOW() - INTERVAL '1 hour' THEN '🟢 Щойно'
    WHEN last_login >= NOW() - INTERVAL '24 hours' THEN '🟢 Сьогодні'
    WHEN last_login >= NOW() - INTERVAL '7 days' THEN '🟡 Цього тижня'
    WHEN last_login >= NOW() - INTERVAL '30 days' THEN '🟠 Цього місяця'
    ELSE '🔴 Давно'
  END AS activity_status,
  "createdAt" AS registered_at
FROM "User"
ORDER BY 
  CASE 
    WHEN last_login IS NULL THEN '9999-12-31'::timestamp
    ELSE last_login 
  END DESC;


-- 9️⃣ РИСК ВІДТОКУ (Churn Risk) - не заходили 14+ днів
SELECT
  COUNT(*) AS at_risk_users,
  ROUND(COUNT(*) * 100.0 / (SELECT COUNT(*) FROM "User" WHERE status = 'active'), 2) AS churn_risk_percentage
FROM "User"
WHERE (last_login < NOW() - INTERVAL '14 days' OR last_login IS NULL)
  AND status = 'active';


-- 🔟 ДЕТАЛЬНИЙ АНАЛІЗ РИЗИКУ ВІДТОКУ
SELECT
  id,
  name,
  email,
  role,
  last_login,
  CASE
    WHEN last_login IS NULL THEN 'Never logged in'
    ELSE EXTRACT(DAY FROM NOW() - last_login)::TEXT || ' days ago'
  END AS last_seen,
  "createdAt" AS registered_at
FROM "User"
WHERE (last_login < NOW() - INTERVAL '14 days' OR last_login IS NULL)
  AND status = 'active'
ORDER BY last_login ASC NULLS FIRST;
