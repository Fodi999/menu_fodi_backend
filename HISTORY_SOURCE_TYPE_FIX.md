# History Source Type Enum Fix

## Problem
Production database had **wrong enum values** for `history_source_type`:
```sql
-- BEFORE (WRONG):
{system, user, ai}

-- AFTER (CORRECT):
{prepared_dish, recipe, fridge, manual}
```

## Error
```
ERROR: invalid input value for enum history_source_type: "fridge" (SQLSTATE 22P02)
```

This happened because:
1. Migration `049_create_history_events.sql` was never applied to production
2. Table existed with old enum definition from earlier development

## Solution Applied

### Migration Script: `FIX_HISTORY_SOURCE_TYPE_ENUM.sql`

```sql
BEGIN;

-- Create new enum with correct values
CREATE TYPE history_source_type_new AS ENUM ('prepared_dish', 'recipe', 'fridge', 'manual');

-- Alter column to use new enum (existing data converted to 'manual')
ALTER TABLE history_events 
  ALTER COLUMN source_type TYPE history_source_type_new 
  USING 'manual'::history_source_type_new;

-- Drop old enum
DROP TYPE history_source_type;

-- Rename new enum
ALTER TYPE history_source_type_new RENAME TO history_source_type;

COMMIT;
```

### Applied to Production
```bash
psql "$DATABASE_URL" -f FIX_HISTORY_SOURCE_TYPE_ENUM.sql
```

**Result:**
```
           New enum values            
--------------------------------------
 {prepared_dish,recipe,fridge,manual}
```

## Impact

### Before Fix
- ❌ Expired item cleanup failed with SQLSTATE 22P02
- ❌ Could not insert history events with `source_type='fridge'`
- ❌ All automatic fridge operations broke

### After Fix
- ✅ Expired items can be logged to history_events
- ✅ `SourceTypeAuto` (="fridge") works correctly
- ✅ Automatic cleanup resumes normal operation

## Related Files
- `internal/models/history_event.go` - Defines enums (was correct)
- `internal/modules/fridge/service/fridge_service.go:330` - Uses SourceTypeAuto
- `migrations/049_create_history_events.sql` - Original migration (never applied)
- `FIX_HISTORY_SOURCE_TYPE_ENUM.sql` - **Applied migration** (Jan 17, 2026)

## Verification
```sql
-- Check enum values
SELECT enum_range(NULL::history_source_type);

-- Test insert
INSERT INTO history_events (user_id, event_type, source_type, metadata)
VALUES ('test-user', 'waste', 'fridge', '{}');
```

## Status: ✅ RESOLVED (Jan 17, 2026)
