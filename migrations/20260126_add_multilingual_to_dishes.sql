-- Migration: Add multilingual fields to dishes table
-- Date: 2026-01-26
-- Purpose: Support Polish, English, Russian titles and descriptions for dish cards
--
-- Similar to Recipe table structure, dishes now support:
-- - title_pl, title_en, title_ru (multilingual titles)
-- - description_pl, description_en, description_ru (multilingual descriptions)
--
-- Primary title/description remain for backward compatibility

BEGIN;

-- Add multilingual title columns
ALTER TABLE dishes ADD COLUMN title_pl VARCHAR(255) DEFAULT NULL;
ALTER TABLE dishes ADD COLUMN title_en VARCHAR(255) DEFAULT NULL;
ALTER TABLE dishes ADD COLUMN title_ru VARCHAR(255) DEFAULT NULL;

-- Add multilingual description columns
ALTER TABLE dishes ADD COLUMN description_pl TEXT DEFAULT NULL;
ALTER TABLE dishes ADD COLUMN description_en TEXT DEFAULT NULL;
ALTER TABLE dishes ADD COLUMN description_ru TEXT DEFAULT NULL;

-- Create indexes for language-specific searches
CREATE INDEX idx_dishes_title_pl ON dishes(title_pl) WHERE title_pl IS NOT NULL;
CREATE INDEX idx_dishes_title_en ON dishes(title_en) WHERE title_en IS NOT NULL;
CREATE INDEX idx_dishes_title_ru ON dishes(title_ru) WHERE title_ru IS NOT NULL;

-- Add comment explaining the multilingual structure
COMMENT ON COLUMN dishes.title IS 'Primary title (usually Polish, but can be any language for backward compatibility)';
COMMENT ON COLUMN dishes.title_pl IS 'Polish title for menu display';
COMMENT ON COLUMN dishes.title_en IS 'English title for menu display';
COMMENT ON COLUMN dishes.title_ru IS 'Russian title for menu display';
COMMENT ON COLUMN dishes.description IS 'Primary description (backward compatible)';
COMMENT ON COLUMN dishes.description_pl IS 'Polish product description (marketing copy)';
COMMENT ON COLUMN dishes.description_en IS 'English product description (marketing copy)';
COMMENT ON COLUMN dishes.description_ru IS 'Russian product description (marketing copy)';

COMMIT;
