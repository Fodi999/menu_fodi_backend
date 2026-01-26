-- Migration: Add Role Model 2026 roles and statuses
-- Date: 2026-01-26
-- Description: Adds 'customer' and 'chef_staff' roles, adds 'suspended' status

-- =====================================================
-- STEP 1: Add new roles to Role ENUM
-- =====================================================

-- Add 'customer' role (if not exists)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_enum 
        WHERE enumlabel = 'customer' 
        AND enumtypid = (SELECT oid FROM pg_type WHERE typname = 'Role')
    ) THEN
        ALTER TYPE "Role" ADD VALUE 'customer';
        RAISE NOTICE 'Added role: customer';
    ELSE
        RAISE NOTICE 'Role customer already exists';
    END IF;
END$$;

-- Add 'chef_staff' role (if not exists)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_enum 
        WHERE enumlabel = 'chef_staff' 
        AND enumtypid = (SELECT oid FROM pg_type WHERE typname = 'Role')
    ) THEN
        ALTER TYPE "Role" ADD VALUE 'chef_staff';
        RAISE NOTICE 'Added role: chef_staff';
    ELSE
        RAISE NOTICE 'Role chef_staff already exists';
    END IF;
END$$;

-- =====================================================
-- STEP 2: Update CHECK constraint for status
-- =====================================================

-- Drop old constraint
ALTER TABLE "User" DROP CONSTRAINT IF EXISTS check_user_status;

-- Add new constraint with 'suspended' status
ALTER TABLE "User" ADD CONSTRAINT check_user_status 
CHECK (status IN ('pending', 'active', 'suspended', 'blocked'));

DO $$
BEGIN
    RAISE NOTICE 'Updated status constraint to include: pending, active, suspended, blocked';
END$$;

-- =====================================================
-- STEP 3: Migrate existing users from 'user' to 'customer'
-- =====================================================

-- Update all users with role 'user' to 'customer'
UPDATE "User" 
SET role = 'customer' 
WHERE role = 'user';

-- Log the migration
DO $$
DECLARE
    updated_count INTEGER;
BEGIN
    GET DIAGNOSTICS updated_count = ROW_COUNT;
    RAISE NOTICE 'Migrated % users from role=user to role=customer', updated_count;
END$$;

-- =====================================================
-- STEP 4: Display current role distribution
-- =====================================================

SELECT 
    role,
    COUNT(*) as user_count,
    status,
    COUNT(*) as status_count
FROM "User"
GROUP BY role, status
ORDER BY role, status;

-- =====================================================
-- STEP 5: Display all available roles
-- =====================================================

SELECT enumlabel as available_role 
FROM pg_enum 
WHERE enumtypid = (SELECT oid FROM pg_type WHERE typname = 'Role') 
ORDER BY enumsortorder;

-- =====================================================
-- SUCCESS MESSAGE
-- =====================================================

DO $$
BEGIN
    RAISE NOTICE '✅ Role Model 2026 migration completed successfully!';
    RAISE NOTICE 'Available roles: customer, home_chef, chef_staff, admin, super_admin';
    RAISE NOTICE 'Available statuses: pending, active, suspended, blocked';
END$$;
