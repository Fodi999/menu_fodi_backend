-- ============================================================================
-- Migration: Cleanup Legacy 'users' Table and Fix Foreign Keys
-- Date: 2026-01-24
-- ============================================================================
-- Purpose:
--   1. Fix token_transactions FK to reference "User" instead of users
--   2. Rename legacy users table (safety measure)
--   3. Drop legacy users table after verification
-- ============================================================================

BEGIN;

-- ============================================================================
-- STEP 1: Fix token_transactions Foreign Keys
-- ============================================================================

-- Check current FK constraints
SELECT 
    conname AS constraint_name,
    conrelid::regclass AS table_from,
    confrelid::regclass AS table_to
FROM pg_constraint 
WHERE conrelid = 'token_transactions'::regclass 
  AND contype = 'f';

-- Drop old foreign keys referencing 'users'
ALTER TABLE token_transactions
DROP CONSTRAINT IF EXISTS token_transactions_from_user_fk;

ALTER TABLE token_transactions
DROP CONSTRAINT IF EXISTS token_transactions_to_user_fk;

-- Add new foreign keys referencing "User"
ALTER TABLE token_transactions
ADD CONSTRAINT token_transactions_from_user_fk
FOREIGN KEY (from_user_id) REFERENCES "User"(id) ON DELETE SET NULL;

ALTER TABLE token_transactions
ADD CONSTRAINT token_transactions_to_user_fk
FOREIGN KEY (to_user_id) REFERENCES "User"(id) ON DELETE SET NULL;

-- Verify new constraints
SELECT 
    conname AS constraint_name,
    conrelid::regclass AS table_from,
    confrelid::regclass AS table_to
FROM pg_constraint 
WHERE conrelid = 'token_transactions'::regclass 
  AND contype = 'f';

-- ============================================================================
-- STEP 2: Rename Legacy Table (Safety Measure)
-- ============================================================================

-- Rename 'users' to 'users__to_delete' (keeps data as backup)
ALTER TABLE users RENAME TO users__to_delete;

-- Verify table renamed
SELECT tablename 
FROM pg_tables 
WHERE schemaname = 'public' 
  AND tablename LIKE 'users%';

COMMIT;

-- ============================================================================
-- VERIFICATION QUERIES (Run after migration)
-- ============================================================================

-- 1. Check User table count (should be 40+)
SELECT COUNT(*) as user_table_count FROM "User";

-- 2. Check legacy table is renamed
SELECT COUNT(*) as legacy_table_count FROM users__to_delete;

-- 3. Verify FK constraints point to "User"
SELECT 
    tc.table_name,
    kcu.column_name,
    ccu.table_name AS foreign_table_name,
    ccu.column_name AS foreign_column_name
FROM information_schema.table_constraints AS tc
JOIN information_schema.key_column_usage AS kcu
  ON tc.constraint_name = kcu.constraint_name
JOIN information_schema.constraint_column_usage AS ccu
  ON ccu.constraint_name = tc.constraint_name
WHERE tc.constraint_type = 'FOREIGN KEY'
  AND tc.table_name = 'token_transactions';

-- ============================================================================
-- STEP 3: Final Cleanup (Run ONLY after full verification)
-- ============================================================================

-- ⚠️ DANGER ZONE - Only run after confirming everything works!
-- Uncomment to execute:

-- DROP TABLE IF EXISTS users__to_delete CASCADE;

-- ============================================================================
-- ROLLBACK (If Something Goes Wrong)
-- ============================================================================

-- If you need to rollback:
-- BEGIN;
-- ALTER TABLE users__to_delete RENAME TO users;
-- ALTER TABLE token_transactions DROP CONSTRAINT token_transactions_from_user_fk;
-- ALTER TABLE token_transactions DROP CONSTRAINT token_transactions_to_user_fk;
-- ALTER TABLE token_transactions 
--   ADD CONSTRAINT token_transactions_from_user_fk 
--   FOREIGN KEY (from_user_id) REFERENCES users(id) ON DELETE SET NULL;
-- ALTER TABLE token_transactions 
--   ADD CONSTRAINT token_transactions_to_user_fk 
--   FOREIGN KEY (to_user_id) REFERENCES users(id) ON DELETE SET NULL;
-- COMMIT;
