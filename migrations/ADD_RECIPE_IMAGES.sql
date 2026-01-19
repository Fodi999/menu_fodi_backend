-- =====================================================
-- Добавление полей для изображений рецептов
-- =====================================================
-- Цель: Хранить Cloudinary URLs и metadata
-- Дата: 2026-01-19
-- =====================================================

BEGIN;

-- Добавляем поля для изображений
ALTER TABLE "Recipe"
ADD COLUMN IF NOT EXISTS "imageUrl" TEXT,
ADD COLUMN IF NOT EXISTS "imagePublicId" TEXT;

-- Комментарии для документации
COMMENT ON COLUMN "Recipe"."imageUrl" IS 'Cloudinary CDN URL (https://res.cloudinary.com/...)';
COMMENT ON COLUMN "Recipe"."imagePublicId" IS 'Cloudinary public ID для трансформаций и удаления';

-- Индекс для поиска рецептов с изображениями
CREATE INDEX IF NOT EXISTS idx_recipe_has_image
ON "Recipe" ("imageUrl")
WHERE "imageUrl" IS NOT NULL;

COMMIT;

-- Проверка
SELECT 
    'Recipes with images' as metric,
    COUNT(*) as count
FROM "Recipe"
WHERE "imageUrl" IS NOT NULL;

SELECT 
    'Total recipes' as metric,
    COUNT(*) as count
FROM "Recipe";
