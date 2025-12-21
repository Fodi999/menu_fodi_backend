-- Migration: Normalize recipe steps format to simple string array
-- Purpose: Convert object format [{"step": 1, "instruction": "..."}] to simple ["1. ..."]
-- Date: 2025-12-21
--
-- Problem: Some recipes use object format which causes parsing errors in frontend
-- Solution: Normalize all steps to simple string array format

-- Convert object format to string array format for 6 recipes
DO $$
DECLARE
    recipe_record RECORD;
    step_record RECORD;
    new_steps JSONB := '[]'::jsonb;
    step_text TEXT;
BEGIN
    -- Process each recipe with object-format steps
    FOR recipe_record IN 
        SELECT id, "localName", steps
        FROM "Recipe"
        WHERE steps::jsonb->0 ? 'instruction'
    LOOP
        new_steps := '[]'::jsonb;
        
        -- Convert each step object to formatted string
        FOR step_record IN 
            SELECT 
                (value->>'step')::int as step_num,
                value->>'instruction' as instruction
            FROM jsonb_array_elements(recipe_record.steps::jsonb)
            ORDER BY (value->>'step')::int
        LOOP
            step_text := step_record.step_num || '. ' || step_record.instruction;
            new_steps := new_steps || jsonb_build_array(step_text);
        END LOOP;
        
        -- Update recipe with normalized steps
        UPDATE "Recipe"
        SET steps = new_steps::json
        WHERE id = recipe_record.id;
        
        RAISE NOTICE 'Normalized steps for: %', recipe_record."localName";
    END LOOP;
END $$;

-- Verification query (shows first step of each recipe):
-- SELECT 
--   "localName",
--   steps::jsonb->0 as first_step,
--   jsonb_array_length(steps::jsonb) as steps_count
-- FROM "Recipe"
-- WHERE "localName" IN ('Sałatka grecka', 'Bigos myśliwski', 'Jajecznica', 'Pierogi ruskie', 'Pizza Margherita', 'Spaghetti alla Carbonara')
-- ORDER BY "localName";
