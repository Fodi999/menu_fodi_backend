-- Migration: Create UserFridgeItem table for home chefs
-- Author: AI Assistant
-- Date: 2025-12-14
-- Description: Personal fridge management for home_chef users

CREATE TABLE "UserFridgeItem" (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "userId" UUID NOT NULL REFERENCES "User"(id) ON DELETE CASCADE,
    "ingredientId" UUID REFERENCES "Ingredient"(id) ON DELETE SET NULL,
    name TEXT NOT NULL,
    quantity TEXT,
    price DECIMAL(10,2),
    "purchasedAt" TIMESTAMPTZ,
    "expiryDate" TIMESTAMPTZ,
    "createdAt" TIMESTAMPTZ DEFAULT NOW(),
    "updatedAt" TIMESTAMPTZ DEFAULT NOW(),
    "deletedAt" TIMESTAMPTZ -- Soft delete для аналитики
);

-- Indexes for performance
CREATE INDEX idx_user_fridge_user ON "UserFridgeItem"("userId");
CREATE INDEX idx_user_fridge_ingredient ON "UserFridgeItem"("ingredientId");
CREATE INDEX idx_user_fridge_expiry ON "UserFridgeItem"("expiryDate");
CREATE INDEX idx_user_fridge_created ON "UserFridgeItem"("createdAt" DESC);
CREATE INDEX idx_user_fridge_deleted ON "UserFridgeItem"("deletedAt"); -- Для soft delete

-- Comment
COMMENT ON TABLE "UserFridgeItem" IS 'Personal fridge for home_chef users - simple ingredient tracking with soft delete support';
