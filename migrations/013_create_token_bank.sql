-- Create token_bank table for tracking user token balance
CREATE TABLE IF NOT EXISTS token_bank (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL UNIQUE REFERENCES "User"(id) ON DELETE CASCADE,
  balance BIGINT NOT NULL DEFAULT 0,
  total_allocated BIGINT NOT NULL DEFAULT 0,
  total_used BIGINT NOT NULL DEFAULT 0,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create index for faster lookups
CREATE INDEX IF NOT EXISTS idx_token_bank_user_id ON token_bank(user_id);

-- Add comment for documentation
COMMENT ON TABLE token_bank IS 'Stores token balance information for each user';
COMMENT ON COLUMN token_bank.balance IS 'Current available tokens';
COMMENT ON COLUMN token_bank.total_allocated IS 'Total tokens allocated to user by admin';
COMMENT ON COLUMN token_bank.total_used IS 'Total tokens used by user';
