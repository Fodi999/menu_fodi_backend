-- Fix history_source_type enum values in production
-- Current values: {system, user, ai}
-- Target values: {prepared_dish, recipe, fridge, manual}

BEGIN;

-- Step 1: Create new enum with correct values
CREATE TYPE history_source_type_new AS ENUM ('prepared_dish', 'recipe', 'fridge', 'manual');

-- Step 2: Alter table column to use new enum (with USING clause to convert existing data)
-- Since old values (system/user/ai) don't map cleanly, we'll use 'manual' as default
ALTER TABLE history_events 
  ALTER COLUMN source_type TYPE history_source_type_new 
  USING 'manual'::history_source_type_new;

-- Step 3: Drop old enum
DROP TYPE history_source_type;

-- Step 4: Rename new enum to original name
ALTER TYPE history_source_type_new RENAME TO history_source_type;

-- Verify the change
SELECT enum_range(NULL::history_source_type) AS "New enum values";

COMMIT;
