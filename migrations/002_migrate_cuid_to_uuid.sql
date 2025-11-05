-- Migration: Convert CUID user IDs to UUID
-- Date: 2025-11-04
-- Description: Migrates old CUID format IDs to UUID format for consistency

-- Step 1: Create mapping table for old ID -> new ID
CREATE TEMP TABLE user_id_mapping AS
SELECT 
    id as old_id,
    gen_random_uuid() as new_id
FROM "User"
WHERE id !~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$';

-- Step 2: Update foreign keys in related tables
-- UserProfile
UPDATE "UserProfile" up
SET user_id = m.new_id::text
FROM user_id_mapping m
WHERE up.user_id = m.old_id;

-- PersonalRecipe
UPDATE "PersonalRecipe" pr
SET user_id = m.new_id::text
FROM user_id_mapping m
WHERE pr.user_id = m.old_id;

-- Certificate
UPDATE "Certificate" c
SET user_id = m.new_id::text
FROM user_id_mapping m
WHERE c.user_id = m.old_id;

-- UserProgress
UPDATE "UserProgress" upr
SET user_id = m.new_id::text
FROM user_id_mapping m
WHERE upr.user_id = m.old_id;

-- WalletTransaction
UPDATE "WalletTransaction" wt
SET user_id = m.new_id::text
FROM user_id_mapping m
WHERE wt.user_id = m.old_id;

-- MarketPurchase (buyer_id and seller_id)
UPDATE "MarketPurchase" mp
SET buyer_id = m.new_id::text
FROM user_id_mapping m
WHERE mp.buyer_id = m.old_id;

UPDATE "MarketPurchase" mp
SET seller_id = m.new_id::text
FROM user_id_mapping m
WHERE mp.seller_id = m.old_id;

-- RecipePurchase (buyer_id and seller_id)
UPDATE "RecipePurchase" rp
SET buyer_id = m.new_id::text
FROM user_id_mapping m
WHERE rp.buyer_id = m.old_id;

UPDATE "RecipePurchase" rp
SET seller_id = m.new_id::text
FROM user_id_mapping m
WHERE rp.seller_id = m.old_id;

-- UserAchievement
UPDATE "UserAchievement" ua
SET user_id = m.new_id::text
FROM user_id_mapping m
WHERE ua.user_id = m.old_id;

-- MentorSession
UPDATE "MentorSession" ms
SET user_id = m.new_id::text
FROM user_id_mapping m
WHERE ms.user_id = m.old_id;

-- MentorMessage
UPDATE "MentorMessage" mm
SET user_id = m.new_id::text
FROM user_id_mapping m
WHERE mm.user_id = m.old_id;

-- UserQuiz
UPDATE "UserQuiz" uq
SET user_id = m.new_id::text
FROM user_id_mapping m
WHERE uq.user_id = m.old_id;

-- Business (owner_id)
UPDATE "Business" b
SET owner_id = m.new_id::text
FROM user_id_mapping m
WHERE b.owner_id = m.old_id;

-- BusinessSubscription
UPDATE "BusinessSubscription" bs
SET user_id = m.new_id::text
FROM user_id_mapping m
WHERE bs.user_id = m.old_id;

-- Step 3: Update User table itself
UPDATE "User" u
SET id = m.new_id::text
FROM user_id_mapping m
WHERE u.id = m.old_id;

-- Step 4: Verify migration
SELECT 
    COUNT(*) as total_users,
    COUNT(CASE WHEN id ~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$' THEN 1 END) as uuid_count,
    COUNT(CASE WHEN id !~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$' THEN 1 END) as cuid_count
FROM "User";

COMMENT ON COLUMN "User".id IS 'UUID primary key (migrated from CUID on 2025-11-04)';
