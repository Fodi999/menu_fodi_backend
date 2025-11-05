-- Add recipe metrics: nutrition, costs, and ChefTokens rewards
-- Migration 005: Extended recipe analytics

ALTER TABLE "Recipe"
ADD COLUMN gross_weight INTEGER,           -- Брутто (вес сырых продуктов в г)
ADD COLUMN net_weight INTEGER,             -- Нетто (вес готового блюда в г)
ADD COLUMN calories INTEGER,               -- Калорийность (ккал на всё блюдо)
ADD COLUMN protein NUMERIC(10,2),          -- Белки (г)
ADD COLUMN fats NUMERIC(10,2),             -- Жиры (г)
ADD COLUMN carbs NUMERIC(10,2),            -- Углеводы (г)
ADD COLUMN recipe_yield INTEGER,           -- Выход (вес готового блюда в г)
ADD COLUMN cost NUMERIC(10,2),             -- Себестоимость (PLN/USD)
ADD COLUMN tokens_reward INTEGER DEFAULT 10, -- Награда ChefTokens за создание
ADD COLUMN views_count INTEGER DEFAULT 0,    -- Количество просмотров
ADD COLUMN tokens_earned INTEGER DEFAULT 0;  -- Всего заработано токенов

-- Комментарии к полям
COMMENT ON COLUMN "Recipe".gross_weight IS 'Брутто - вес сырых ингредиентов (г)';
COMMENT ON COLUMN "Recipe".net_weight IS 'Нетто - вес обработанных ингредиентов (г)';
COMMENT ON COLUMN "Recipe".calories IS 'Калорийность всего блюда (ккал)';
COMMENT ON COLUMN "Recipe".protein IS 'Белки на всё блюдо (г)';
COMMENT ON COLUMN "Recipe".fats IS 'Жиры на всё блюдо (г)';
COMMENT ON COLUMN "Recipe".carbs IS 'Углеводы на всё блюдо (г)';
COMMENT ON COLUMN "Recipe".recipe_yield IS 'Выход готового блюда (г)';
COMMENT ON COLUMN "Recipe".cost IS 'Себестоимость блюда (PLN)';
COMMENT ON COLUMN "Recipe".tokens_reward IS 'Награда ChefTokens за публикацию рецепта';
COMMENT ON COLUMN "Recipe".views_count IS 'Количество просмотров рецепта';
COMMENT ON COLUMN "Recipe".tokens_earned IS 'Всего заработано ChefTokens с этого рецепта';

-- Обновим существующие рецепты значениями по умолчанию
UPDATE "Recipe"
SET 
  gross_weight = 500,
  net_weight = 400,
  calories = 350,
  protein = 15.0,
  fats = 12.0,
  carbs = 45.0,
  recipe_yield = 400,
  cost = 25.00,
  tokens_reward = 15,
  views_count = 0,
  tokens_earned = 0
WHERE gross_weight IS NULL;

-- Создадим индекс для быстрого поиска по калорийности и стоимости
CREATE INDEX IF NOT EXISTS idx_recipe_calories ON "Recipe"(calories);
CREATE INDEX IF NOT EXISTS idx_recipe_cost ON "Recipe"(cost);
CREATE INDEX IF NOT EXISTS idx_recipe_tokens_earned ON "Recipe"(tokens_earned DESC);
