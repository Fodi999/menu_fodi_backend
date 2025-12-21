-- Migration 035: Create Recipe Catalog
-- Created: 2024-12-21
-- Purpose: Create structured catalog of real recipes for matching and filtering

-- ==============================================
-- 1. RECIPES TABLE (main catalog)
-- ==============================================
CREATE TABLE IF NOT EXISTS "Recipe" (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Names
    "canonicalName" VARCHAR(255) NOT NULL UNIQUE, -- English name for system (e.g., "Pierogi Ruskie")
    "localName" VARCHAR(255) NOT NULL,             -- Local name (e.g., "Pierogi ruskie")
    
    -- Origin
    country VARCHAR(100) NOT NULL,                 -- Poland, Italy, France, etc.
    region VARCHAR(100),                           -- Małopolska, Toscana, etc.
    
    -- Classification
    category VARCHAR(50) NOT NULL,                 -- appetizer, main, dessert, soup, salad
    difficulty VARCHAR(20) NOT NULL,               -- easy, medium, hard
    
    -- Metrics
    "timeMinutes" INTEGER NOT NULL,                -- Total cooking time
    servings INTEGER NOT NULL DEFAULT 4,           -- Number of servings
    
    -- Steps (stored as JSONB array)
    steps JSONB NOT NULL DEFAULT '[]'::jsonb,      -- [{step: 1, instruction: "..."}]
    
    -- Nutrition profile
    "nutritionProfile" JSONB DEFAULT '{}'::jsonb,  -- {type: "balanced|high-protein|low-carb", calories: 450}
    
    -- Source reference
    source JSONB NOT NULL,                         -- {type: "cookbook|website|traditional", reference: "URL or title"}
    
    -- Metadata
    "createdAt" TIMESTAMP NOT NULL DEFAULT NOW(),
    "updatedAt" TIMESTAMP NOT NULL DEFAULT NOW(),
    
    -- Indexes for filtering
    CONSTRAINT recipe_difficulty_check CHECK (difficulty IN ('easy', 'medium', 'hard')),
    CONSTRAINT recipe_category_check CHECK (category IN ('appetizer', 'soup', 'salad', 'main', 'side', 'dessert', 'beverage'))
);

CREATE INDEX idx_recipe_country ON "Recipe"(country);
CREATE INDEX idx_recipe_category ON "Recipe"(category);
CREATE INDEX idx_recipe_difficulty ON "Recipe"(difficulty);
CREATE INDEX idx_recipe_time ON "Recipe"("timeMinutes");

-- ==============================================
-- 2. RECIPE INGREDIENTS (junction table)
-- ==============================================
CREATE TABLE IF NOT EXISTS "RecipeIngredient" (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "recipeId" UUID NOT NULL REFERENCES "Recipe"(id) ON DELETE CASCADE,
    "ingredientId" TEXT NOT NULL REFERENCES "Ingredient"(id) ON DELETE RESTRICT,
    
    -- Ingredient details
    "ingredientKey" VARCHAR(255) NOT NULL,         -- Normalized key for matching (e.g., "potato")
    quantity DECIMAL(10,2) NOT NULL,               -- Amount needed
    unit VARCHAR(50) NOT NULL,                     -- g, ml, pcs, etc.
    optional BOOLEAN DEFAULT FALSE,                -- Can be omitted?
    
    -- Order in recipe
    "sortOrder" INTEGER NOT NULL DEFAULT 0,
    
    "createdAt" TIMESTAMP NOT NULL DEFAULT NOW(),
    
    CONSTRAINT recipe_ingredient_unique UNIQUE ("recipeId", "ingredientId")
);

CREATE INDEX idx_recipe_ingredient_recipe ON "RecipeIngredient"("recipeId");
CREATE INDEX idx_recipe_ingredient_key ON "RecipeIngredient"("ingredientKey");

-- ==============================================
-- 3. ALLERGENS (enum + junction)
-- ==============================================
CREATE TABLE IF NOT EXISTS "Allergen" (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) NOT NULL UNIQUE,              -- gluten, lactose, nuts, eggs, fish, shellfish, soy, celery, mustard, sesame
    "displayName" VARCHAR(100) NOT NULL,           -- Human-readable name
    icon VARCHAR(50),                              -- Emoji or icon name
    "createdAt" TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS "RecipeAllergen" (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "recipeId" UUID NOT NULL REFERENCES "Recipe"(id) ON DELETE CASCADE,
    "allergenId" UUID NOT NULL REFERENCES "Allergen"(id) ON DELETE CASCADE,
    
    CONSTRAINT recipe_allergen_unique UNIQUE ("recipeId", "allergenId")
);

