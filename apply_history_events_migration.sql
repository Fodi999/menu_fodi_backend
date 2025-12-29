-- ====================================================================
-- 🔧 PRODUCTION MIGRATION: Create history_events table
-- ====================================================================
-- Execute this EXACTLY as written in Koyeb PostgreSQL Console
-- Source: migrations/049_create_history_events.sql
-- ====================================================================

-- Step 1: Create ENUM types
CREATE TYPE history_event_type AS ENUM (
    'cook', 
    'consume', 
    'waste',
    'manual', 
    'fridge_add', 
    'fridge_remove'
);

CREATE TYPE history_source_type AS ENUM (
    'prepared_dish', 
    'recipe', 
    'fridge',
    'manual'
);

-- Step 2: Create table
CREATE TABLE IF NOT EXISTS history_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT NOT NULL REFERENCES "User"(id) ON DELETE CASCADE,
    event_type history_event_type NOT NULL,
    source_type history_source_type NOT NULL,
    source_id TEXT,
    portions INT,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Step 3: Create indexes
CREATE INDEX IF NOT EXISTS idx_history_events_user_created ON history_events(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_history_events_type ON history_events(user_id, event_type);
CREATE INDEX IF NOT EXISTS idx_history_events_source ON history_events(source_type, source_id);
CREATE INDEX IF NOT EXISTS idx_history_events_metadata ON history_events USING GIN (metadata);

-- Step 4: Verify
SELECT 
    'ENUM types created' as check_name,
    COUNT(*) as count 
FROM pg_type 
WHERE typname IN ('history_event_type', 'history_source_type');

SELECT 
    'Table created' as check_name,
    COUNT(*) as count 
FROM information_schema.tables 
WHERE table_name = 'history_events';
