-- Migration 059: Add user status and activity tracking
-- This migration adds proper user status (active/blocked/pending) and activity tracking (last_login)

-- 1️⃣ Add account status column
ALTER TABLE "User"
ADD COLUMN status VARCHAR(20) NOT NULL DEFAULT 'active';

-- Create index for faster status filtering
CREATE INDEX idx_user_status ON "User"(status);

-- Add check constraint to ensure valid values
ALTER TABLE "User"
ADD CONSTRAINT check_user_status CHECK (status IN ('active', 'blocked', 'pending'));

-- 2️⃣ Add last_login column for activity tracking
ALTER TABLE "User"
ADD COLUMN last_login TIMESTAMP;

-- Create index for activity queries (active_today calculations)
CREATE INDEX idx_user_last_login ON "User"(last_login);

-- 3️⃣ Initialize last_login for existing users (set to account creation time)
UPDATE "User"
SET last_login = created_at
WHERE last_login IS NULL AND created_at IS NOT NULL;

-- Comment explaining the columns
COMMENT ON COLUMN "User".status IS 'Account status: active (normal user), blocked (banned by admin), pending (unverified)';
COMMENT ON COLUMN "User".last_login IS 'Last login timestamp for activity tracking. Used to calculate active_today statistics';
