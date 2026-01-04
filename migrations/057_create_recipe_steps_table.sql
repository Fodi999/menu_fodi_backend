-- Migration 057: Create RecipeStep table for localized cooking instructions
-- Replaces JSONB steps_* columns with a proper relational structure

-- Create RecipeStep table
CREATE TABLE IF NOT EXISTS "RecipeStep" (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    recipe_id UUID NOT NULL REFERENCES "Recipe"(id) ON DELETE CASCADE,
    step_number INTEGER NOT NULL,
    language VARCHAR(5) NOT NULL CHECK (language IN ('pl', 'en', 'ru')),
    instruction TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    -- Ensure unique combination of recipe, step number, and language
    UNIQUE(recipe_id, step_number, language)
);

-- Create indexes for fast queries
CREATE INDEX IF NOT EXISTS idx_recipe_step_recipe_id ON "RecipeStep"(recipe_id);
CREATE INDEX IF NOT EXISTS idx_recipe_step_language ON "RecipeStep"(language);
CREATE INDEX IF NOT EXISTS idx_recipe_step_recipe_lang ON "RecipeStep"(recipe_id, language);

-- Add comments
COMMENT ON TABLE "RecipeStep" IS 'Localized cooking instructions for recipes (supports pl/en/ru, extensible to more languages)';
COMMENT ON COLUMN "RecipeStep".step_number IS 'Sequential order of the step (1, 2, 3, ...)';
COMMENT ON COLUMN "RecipeStep".language IS 'Language code: pl (Polish), en (English), ru (Russian)';
COMMENT ON COLUMN "RecipeStep".instruction IS 'Full instruction text in the specified language';

-- Migrate data from existing steps_pl/steps_en/steps_ru JSONB columns
-- This function extracts JSONB array elements and inserts them as separate rows
DO $$
DECLARE
    recipe RECORD;
    step_idx INTEGER;
    step_text TEXT;
BEGIN
    -- Migrate Polish steps (steps_pl)
    FOR recipe IN 
        SELECT id, steps_pl 
        FROM "Recipe" 
        WHERE steps_pl IS NOT NULL AND jsonb_array_length(steps_pl) > 0
    LOOP
        FOR step_idx IN 0..(jsonb_array_length(recipe.steps_pl) - 1)
        LOOP
            step_text := recipe.steps_pl->>step_idx;
            IF step_text IS NOT NULL AND trim(step_text) != '' THEN
                INSERT INTO "RecipeStep" (recipe_id, step_number, language, instruction)
                VALUES (recipe.id, step_idx + 1, 'pl', step_text)
                ON CONFLICT (recipe_id, step_number, language) DO NOTHING;
            END IF;
        END LOOP;
    END LOOP;

    -- Migrate English steps (steps_en)
    FOR recipe IN 
        SELECT id, steps_en 
        FROM "Recipe" 
        WHERE steps_en IS NOT NULL AND jsonb_array_length(steps_en) > 0
    LOOP
        FOR step_idx IN 0..(jsonb_array_length(recipe.steps_en) - 1)
        LOOP
            step_text := recipe.steps_en->>step_idx;
            IF step_text IS NOT NULL AND trim(step_text) != '' THEN
                INSERT INTO "RecipeStep" (recipe_id, step_number, language, instruction)
                VALUES (recipe.id, step_idx + 1, 'en', step_text)
                ON CONFLICT (recipe_id, step_number, language) DO NOTHING;
            END IF;
        END LOOP;
    END LOOP;

    -- Migrate Russian steps (steps_ru)
    FOR recipe IN 
        SELECT id, steps_ru 
        FROM "Recipe" 
        WHERE steps_ru IS NOT NULL AND jsonb_array_length(steps_ru) > 0
    LOOP
        FOR step_idx IN 0..(jsonb_array_length(recipe.steps_ru) - 1)
        LOOP
            step_text := recipe.steps_ru->>step_idx;
            IF step_text IS NOT NULL AND trim(step_text) != '' THEN
                INSERT INTO "RecipeStep" (recipe_id, step_number, language, instruction)
                VALUES (recipe.id, step_idx + 1, 'ru', step_text)
                ON CONFLICT (recipe_id, step_number, language) DO NOTHING;
            END IF;
        END LOOP;
    END LOOP;
END $$;

-- Verify migration
SELECT 
    COUNT(*) FILTER (WHERE language = 'pl') as polish_steps,
    COUNT(*) FILTER (WHERE language = 'en') as english_steps,
    COUNT(*) FILTER (WHERE language = 'ru') as russian_steps,
    COUNT(DISTINCT recipe_id) as recipes_with_steps
FROM "RecipeStep";

-- Note: We keep the old steps_* columns for backward compatibility
-- They can be dropped later after full migration and testing:
-- ALTER TABLE "Recipe" DROP COLUMN IF EXISTS steps_pl;
-- ALTER TABLE "Recipe" DROP COLUMN IF EXISTS steps_en;
-- ALTER TABLE "Recipe" DROP COLUMN IF EXISTS steps_ru;
