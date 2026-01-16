-- Remove obsolete entity_type and entity_id columns from history_events
-- These columns are from old schema and are not used in current Go model

BEGIN;

-- Drop NOT NULL constraint first (if needed), then drop columns
ALTER TABLE history_events DROP COLUMN IF EXISTS entity_type;
ALTER TABLE history_events DROP COLUMN IF EXISTS entity_id;

-- Verify remaining columns match Go model
\d history_events

COMMIT;
