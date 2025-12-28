-- Add 'expired' to history_event_type enum
ALTER TYPE history_event_type ADD VALUE IF NOT EXISTS 'expired';

-- Add 'auto' to history_source_type enum
ALTER TYPE history_source_type ADD VALUE IF NOT EXISTS 'auto';

-- Add comment for documentation
COMMENT ON TYPE history_event_type IS 'Event types: cook, consume, waste, manual, fridge_add, fridge_remove, expired';
COMMENT ON TYPE history_source_type IS 'Source types: prepared_dish, recipe, fridge, manual, auto';

-- Create index for expired events (for analytics)
CREATE INDEX IF NOT EXISTS idx_history_events_expired 
ON history_events(user_id, created_at DESC) 
WHERE event_type = 'expired';

-- Add index for event_type filtering
CREATE INDEX IF NOT EXISTS idx_history_events_type_user 
ON history_events(event_type, user_id, created_at DESC);
