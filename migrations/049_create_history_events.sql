-- +goose Up
CREATE TYPE history_event_type AS ENUM ('cook', 'consume', 'waste', 'manual', 'fridge_add', 'fridge_remove');
CREATE TYPE history_source_type AS ENUM ('prepared_dish', 'recipe', 'fridge', 'manual');

CREATE TABLE IF NOT EXISTS history_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT NOT NULL REFERENCES "User"(id) ON DELETE CASCADE,
    event_type history_event_type NOT NULL,
    source_type history_source_type NOT NULL,
    source_id TEXT,  -- UUID or identifier depending on source_type
    
    -- Event details
    portions INT,  -- For cook/consume events
    metadata JSONB,  -- Flexible storage for additional data
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for fetching user's history sorted by date
CREATE INDEX IF NOT EXISTS idx_history_events_user_created ON history_events(user_id, created_at DESC);

-- Index for filtering by event type
CREATE INDEX IF NOT EXISTS idx_history_events_type ON history_events(user_id, event_type);

-- Index for analytics queries on source
CREATE INDEX IF NOT EXISTS idx_history_events_source ON history_events(source_type, source_id);

-- GIN index for JSONB metadata queries
CREATE INDEX IF NOT EXISTS idx_history_events_metadata ON history_events USING GIN (metadata);

COMMENT ON TABLE history_events IS 'Unified event log for all user actions (cook, consume, waste, fridge operations)';
COMMENT ON COLUMN history_events.event_type IS 'Type of event: cook, consume, waste, manual, fridge_add, fridge_remove';
COMMENT ON COLUMN history_events.source_type IS 'What triggered the event: prepared_dish, recipe, fridge, manual';
COMMENT ON COLUMN history_events.source_id IS 'ID of the source entity (prepared_dish.id, recipe.id, etc)';
COMMENT ON COLUMN history_events.metadata IS 'Additional context: recipe_name, portions_remaining, cost, etc';

-- +goose Down
DROP TABLE IF EXISTS history_events;
DROP TYPE IF EXISTS history_source_type;
DROP TYPE IF EXISTS history_event_type;
