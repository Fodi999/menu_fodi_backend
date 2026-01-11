-- Migration: Recipe Catalog Performance Indexes
-- Description: Добавляем индексы для быстрой фильтрации рецептов
-- Date: 2026-01-11

-- Базовые фильтры
CREATE INDEX IF NOT EXISTS idx_recipes_category ON "Recipe"(category);
CREATE INDEX IF NOT EXISTS idx_recipes_difficulty ON "Recipe"(difficulty);
CREATE INDEX IF NOT EXISTS idx_recipes_time ON "Recipe"("timeMinutes");
CREATE INDEX IF NOT EXISTS idx_recipes_created_at ON "Recipe"("createdAt");

-- JSONB фильтры (используем jsonb_path_ops для оптимизации)
CREATE INDEX IF NOT EXISTS idx_recipes_source_jsonb ON "Recipe" USING gin(source jsonb_path_ops);
CREATE INDEX IF NOT EXISTS idx_recipes_nutrition_jsonb ON "Recipe" USING gin("nutritionProfile" jsonb_path_ops);

-- Композитные индексы для частых комбинаций
CREATE INDEX IF NOT EXISTS idx_recipes_category_difficulty 
ON "Recipe"(category, difficulty);

CREATE INDEX IF NOT EXISTS idx_recipes_category_time 
ON "Recipe"(category, "timeMinutes");

-- Комментарии для документации
COMMENT ON INDEX idx_recipes_category IS 'Fast filtering by category (appetizer, main, dessert, etc.)';
COMMENT ON INDEX idx_recipes_difficulty IS 'Fast filtering by difficulty (easy, medium, hard)';
COMMENT ON INDEX idx_recipes_time IS 'Fast filtering by cooking time';
COMMENT ON INDEX idx_recipes_created_at IS 'Fast sorting by creation date (newest first)';
