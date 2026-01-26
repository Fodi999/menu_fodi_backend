-- Migration: Create dishes table
-- Date: 2026-01-26
-- Description: Creates commercial dish cards for marketplace (based on recipes)

-- =====================================================
-- STEP 1: Create dishes table
-- =====================================================

CREATE TABLE IF NOT EXISTS dishes (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    recipe_id       UUID NOT NULL,  -- FK to Recipe.id
    
    -- Content
    title           VARCHAR(255) NOT NULL,
    description     TEXT,
    image_url       TEXT,
    
    -- Finance (core business logic)
    cost            DECIMAL(10,2) NOT NULL,  -- Cost to make (snapshot from fridge)
    price           DECIMAL(10,2) NOT NULL,  -- Selling price
    margin          DECIMAL(5,2) NOT NULL,   -- Margin percentage (0-100)
    
    -- Status lifecycle
    status          VARCHAR(20) NOT NULL DEFAULT 'draft',
    is_available    BOOLEAN NOT NULL DEFAULT true,
    
    -- Metadata
    created_by      TEXT NOT NULL,  -- FK to User.id (TEXT type)
    approved_by     TEXT,           -- FK to User.id (TEXT type)
    approved_at     TIMESTAMP,
    
    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT dishes_status_check CHECK (status IN ('draft', 'approved', 'published')),
    CONSTRAINT dishes_cost_positive CHECK (cost >= 0),
    CONSTRAINT dishes_price_positive CHECK (price >= 0),
    CONSTRAINT dishes_margin_range CHECK (margin >= 0 AND margin <= 100)
);

-- =====================================================
-- STEP 2: Create indexes for performance
-- =====================================================

-- Index for filtering by recipe
CREATE INDEX IF NOT EXISTS idx_dishes_recipe_id ON dishes(recipe_id);

-- Index for filtering by status
CREATE INDEX IF NOT EXISTS idx_dishes_status ON dishes(status);

-- Index for filtering available dishes
CREATE INDEX IF NOT EXISTS idx_dishes_is_available ON dishes(is_available);

-- Index for filtering by creator
CREATE INDEX IF NOT EXISTS idx_dishes_created_by ON dishes(created_by);

-- Composite index for marketplace queries (published + available)
CREATE INDEX IF NOT EXISTS idx_dishes_marketplace ON dishes(status, is_available) 
WHERE status = 'published' AND is_available = true;

-- =====================================================
-- STEP 3: Add foreign keys (if tables exist)
-- =====================================================

-- Foreign key to Recipe (UUID)
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'Recipe') THEN
        ALTER TABLE dishes 
        ADD CONSTRAINT fk_dishes_recipe 
        FOREIGN KEY (recipe_id) 
        REFERENCES "Recipe"(id) 
        ON DELETE CASCADE;
        
        RAISE NOTICE 'Foreign key to Recipe created';
    ELSE
        RAISE NOTICE 'Table Recipe not found, skipping FK creation';
    END IF;
END$$;

-- Foreign key to User (creator)
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'User') THEN
        ALTER TABLE dishes 
        ADD CONSTRAINT fk_dishes_created_by 
        FOREIGN KEY (created_by) 
        REFERENCES "User"(id) 
        ON DELETE CASCADE;
        
        RAISE NOTICE 'Foreign key to User (created_by) created';
    ELSE
        RAISE NOTICE 'Table User not found, skipping FK creation';
    END IF;
END$$;

-- Foreign key to User (approver)
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'User') THEN
        ALTER TABLE dishes 
        ADD CONSTRAINT fk_dishes_approved_by 
        FOREIGN KEY (approved_by) 
        REFERENCES "User"(id) 
        ON DELETE SET NULL;
        
        RAISE NOTICE 'Foreign key to User (approved_by) created';
    ELSE
        RAISE NOTICE 'Table User not found, skipping FK creation';
    END IF;
END$$;

-- =====================================================
-- STEP 4: Display table info
-- =====================================================

SELECT 
    column_name,
    data_type,
    is_nullable,
    column_default
FROM information_schema.columns
WHERE table_name = 'dishes'
ORDER BY ordinal_position;

-- Display indexes
SELECT indexname, indexdef 
FROM pg_indexes 
WHERE tablename = 'dishes';

-- =====================================================
-- SUCCESS MESSAGE
-- =====================================================

DO $$
BEGIN
    RAISE NOTICE '✅ Dishes table created successfully!';
    RAISE NOTICE 'Columns: id, recipe_id, title, description, image_url';
    RAISE NOTICE 'Finance: cost, price, margin';
    RAISE NOTICE 'Status: status (draft/approved/published), is_available';
    RAISE NOTICE 'Metadata: created_by, approved_by, approved_at';
END$$;
