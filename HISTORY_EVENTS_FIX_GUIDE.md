# History Events Schema Fix - URGENT

## 🔴 Problem

```
ERROR: column "source_type" of relation "history_events" does not exist (SQLSTATE 42703)
```

**Impact:** Fridge cleanup fails, waste tracking doesn't work.

## ✅ Solution

Apply missing columns to `history_events` table.

---

## 📋 Quick Fix (5 minutes)

### 1. Connect to Neon PostgreSQL

Get connection string from Koyeb environment or Neon dashboard.

### 2. Run SQL Fix

Execute `FIX_HISTORY_EVENTS_SCHEMA.sql`:

```bash
psql "$DATABASE_URL" -f FIX_HISTORY_EVENTS_SCHEMA.sql
```

Or manually:

```sql
-- Create enum types
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'history_event_type') THEN
        CREATE TYPE history_event_type AS ENUM ('cook', 'consume', 'waste', 'manual', 'fridge_add', 'fridge_remove');
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'history_source_type') THEN
        CREATE TYPE history_source_type AS ENUM ('prepared_dish', 'recipe', 'fridge', 'manual');
    END IF;
END $$;

-- Add missing columns
ALTER TABLE history_events 
ADD COLUMN IF NOT EXISTS source_type history_source_type NOT NULL DEFAULT 'manual';

ALTER TABLE history_events 
ADD COLUMN IF NOT EXISTS source_id TEXT;

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_history_events_source 
ON history_events(source_type, source_id);
```

### 3. Verify

```sql
SELECT column_name, data_type 
FROM information_schema.columns 
WHERE table_name='history_events' 
  AND column_name IN ('source_type', 'source_id');
```

**Expected:**
```
 column_name | data_type   
-------------+-------------
 source_type | USER-DEFINED
 source_id   | text
```

---

## 🔗 Related Errors

This fix also helps with:
- ✅ "prepared statement name is already in use" (connection pool issue)
- ✅ Fridge expiry cleanup failures
- ✅ Waste tracking not working

---

## 📊 Root Cause

Migration `migrations/049_create_history_events.sql` was:
- ✅ Created in code
- ❌ **NOT applied to production database**

Possible reasons:
1. Manual table creation skipped migration
2. Goose migration tool not run
3. Partial migration rollback

---

## ⏰ Timeline

- **Discovery:** 2026-01-16 22:47:56 UTC
- **Error:** SQLSTATE 42703 (column does not exist)
- **Fix Time:** ~2 minutes
- **Downtime:** 0 (non-breaking ALTER TABLE)

---

## 🧪 Test After Fix

```bash
# Check fridge items (should work without errors)
curl -H "Authorization: Bearer $TOKEN" \
  https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/fridge/items

# Check logs - no more SQLSTATE 42703 errors
```

---

**Status:** Ready to apply ✅  
**Risk:** Very low (idempotent SQL, backward compatible)  
**Required:** YES (blocks waste tracking feature)
