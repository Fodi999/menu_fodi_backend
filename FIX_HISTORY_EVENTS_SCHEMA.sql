-- ========================================
-- PRODUCTION FIX: history_events table schema
-- ========================================
-- Date: 2026-01-16
-- Issue: "column source_type of relation history_events does not exist"
-- Root Cause: Migration 049 was not fully applied OR table was created manually

-- Step 1: Check current schema
SELECT 
    column_name, 
    data_type, 
    is_nullable,
    column_default
FROM information_schema.columns 
WHERE table_name = 'history_events'
ORDER BY ordinal_position;

-- Step 2: Check if source_type column exists
SELECT column_name 
FROM information_schema.columns 
WHERE table_name = 'history_events' 
  AND column_name = 'source_type';

-- If (0 rows), then we need to add the column:
-- NOTE: First check if enum types exist

-- Step 3: Create enum types if they don't exist
DO $$ 
BEGIN
    -- Create history_event_type enum
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'history_event_type') THEN
        CREATE TYPE history_event_type AS ENUM ('cook', 'consume', 'waste', 'manual', 'fridge_add', 'fridge_remove');
    END IF;
    
    -- Create history_source_type enum
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'history_source_type') THEN
        CREATE TYPE history_source_type AS ENUM ('prepared_dish', 'recipe', 'fridge', 'manual');
    END IF;
END $$;

-- Step 4: Add source_type column if missing
ALTER TABLE history_events 
ADD COLUMN IF NOT EXISTS source_type history_source_type NOT NULL DEFAULT 'manual';

-- Step 5: Add source_id column if missing
ALTER TABLE history_events 
ADD COLUMN IF NOT EXISTS source_id TEXT;

-- Step 6: Create indexes if they don't exist
CREATE INDEX IF NOT EXISTS idx_history_events_user_created 
ON history_events(user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_history_events_type 
ON history_events(user_id, event_type);

CREATE INDEX IF NOT EXISTS idx_history_events_source 
ON history_events(source_type, source_id);

CREATE INDEX IF NOT EXISTS idx_history_events_metadata 
ON history_events USING GIN (metadata);

-- Step 7: Verify the fix
SELECT 
    table_name,
    column_name, 
    data_type, 
    is_nullable
FROM information_schema.columns 
WHERE table_name = 'history_events' 
  AND column_name IN ('source_type', 'source_id')
ORDER BY ordinal_position;

-- Expected output:
-- table_name     | column_name  | data_type           | is_nullable
-- history_events | source_type  | USER-DEFINED        | NO
-- history_events | source_id    | text                | YES

-- Step 8: Sample data check
SELECT 
    id,
    user_id,
    event_type,
    source_type,
    source_id,
    created_at
FROM history_events
ORDER BY created_at DESC
LIMIT 5;
