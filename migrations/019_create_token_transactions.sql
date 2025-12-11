-- Migration: Create token_transactions table for audit log
-- Description: Tracks all token operations (allocations, revokes, spending) for Treasury and Users

CREATE TABLE IF NOT EXISTS token_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    from_user_id UUID NULL,                              -- NULL = Treasury
    to_user_id UUID NULL,                                -- NULL = Burn/Revoke
    amount BIGINT NOT NULL CHECK (amount > 0),
    type VARCHAR(50) NOT NULL,                           -- WELCOME_BONUS, QUEST_REWARD, AI_USAGE, etc.
    description TEXT,
    metadata JSONB,                                       -- Additional data (quest_id, achievement_id, etc.)
    created_at TIMESTAMP DEFAULT NOW(),
    
    -- Indexes for faster queries
    CONSTRAINT token_transactions_from_user_fk FOREIGN KEY (from_user_id) REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT token_transactions_to_user_fk FOREIGN KEY (to_user_id) REFERENCES users(id) ON DELETE SET NULL
);

-- Indexes for performance
CREATE INDEX idx_token_transactions_from_user ON token_transactions(from_user_id);
CREATE INDEX idx_token_transactions_to_user ON token_transactions(to_user_id);
CREATE INDEX idx_token_transactions_type ON token_transactions(type);
CREATE INDEX idx_token_transactions_created_at ON token_transactions(created_at DESC);

-- Composite index for user-specific history
CREATE INDEX idx_token_transactions_user_type ON token_transactions(to_user_id, type, created_at DESC);

COMMENT ON TABLE token_transactions IS 'Complete audit log of all token operations';
COMMENT ON COLUMN token_transactions.from_user_id IS 'Source user ID (NULL = Treasury)';
COMMENT ON COLUMN token_transactions.to_user_id IS 'Destination user ID (NULL = Burn/Revoke)';
COMMENT ON COLUMN token_transactions.type IS 'Operation type: WELCOME_BONUS, QUEST_REWARD, AI_USAGE, MARKETPLACE_PURCHASE, etc.';
COMMENT ON COLUMN token_transactions.metadata IS 'Additional context data as JSON';
