-- Migration: Create users table
-- Date: 2025-11-03
-- Description: Creates users table for authentication

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'user',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create index on email for faster lookups
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- Create index on role for filtering
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);

-- Insert test users (with bcrypt hashed passwords for "password123")
INSERT INTO users (id, email, password, name, role, created_at) 
VALUES 
    ('ef03cd81-71fd-429f-bb5f-8be5c9172ca8', 'dima@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'Dima Fomin', 'user', '2025-11-01T10:00:00Z'),
    ('fba50be3-e3c5-4d73-8ed8-cfb6422f7034', 'anna@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'Anna Kowalska', 'user', '2025-11-01T10:00:00Z')
ON CONFLICT (id) DO NOTHING;

-- Add comment
COMMENT ON TABLE users IS 'Main authentication table for user accounts';
