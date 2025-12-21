-- Migration: Fix missing ingredients in recipes from 038
-- Problem: Migration 038 used incorrect ingredient names that don't exist in catalog
-- Example: 'Wieprzowina' instead of 'Wieprzowina (schab)', 'Jajko' instead of 'Jaja'
-- Result: Many recipes have only 1-2 ingredients instead of full recipe
--
-- This migration adds missing ingredients using correct names from Ingredient table

-- Fix Kotlet schabowy (should have 4 ingredients, has only 1)
DO $$
DECLARE
    recipe_id UUID;
    ing_wieprzowina UUID;
    ing_jaja UUID;
    ing_bulka UUID;
BEGIN
    SELECT id INTO recipe_id FROM "Recipe" WHERE "localName" = 'Kotlet schabowy' LIMIT 1;
    SELECT id INTO ing_wieprzowina FROM "Ingredient" WHERE name = 'Wieprzowina (schab)' LIMIT 1;
    SELECT id INTO ing_jaja FROM "Ingredient" WHERE name = 'Jaja' LIMIT 1;
    SELECT id INTO ing_bulka FROM "Ingredient" WHERE name = 'Bułka' LIMIT 1;
    
    IF recipe_id IS NOT NULL THEN
        -- Add missing ingredients (skip if already exists)
        IF ing_wieprzowina IS NOT NULL THEN
            INSERT INTO "RecipeIngredient" ("recipeId", "ingredientId", "ingredientKey", quantity, unit, "sortOrder", optional)
            VALUES (recipe_id, ing_wieprzowina, 'wieprzowina-schab', 600, 'g', 1, false)
            ON CONFLICT DO NOTHING;
        END IF;
        
        IF ing_jaja IS NOT NULL THEN
            INSERT INTO "RecipeIngredient" ("recipeId", "ingredientId", "ingredientKey", quantity, unit, "sortOrder", optional)
            VALUES (recipe_id, ing_jaja, 'jaja', 2, 'szt', 2, false)
            ON CONFLICT DO NOTHING;
        END IF;
        
        IF ing_bulka IS NOT NULL THEN
            INSERT INTO "RecipeIngredient" ("recipeId", "ingredientId", "ingredientKey", quantity, unit, "sortOrder", optional)
            VALUES (recipe_id, ing_bulka, 'bulka-tarta', 100, 'g', 3, false)
            ON CONFLICT DO NOTHING;
        END IF;
    END IF;
END $$;

-- Verification query:
-- SELECT
--   r."localName",
--   COUNT(ri.id) as ingredient_count,
--   STRING_AGG(i.name, ', ' ORDER BY ri."sortOrder") as ingredients
-- FROM "Recipe" r
-- LEFT JOIN "RecipeIngredient" ri ON r.id = ri."recipeId"
-- LEFT JOIN "Ingredient" i ON i.id = ri."ingredientId"
-- WHERE r."localName" = 'Kotlet schabowy'
-- GROUP BY r.id, r."localName";
