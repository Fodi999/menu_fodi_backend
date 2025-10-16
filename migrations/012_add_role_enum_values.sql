-- Migration: Add business_owner and investor roles to Role enum
-- Date: 2025-10-16

-- Add business_owner role if not exists
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_enum WHERE enumlabel = 'business_owner' AND enumtypid = (SELECT oid FROM pg_type WHERE typname = 'Role')) THEN
        ALTER TYPE "Role" ADD VALUE 'business_owner';
    END IF;
END$$;

-- Add investor role if not exists
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_enum WHERE enumlabel = 'investor' AND enumtypid = (SELECT oid FROM pg_type WHERE typname = 'Role')) THEN
        ALTER TYPE "Role" ADD VALUE 'investor';
    END IF;
END$$;

-- Verify enum values
SELECT enumlabel as role_value FROM pg_enum WHERE enumtypid = (SELECT oid FROM pg_type WHERE typname = 'Role') ORDER BY enumsortorder;
