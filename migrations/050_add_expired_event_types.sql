-- +goose Up
-- Add 'expired' to history_event_type enum
ALTER TYPE history_event_type ADD VALUE IF NOT EXISTS 'expired';

-- Add 'auto' to history_source_type enum
ALTER TYPE history_source_type ADD VALUE IF NOT EXISTS 'auto';

-- +goose Down
-- Cannot remove enum values in PostgreSQL, but for reference:
-- Would need to recreate the entire enum type
