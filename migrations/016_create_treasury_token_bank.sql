-- Migration: Create Treasury Token Bank
-- Date: 2025-11-12
-- Description: Creates a special treasury token bank with 1 billion tokens for the platform

-- Step 1: Create Treasury system user (if not exists)
INSERT INTO "User" (id, email, name, password, role, "createdAt")
VALUES (
    'TREASURY',
    'treasury@system.internal',
    'Platform Treasury',
    '$2a$10$TREASURY_SYSTEM_ACCOUNT_NOT_FOR_LOGIN',  -- Invalid bcrypt hash to prevent login
    'admin',
    NOW()
)
ON CONFLICT (id) DO NOTHING;

-- Step 2: Create Treasury token bank with fixed supply of 1,000,000,000 tokens
INSERT INTO token_bank (id, user_id, balance, total_allocated, total_used, created_at, updated_at)
VALUES (
    gen_random_uuid(),
    'TREASURY',
    1000000000,  -- 1 billion tokens initial supply
    1000000000,  -- Total allocated = initial supply (from platform creation)
    0,           -- No tokens distributed yet
    NOW(),
    NOW()
)
ON CONFLICT (user_id) DO UPDATE SET
    balance = 1000000000,
    total_allocated = 1000000000,
    updated_at = NOW();

-- Add comment for documentation
COMMENT ON COLUMN token_bank.user_id IS 'User ID or special identifier (TREASURY for platform treasury system account)';

-- Verify treasury creation
DO $$
DECLARE
    treasury_balance BIGINT;
    treasury_allocated BIGINT;
BEGIN
    SELECT balance, total_allocated INTO treasury_balance, treasury_allocated 
    FROM token_bank 
    WHERE user_id = 'TREASURY';
    
    RAISE NOTICE '✅ Treasury created successfully!';
    RAISE NOTICE '   Balance: % tokens', treasury_balance;
    RAISE NOTICE '   Total Supply: % tokens', treasury_allocated;
    RAISE NOTICE '   Available to allocate: % tokens', treasury_balance;
END $$;
