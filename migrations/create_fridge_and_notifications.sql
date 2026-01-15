-- Create fridge_items table for tracking user's fridge inventory
CREATE TABLE IF NOT EXISTS fridge_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT NOT NULL REFERENCES "User"(id) ON DELETE CASCADE,
    ingredient_id TEXT NOT NULL REFERENCES "Ingredient"(id) ON DELETE RESTRICT,
    quantity NUMERIC(10,2) NOT NULL CHECK (quantity > 0),
    unit VARCHAR(20) NOT NULL,
    expires_at TIMESTAMPTZ,
    status VARCHAR(20) NOT NULL DEFAULT 'fresh' CHECK (status IN ('fresh', 'expired', 'discarded')),
    days_left INTEGER,
    price_total NUMERIC(10,2) DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for fridge_items
CREATE INDEX IF NOT EXISTS idx_fridge_items_user_id ON fridge_items(user_id);
CREATE INDEX IF NOT EXISTS idx_fridge_items_ingredient_id ON fridge_items(ingredient_id);
CREATE INDEX IF NOT EXISTS idx_fridge_items_status ON fridge_items(status);
CREATE INDEX IF NOT EXISTS idx_fridge_items_expires_at ON fridge_items(expires_at) WHERE expires_at IS NOT NULL;

-- Create notifications table for unified notification system
CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT NOT NULL REFERENCES "User"(id) ON DELETE CASCADE,
    type VARCHAR(20) NOT NULL CHECK (type IN ('system', 'order', 'user', 'fridge', 'ai', 'backup')),
    level VARCHAR(20) NOT NULL CHECK (level IN ('info', 'warning', 'critical')),
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    meta JSONB,
    read_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for notifications
CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_type ON notifications(type);
CREATE INDEX IF NOT EXISTS idx_notifications_level ON notifications(level);
CREATE INDEX IF NOT EXISTS idx_notifications_created_at ON notifications(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_notifications_unread ON notifications(user_id, read_at) WHERE read_at IS NULL;

-- Comments
COMMENT ON TABLE fridge_items IS 'User fridge inventory with expiration tracking';
COMMENT ON COLUMN fridge_items.days_left IS 'Calculated on backend: days until expiration';
COMMENT ON COLUMN fridge_items.price_total IS 'Total value for loss calculation';
COMMENT ON COLUMN fridge_items.status IS 'fresh=OK, expired=past date, discarded=thrown away';

COMMENT ON TABLE notifications IS 'Unified notification system for all events';
COMMENT ON COLUMN notifications.type IS 'system, order, user, fridge, ai, backup';
COMMENT ON COLUMN notifications.level IS 'info, warning, critical';
COMMENT ON COLUMN notifications.meta IS 'JSON metadata (e.g., fridgeItemId, orderId)';
