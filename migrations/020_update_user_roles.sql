-- Migration: Update user roles for role-based ingredient management
-- Author: AI Assistant
-- Date: 2025-12-14
-- Description: Changes user.role values from 'user'/'admin' to 'home_chef'/'pro_chef'/'admin'

-- Add new role values to enum (if using enum type)
-- If role is TEXT (not enum), this is just documentation

-- Update existing users: 'user' → 'home_chef'
UPDATE "User"
SET role = 'home_chef'
WHERE role = 'user';

-- Existing admins remain 'admin'
-- No changes needed for admin users

-- Set default for new users
ALTER TABLE "User"
ALTER COLUMN role SET DEFAULT 'home_chef';

-- Add comment for clarity
COMMENT ON COLUMN "User".role IS 'User role: home_chef (домашний повар), pro_chef (ресторан/бизнес), admin';
