-- Migration: Create Token Transactions table
-- Description: Creates table for tracking all token movements (earn/spend)
-- Date: 2025-12-11

-- ============================================
-- Create token_transactions table
-- ============================================
CREATE TABLE IF NOT EXISTS token_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT NOT NULL,
    amount BIGINT NOT NULL,
    type VARCHAR(50) NOT NULL, -- 'earn' or 'spend'
    reason VARCHAR(100) NOT NULL,
    description TEXT,
    balance_before BIGINT NOT NULL,
    balance_after BIGINT NOT NULL,
    metadata JSONB, -- Additional data (e.g., AI request details, task ID, etc.)
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Foreign key
    CONSTRAINT fk_transaction_user 
        FOREIGN KEY (user_id) 
        REFERENCES "User"(id) 
        ON DELETE CASCADE,
    
    -- Check constraints
    CONSTRAINT check_type 
        CHECK (type IN ('earn', 'spend')),
    
    CONSTRAINT check_amount_sign
        CHECK (
            (type = 'earn' AND amount > 0) OR 
            (type = 'spend' AND amount < 0)
        )
);

-- ============================================
-- Create indexes
-- ============================================

-- Index by user_id for user transaction history
CREATE INDEX IF NOT EXISTS idx_token_transactions_user_id 
    ON token_transactions(user_id);

-- Index by type for analytics (earn vs spend)
CREATE INDEX IF NOT EXISTS idx_token_transactions_type 
    ON token_transactions(type);

-- Index by reason for categorizing transactions
CREATE INDEX IF NOT EXISTS idx_token_transactions_reason 
    ON token_transactions(reason);

-- Index by created_at for time-based queries
CREATE INDEX IF NOT EXISTS idx_token_transactions_created_at 
    ON token_transactions(created_at DESC);

-- Composite index for user + time queries
CREATE INDEX IF NOT EXISTS idx_token_transactions_user_time 
    ON token_transactions(user_id, created_at DESC);

-- ============================================
-- Seed example transactions (for testing)
-- ============================================

-- Get first user for demo
DO $$
DECLARE
    demo_user_id TEXT;
BEGIN
    SELECT id INTO demo_user_id FROM "User" WHERE role = 'user' LIMIT 1;
    
    IF demo_user_id IS NOT NULL THEN
        -- Welcome bonus
        INSERT INTO token_transactions (user_id, amount, type, reason, description, balance_before, balance_after, metadata)
        VALUES (
            demo_user_id,
            100,
            'earn',
            'welcome_bonus',
            'New user welcome bonus',
            0,
            100,
            '{"source": "registration"}'::jsonb
        );
        
        -- Daily task completion
        INSERT INTO token_transactions (user_id, amount, type, reason, description, balance_before, balance_after, metadata)
        VALUES (
            demo_user_id,
            10,
            'earn',
            'daily_task',
            'Daily login streak',
            100,
            110,
            '{"task_id": "daily_login", "streak": 3}'::jsonb
        );
        
        -- AI request
        INSERT INTO token_transactions (user_id, amount, type, reason, description, balance_before, balance_after, metadata)
        VALUES (
            demo_user_id,
            -5,
            'spend',
            'ai_request',
            'AI Recipe Generation (Pro)',
            110,
            105,
            '{"complexity": "pro", "message_length": 250, "cost_breakdown": {"base": 3, "multiplier": 2}}'::jsonb
        );
    END IF;
END $$;

-- ============================================
-- Create view for transaction analytics
-- ============================================

CREATE OR REPLACE VIEW token_transaction_analytics AS
SELECT 
    user_id,
    COUNT(*) as total_transactions,
    COUNT(CASE WHEN type = 'earn' THEN 1 END) as earn_count,
    COUNT(CASE WHEN type = 'spend' THEN 1 END) as spend_count,
    SUM(CASE WHEN type = 'earn' THEN amount ELSE 0 END) as total_earned,
    ABS(SUM(CASE WHEN type = 'spend' THEN amount ELSE 0 END)) as total_spent,
    MAX(created_at) as last_transaction_at
FROM token_transactions
GROUP BY user_id;

-- ============================================
-- Create function to log transactions automatically
-- ============================================

CREATE OR REPLACE FUNCTION log_token_transaction(
    p_user_id TEXT,
    p_amount BIGINT,
    p_type VARCHAR(50),
    p_reason VARCHAR(100),
    p_description TEXT,
    p_metadata JSONB DEFAULT '{}'::jsonb
) RETURNS UUID AS $$
DECLARE
    v_balance_before BIGINT;
    v_balance_after BIGINT;
    v_transaction_id UUID;
BEGIN
    -- Get current balance
    SELECT balance INTO v_balance_before
    FROM token_bank
    WHERE user_id = p_user_id;
    
    IF v_balance_before IS NULL THEN
        RAISE EXCEPTION 'User token bank not found: %', p_user_id;
    END IF;
    
    -- Calculate new balance
    v_balance_after := v_balance_before + p_amount;
    
    -- Insert transaction record
    INSERT INTO token_transactions (
        user_id, amount, type, reason, description, 
        balance_before, balance_after, metadata
    )
    VALUES (
        p_user_id, p_amount, p_type, p_reason, p_description,
        v_balance_before, v_balance_after, p_metadata
    )
    RETURNING id INTO v_transaction_id;
    
    RETURN v_transaction_id;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- Example usage of log_token_transaction
-- ============================================

-- SELECT log_token_transaction(
--     'user-123',
--     -3,
--     'spend',
--     'ai_request',
--     'AI Simple Question',
--     '{"complexity": "basic", "cost": 3}'::jsonb
-- );

-- ============================================
-- Verification queries
-- ============================================

-- Check table structure
-- SELECT column_name, data_type, is_nullable 
-- FROM information_schema.columns 
-- WHERE table_name = 'token_transactions';

-- Check recent transactions
-- SELECT * FROM token_transactions ORDER BY created_at DESC LIMIT 10;

-- Check analytics view
-- SELECT * FROM token_transaction_analytics LIMIT 10;
