-- Migration: Recipe Catalog Performance Indexes
-- Description: Добавляем индексы для быстрой фильтрации рецептов
-- Date: 2026-01-11

-- Базовые фильтры
CREATE INDEX IF NOT EXISTS idx_recipes_category ON "Recipe"(category);
CREATE INDEX IF NOT EXISTS idx_recipes_difficulty ON "Recipe"(difficulty);
CREATE INDEX IF NOT EXISTS idx_recipes_time ON "Recipe"("timeMinutes");
CREATE INDEX IF NOT EXISTS idx_recipes_created_at ON "Recipe"("createdAt");

-- JSONB фильтры (для source и nutrition_profile)
CREATE INDEX IF NOT EXISTS idx_recipes_source_type ON "Recipe" USING gin((source->>'type'));
CREATE INDEX IF NOT EXISTS idx_recipes_source_author ON "Recipe" USING gin((source->>'authorId'));
CREATE INDEX IF NOT EXISTS idx_recipes_calories ON "Recipe" USING btree(((("nutritionProfile"->>'calories'))::int));

-- Индекс для связи рецептов с ингредиентами (JOIN optimization)
CREATE INDEX IF NOT EXISTS idx_recipe_ingredients_ingredient_id 
ON recipe_ingredients(ingredient_id);

CREATE INDEX IF NOT EXISTS idx_recipe_ingredients_recipe_id 
ON recipe_ingredients(recipe_id);

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
COMMENT ON INDEX idx_recipes_source_type IS 'Fast filtering by source type (ai, manual, traditional)';
COMMENT ON INDEX idx_recipe_ingredients_ingredient_id IS 'Fast JOIN with ingredients table';
