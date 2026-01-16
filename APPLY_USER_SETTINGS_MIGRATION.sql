-- ========================================
-- PRODUCTION MIGRATION: Add User.settings
-- ========================================
-- Date: 2026-01-16
-- Issue: Frontend language switch failing with "column settings does not exist"
-- Fix: Add JSONB settings column to User table

-- Step 1: Add settings column with default values
ALTER TABLE "User"
ADD COLUMN IF NOT EXISTS settings JSONB DEFAULT '{
  "language": "pl",
  "timeFormat": "24h",
  "units": "metric",
  "aiStyle": "mentor"
}'::jsonb NOT NULL;

-- Step 2: Add GIN index for faster language queries
-- GIN index is optimal for JSONB queries like: WHERE settings->>'language' = 'ru'
CREATE INDEX IF NOT EXISTS idx_user_settings_language 
ON "User" USING gin ((settings->'language'));

-- Step 3: Verify column was added
SELECT 
    table_name,
    column_name, 
    data_type, 
    is_nullable,
    column_default
FROM information_schema.columns 
WHERE table_name = 'User' 
  AND column_name = 'settings';

-- Step 4: Show sample data (first 3 users)
SELECT 
    id,
    email,
    settings
FROM "User"
LIMIT 3;

-- Expected output:
-- ✅ Column: settings | Type: jsonb | Nullable: NO
-- ✅ Default: {"language":"pl","timeFormat":"24h","units":"metric","aiStyle":"mentor"}
-- ✅ Index: idx_user_settings_language created

-- After this migration:
-- ✅ Frontend can call PATCH /api/settings
-- ✅ Backend can UPDATE "User" SET settings = ...
-- ✅ Language switching works without page reload
