-- ============================================================================
-- Миграция: Канонические продукты с системой алиасов
-- Дата: 2026-01-15
-- Цель: Убрать дубликаты, ввести canonicalKey и aliases
-- ============================================================================

-- 1. Создаем таблицу канонических продуктов
CREATE TABLE IF NOT EXISTS "CanonicalIngredient" (
    "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "canonicalKey" VARCHAR(255) NOT NULL UNIQUE, -- нормализованный ключ (onion, garlic)
    "canonicalName" VARCHAR(255) NOT NULL,        -- основное название (Лук репчатый)
    "category" VARCHAR(50) NOT NULL,              -- fish, meat, vegetable, etc.
    "nutritionGroup" VARCHAR(50) NOT NULL,        -- protein, carbohydrate, etc.
    "baseUnit" VARCHAR(10) NOT NULL DEFAULT 'g',  -- g, ml, pcs
    "defaultShelfLifeDays" INTEGER,
    "defaultPricePerUnit" DECIMAL(10,2),
    "imageUrl" TEXT,
    "status" VARCHAR(20) NOT NULL DEFAULT 'active', -- active, archived
    "createdAt" TIMESTAMP NOT NULL DEFAULT NOW(),
    "updatedAt" TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 2. Создаем таблицу алиасов (языковые варианты, опечатки, синонимы)
CREATE TABLE IF NOT EXISTS "IngredientAlias" (
    "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "canonicalIngredientId" UUID NOT NULL REFERENCES "CanonicalIngredient"("id") ON DELETE CASCADE,
    "name" VARCHAR(255) NOT NULL,                -- "лук", "Onion", "репчатый лук"
    "normalizedName" VARCHAR(255) NOT NULL,      -- "лук", "onion", "репчатый лук" (нормализованное)
    "language" VARCHAR(10),                      -- 'pl', 'en', 'ru', 'uk', NULL (любой)
    "aliasType" VARCHAR(50) DEFAULT 'synonym',   -- 'primary', 'translation', 'synonym', 'typo'
    "createdAt" TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 3. Уникальные индексы (защита от дублей)
CREATE UNIQUE INDEX IF NOT EXISTS "idx_canonical_ingredient_key" 
    ON "CanonicalIngredient"("canonicalKey");

CREATE UNIQUE INDEX IF NOT EXISTS "idx_ingredient_alias_normalized" 
    ON "IngredientAlias"("normalizedName");

-- 4. Индексы для производительности
CREATE INDEX IF NOT EXISTS "idx_ingredient_alias_canonical_id" 
    ON "IngredientAlias"("canonicalIngredientId");

CREATE INDEX IF NOT EXISTS "idx_ingredient_alias_language" 
    ON "IngredientAlias"("language");

CREATE INDEX IF NOT EXISTS "idx_canonical_ingredient_status" 
    ON "CanonicalIngredient"("status");

CREATE INDEX IF NOT EXISTS "idx_canonical_ingredient_category" 
    ON "CanonicalIngredient"("category");

-- 5. Триггер для автоматического обновления updatedAt
CREATE OR REPLACE FUNCTION update_canonical_ingredient_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW."updatedAt" = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_canonical_ingredient_updated_at
    BEFORE UPDATE ON "CanonicalIngredient"
    FOR EACH ROW
    EXECUTE FUNCTION update_canonical_ingredient_updated_at();

-- 6. Комментарии для документации
COMMENT ON TABLE "CanonicalIngredient" IS 'Канонические продукты - одна запись = один реальный продукт';
COMMENT ON TABLE "IngredientAlias" IS 'Алиасы продуктов - все языковые варианты, синонимы, опечатки';
COMMENT ON COLUMN "CanonicalIngredient"."canonicalKey" IS 'Нормализованный уникальный ключ (onion, garlic, chicken-breast)';
COMMENT ON COLUMN "IngredientAlias"."normalizedName" IS 'Нормализованное имя для поиска (без регистра, спецсимволов)';
COMMENT ON COLUMN "IngredientAlias"."aliasType" IS 'primary=основное, translation=перевод, synonym=синоним, typo=опечатка';

-- ============================================================================
-- Пока НЕ удаляем старую таблицу Ingredient!
-- Сначала мигрируем данные, потом архивируем
-- ============================================================================
