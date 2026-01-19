# History Events Entity Columns Fix

## Problem
Production database had **obsolete columns** in `history_events` table that don't exist in Go model:

```sql
-- OLD SCHEMA (PRODUCTION):
entity_type TEXT NOT NULL ❌  -- Missing in Go model
entity_id UUID            ❌  -- Missing in Go model
```

**Error:**
```
ERROR: null value in column "entity_type" of relation "history_events" 
violates not-null constraint (SQLSTATE 23502)
```

This happened because:
1. Early development used `entity_type` + `entity_id` pattern
2. Design changed to `source_type` + `source_id` pattern
3. Migration `049_create_history_events.sql` reflects new design
4. Production table never migrated from old to new schema

## Go Model (Correct Design)
```go
type HistoryEvent struct {
    ID         string            `gorm:"column:id;type:uuid;primaryKey"`
    UserID     string            `gorm:"column:user_id;type:text;not null"`
    EventType  HistoryEventType  `gorm:"column:event_type;type:history_event_type;not null"`
    SourceType HistorySourceType `gorm:"column:source_type;type:history_source_type;not null"`
    SourceID   *string           `gorm:"column:source_id;type:text"`
    Portions   *int              `gorm:"column:portions"`
    Metadata   datatypes.JSON    `gorm:"column:metadata;type:jsonb"`
    CreatedAt  time.Time         `gorm:"column:created_at;not null;default:NOW()"`
}
```

## Solution Applied

### Migration Script: `REMOVE_HISTORY_ENTITY_COLUMNS.sql`

```sql
BEGIN;

-- Remove obsolete columns from old schema
ALTER TABLE history_events DROP COLUMN IF EXISTS entity_type;
ALTER TABLE history_events DROP COLUMN IF EXISTS entity_id;

COMMIT;
```

### Applied to Production
```bash
psql "$DATABASE_URL" -f REMOVE_HISTORY_ENTITY_COLUMNS.sql
```

**Result:**
```
                      Table "public.history_events"
   Column    |           Type           | Nullable | Default 
-------------+--------------------------+----------+---------
 id          | uuid                     | not null | 
 user_id     | uuid                     | not null | 
 event_type  | history_event_type       | not null | 
 source_type | history_source_type      | not null | 
 metadata    | jsonb                    |          | 
 created_at  | timestamp with time zone | not null | now()
 source_id   | text                     |          | 
 portions    | integer                  |          | 
```

## Impact

### Before Fix
- ❌ `INSERT` failed: entity_type NOT NULL constraint
- ❌ Expired item cleanup broken
- ❌ Go model ↔ DB schema mismatch

### After Fix
- ✅ Schema matches Go model perfectly
- ✅ INSERT operations work without errors
- ✅ Expired items logged to history_events
- ✅ No unused columns cluttering the table

## Design Pattern Change

### Old Pattern (Removed)
```sql
entity_type TEXT  -- Generic: 'recipe', 'ingredient', 'dish'
entity_id UUID    -- ID of that entity
```

### New Pattern (Current)
```sql
source_type history_source_type  -- Enum: prepared_dish, recipe, fridge, manual
source_id TEXT                   -- Flexible: UUID or other identifier
```

**Why better:**
- Stronger typing with enum vs free text
- More specific semantics (source = what triggered event)
- Aligns with event-driven architecture

## Related Files
- `internal/models/history_event.go` - Go model (correct)
- `migrations/049_create_history_events.sql` - Target schema (correct)
- `REMOVE_HISTORY_ENTITY_COLUMNS.sql` - **Applied migration** (Jan 17, 2026)
- `FIX_HISTORY_SOURCE_TYPE_ENUM.sql` - Previous enum fix

## Verification
```sql
-- Check schema matches Go model
\d history_events

-- Test insert with new schema
INSERT INTO history_events (
    id, user_id, event_type, source_type, source_id, metadata
) VALUES (
    gen_random_uuid(),
    '407582be-59d5-4d21-873b-1a72d31b0d42',
    'waste',
    'fridge',
    'test-item-id',
    '{"test": true}'::jsonb
);
```

## Status: ✅ RESOLVED (Jan 17, 2026)
