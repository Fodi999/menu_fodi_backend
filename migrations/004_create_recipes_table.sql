-- Create Recipe table for user-generated recipes
-- This table stores recipes that users post to their profile and the main feed

CREATE TABLE IF NOT EXISTS "Recipe" (
    id VARCHAR(255) PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    image_url VARCHAR(500),
    author_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Foreign key to User table
    CONSTRAINT fk_recipe_author FOREIGN KEY (author_id) 
        REFERENCES "User"(id) ON DELETE CASCADE
);

-- Indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_recipe_author_id ON "Recipe"(author_id);
CREATE INDEX IF NOT EXISTS idx_recipe_created_at ON "Recipe"(created_at DESC);

-- Insert sample recipes for testing
INSERT INTO "Recipe" (id, title, description, image_url, author_id, created_at) VALUES
(
    'recipe-001',
    'Fresh Salmon Nigiri',
    'Autentyczne nigiri z łososiem - tradycyjna japońska kuchnia',
    'https://images.unsplash.com/photo-1579584425555-c3ce17fd4351?w=800',
    'ef03cd81-71fd-429f-bb5f-8be5c9172ca8', -- dima@example.com
    CURRENT_TIMESTAMP
),
(
    'recipe-002',
    'Spicy Tuna Maki Roll',
    'Pikantne maki z tuńczykiem i awokado',
    'https://images.unsplash.com/photo-1617196034183-421b4917c92d?w=800',
    'fba50be3-e3c5-4d73-8ed8-cfb6422f7034', -- anna@example.com
    CURRENT_TIMESTAMP - INTERVAL '2 hours'
),
(
    'recipe-003',
    'Rainbow Sushi Platter',
    'Kolorowy zestaw sushi z różnymi rodzajami ryb',
    'https://images.unsplash.com/photo-1611143669185-af224c5e3252?w=800',
    '407582be-59d5-4d21-873b-1a72d31b0d42', -- fodi85@gmail.ru
    CURRENT_TIMESTAMP - INTERVAL '5 hours'
),
(
    'recipe-004',
    'Dragon Roll',
    'Dragon roll z węgorzem i awokado - premium sushi',
    'https://images.unsplash.com/photo-1579584425555-c3ce17fd4351?w=800',
    'ef03cd81-71fd-429f-bb5f-8be5c9172ca8', -- dima@example.com
    CURRENT_TIMESTAMP - INTERVAL '1 day'
),
(
    'recipe-005',
    'Vegetarian Tempura Maki',
    'Wegańskie maki z tempurą warzywną',
    'https://images.unsplash.com/photo-1617196034796-ca11959d7f34?w=800',
    'fba50be3-e3c5-4d73-8ed8-cfb6422f7034', -- anna@example.com
    CURRENT_TIMESTAMP - INTERVAL '2 days'
);

-- Comment
COMMENT ON TABLE "Recipe" IS 'User-generated recipes for the main feed and user profiles';
COMMENT ON COLUMN "Recipe".author_id IS 'Foreign key to User table - the user who created this recipe';
