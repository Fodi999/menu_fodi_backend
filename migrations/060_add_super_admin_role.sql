-- Migration: Add super_admin role
-- Date: 2026-01-04
-- Description: Adds super_admin role to Role enum and upgrades admin@example.com to super_admin

-- Add super_admin role to enum if not exists
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_enum WHERE enumlabel = 'super_admin' AND enumtypid = '"Role"'::regtype) THEN
        ALTER TYPE "Role" ADD VALUE 'super_admin';
    END IF;
END
$$;

-- Upgrade admin@example.com to super_admin
UPDATE "User"
SET role = 'super_admin'
WHERE email = 'admin@example.com';

-- Add comment
COMMENT ON COLUMN "User".role IS 'User role: home_chef (домашний повар), pro_chef (профессиональный повар), admin (администратор), super_admin (суперадмин с правами назначения ролей)';

-- Verify migration
DO $$
DECLARE
    super_admin_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO super_admin_count
    FROM "User"
    WHERE role = 'super_admin';
    
    IF super_admin_count = 0 THEN
        RAISE EXCEPTION 'Migration failed: No super_admin users found';
    END IF;
    
    RAISE NOTICE 'Migration successful: % super_admin(s) created', super_admin_count;
END
$$;
