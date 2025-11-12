-- Migration: Create token_bank table
-- Date: 2025-11-12
-- Description: Creates token_bank table for tracking user token balances

-- Create token_bank table with TEXT for user_id (matching User.id type)
CREATE TABLE IF NOT EXISTS token_bank (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id TEXT NOT NULL UNIQUE REFERENCES "User"(id) ON DELETE CASCADE,
  balance BIGINT NOT NULL DEFAULT 0,
  total_allocated BIGINT NOT NULL DEFAULT 0,
  total_used BIGINT NOT NULL DEFAULT 0,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create index for faster lookups
CREATE INDEX IF NOT EXISTS idx_token_bank_user_id ON token_bank(user_id);

-- Initialize token banks for all existing users
INSERT INTO token_bank (user_id, balance, total_allocated, total_used, created_at, updated_at)
SELECT id, 0, 0, 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
FROM "User"
ON CONFLICT (user_id) DO NOTHING;

-- Add comment for documentation
COMMENT ON TABLE token_bank IS 'Stores token balance information for each user';
COMMENT ON COLUMN token_bank.balance IS 'Current available tokens';
COMMENT ON COLUMN token_bank.total_allocated IS 'Total tokens allocated to user by admin';
COMMENT ON COLUMN token_bank.total_used IS 'Total tokens used by user';
