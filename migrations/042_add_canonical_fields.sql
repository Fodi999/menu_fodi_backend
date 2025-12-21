-- Migration: Add canonical fields to Recipe and RecipeIngredient
-- Purpose: Align database structure with canonical format
-- Date: 2025-12-21
--
-- Adds:
-- 1. Recipe.description - Recipe description for UI
-- 2. Recipe.imageUrl - Recipe image URL
-- 3. RecipeIngredient.groupName - Ingredient grouping ("baza", "sos", "dodatki")
--
-- All fields are nullable for backward compatibility

-- Add description and imageUrl to Recipe
ALTER TABLE "Recipe" 
ADD COLUMN IF NOT EXISTS description TEXT,
ADD COLUMN IF NOT EXISTS "imageUrl" TEXT;

-- Add groupName to RecipeIngredient
ALTER TABLE "RecipeIngredient"
ADD COLUMN IF NOT EXISTS "groupName" VARCHAR(50);

-- Create index for faster groupName filtering
CREATE INDEX IF NOT EXISTS idx_recipe_ingredient_group ON "RecipeIngredient" ("groupName");

-- Verification query:
-- SELECT 
--   column_name, 
--   data_type, 
--   is_nullable
-- FROM information_schema.columns
-- WHERE table_name = 'Recipe' 
-- AND column_name IN ('description', 'imageUrl')
-- UNION ALL
-- SELECT 
--   column_name, 
--   data_type, 
--   is_nullable
-- FROM information_schema.columns
-- WHERE table_name = 'RecipeIngredient' 
-- AND column_name = 'groupName';
