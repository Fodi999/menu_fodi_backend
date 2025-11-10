-- Migration: Update admin user role
-- Description: Set role='admin' for admin@example.com user
-- Date: 2025-11-10

-- Update the admin user role
UPDATE "User" SET role = 'admin' WHERE email = 'admin@example.com';

-- Verify the update
-- SELECT id, email, name, role FROM "User" WHERE email = 'admin@example.com';