CREATE INDEX idx_recipe_allergen_recipe ON "RecipeAllergen"("recipeId");
CREATE INDEX idx_recipe_allergen_allergen ON "RecipeAllergen"("allergenId");

-- ==============================================
-- 4. DIET TAGS (junction)
-- ==============================================
CREATE TABLE IF NOT EXISTS "DietTag" (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) NOT NULL UNIQUE,              -- vegetarian, vegan, keto, paleo, gluten-free, dairy-free, low-carb, high-protein
    "displayName" VARCHAR(100) NOT NULL,
    description TEXT,
    "createdAt" TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS "RecipeDietTag" (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "recipeId" UUID NOT NULL REFERENCES "Recipe"(id) ON DELETE CASCADE,
    "dietTagId" UUID NOT NULL REFERENCES "DietTag"(id) ON DELETE CASCADE,
    
    CONSTRAINT recipe_diet_tag_unique UNIQUE ("recipeId", "dietTagId")
);

CREATE INDEX idx_recipe_diet_tag_recipe ON "RecipeDietTag"("recipeId");
CREATE INDEX idx_recipe_diet_tag_tag ON "RecipeDietTag"("dietTagId");

-- ==============================================
-- 5. SEED COMMON ALLERGENS
-- ==============================================
INSERT INTO "Allergen" (name, "displayName", icon) VALUES
    ('gluten', 'Gluten', '🌾'),
    ('lactose', 'Lactose (Dairy)', '🥛'),
    ('eggs', 'Eggs', '🥚'),
    ('fish', 'Fish', '🐟'),
    ('shellfish', 'Shellfish', '🦐'),
    ('nuts', 'Tree Nuts', '🌰'),
    ('peanuts', 'Peanuts', '🥜'),
    ('soy', 'Soy', '🫘'),
    ('celery', 'Celery', '🥬'),
    ('mustard', 'Mustard', '🌭'),
    ('sesame', 'Sesame', '🌱'),
    ('sulfites', 'Sulfites', '🍷'),
    ('lupin', 'Lupin', '🌾'),
    ('molluscs', 'Molluscs', '🐚')
ON CONFLICT (name) DO NOTHING;

-- ==============================================
-- 6. SEED COMMON DIET TAGS
-- ==============================================
INSERT INTO "DietTag" (name, "displayName", description) VALUES
    ('vegetarian', 'Vegetarian', 'No meat or fish'),
    ('vegan', 'Vegan', 'No animal products'),
    ('gluten-free', 'Gluten-Free', 'No wheat, barley, rye'),
    ('dairy-free', 'Dairy-Free', 'No milk products'),
    ('keto', 'Keto', 'Low carb, high fat'),
    ('paleo', 'Paleo', 'No grains, legumes, dairy'),
    ('low-carb', 'Low-Carb', 'Reduced carbohydrates'),
    ('high-protein', 'High-Protein', 'Protein-focused meals'),
    ('pescatarian', 'Pescatarian', 'Fish but no meat'),
    ('halal', 'Halal', 'Islamic dietary laws'),
    ('kosher', 'Kosher', 'Jewish dietary laws')
ON CONFLICT (name) DO NOTHING;

-- ==============================================
-- 7. HELPER VIEWS
-- ==============================================

-- View: Recipe with ingredient count
CREATE OR REPLACE VIEW "RecipeWithStats" AS
SELECT 
    r.*,
    COUNT(DISTINCT ri."ingredientId") as "ingredientCount",
    COUNT(DISTINCT ra."allergenId") as "allergenCount",
    COUNT(DISTINCT rdt."dietTagId") as "dietTagCount"
FROM "Recipe" r
LEFT JOIN "RecipeIngredient" ri ON ri."recipeId" = r.id
LEFT JOIN "RecipeAllergen" ra ON ra."recipeId" = r.id
LEFT JOIN "RecipeDietTag" rdt ON rdt."recipeId" = r.id
GROUP BY r.id;

-- ==============================================
-- NOTES:
-- ==============================================
-- ✅ canonicalName - unique English name for system reference
-- ✅ localName - native language name for display
-- ✅ ingredientKey - normalized for matching (lowercase, singular)
-- ✅ optional ingredients - marked for flexible matching
-- ✅ steps - JSONB array for structured cooking instructions
-- ✅ allergens - pre-seeded with EU 14 allergens
-- ✅ dietTags - common diet restrictions
-- ✅ Indexes on country, category, difficulty, time for fast filtering
