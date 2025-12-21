-- Migration: Create recipe cooking history table
-- Description: Tracks when users cook recipes and what ingredients were used

-- Drop existing table if needed (for development)
DROP TABLE IF EXISTS "RecipeCookLog" CASCADE;

-- Create RecipeCookLog table
CREATE TABLE "RecipeCookLog" (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "userId" TEXT NOT NULL REFERENCES "User"(id) ON DELETE CASCADE,
    "recipeId" UUID NOT NULL REFERENCES "Recipe"(id) ON DELETE CASCADE,
    
    -- Cooking details
    "servingsMultiplier" DECIMAL(10, 2) NOT NULL DEFAULT 1.0, -- 1.0 = original servings, 2.0 = double, 0.5 = half
    "cookedAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    -- Economy snapshot at cooking time
    "usedValue" DECIMAL(10, 2) NOT NULL DEFAULT 0, -- PLN: cost of ingredients used from fridge
    "wasteRiskSaved" DECIMAL(10, 2) NOT NULL DEFAULT 0, -- PLN: value of expiring items used
    "totalRecipeCost" DECIMAL(10, 2) NOT NULL DEFAULT 0, -- PLN: full recipe cost
    
    -- Idempotency
    "idempotencyKey" VARCHAR(255), -- Unique key to prevent duplicate cooking
    
    -- Metadata
    "createdAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    "updatedAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    -- Indexes
    CONSTRAINT "RecipeCookLog_userId_recipeId_cookedAt_key" UNIQUE ("userId", "recipeId", "cookedAt"),
    CONSTRAINT "RecipeCookLog_idempotencyKey_key" UNIQUE ("idempotencyKey")
);

-- Create indexes for performance
CREATE INDEX "RecipeCookLog_userId_idx" ON "RecipeCookLog"("userId");
CREATE INDEX "RecipeCookLog_recipeId_idx" ON "RecipeCookLog"("recipeId");
CREATE INDEX "RecipeCookLog_cookedAt_idx" ON "RecipeCookLog"("cookedAt" DESC);

-- Create RecipeCookIngredients table (what was actually deducted)
CREATE TABLE "RecipeCookIngredient" (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "cookLogId" UUID NOT NULL REFERENCES "RecipeCookLog"(id) ON DELETE CASCADE,
    "ingredientId" TEXT NOT NULL REFERENCES "Ingredient"(id) ON DELETE RESTRICT,
    
    -- What was deducted
    "quantityUsed" DECIMAL(10, 3) NOT NULL, -- How much was used
    "unit" VARCHAR(50) NOT NULL, -- g, ml, pcs, etc.
    "pricePerUnit" DECIMAL(10, 4), -- Price at time of cooking
    "totalCost" DECIMAL(10, 2), -- quantityUsed * pricePerUnit
    
    -- Was it expiring?
    "wasExpiringSoon" BOOLEAN NOT NULL DEFAULT FALSE,
    
    -- Metadata
    "createdAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    -- Indexes
    CONSTRAINT "RecipeCookIngredient_cookLogId_ingredientId_key" UNIQUE ("cookLogId", "ingredientId")
);

CREATE INDEX "RecipeCookIngredient_cookLogId_idx" ON "RecipeCookIngredient"("cookLogId");
CREATE INDEX "RecipeCookIngredient_ingredientId_idx" ON "RecipeCookIngredient"("ingredientId");

-- Comments
COMMENT ON TABLE "RecipeCookLog" IS 'History of recipes cooked by users';
COMMENT ON TABLE "RecipeCookIngredient" IS 'Ingredients deducted from fridge when cooking';
COMMENT ON COLUMN "RecipeCookLog"."servingsMultiplier" IS 'Multiplier for servings (1.0 = original, 2.0 = double recipe)';
COMMENT ON COLUMN "RecipeCookLog"."idempotencyKey" IS 'Prevents duplicate cooking on double-click';
COMMENT ON COLUMN "RecipeCookIngredient"."wasExpiringSoon" IS 'TRUE if ingredient was expiring soon (waste prevention)';
