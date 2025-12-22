-- +goose Up
CREATE TABLE IF NOT EXISTS prepared_dishes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT NOT NULL REFERENCES "User"(id) ON DELETE CASCADE,
    recipe_id UUID NOT NULL REFERENCES "Recipe"(id) ON DELETE CASCADE,
    portions_available INT NOT NULL DEFAULT 0 CHECK (portions_available >= 0),
    portions_initial INT NOT NULL CHECK (portions_initial > 0),
    prepared_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ,
    source TEXT NOT NULL DEFAULT 'cook' CHECK (source IN ('cook', 'manual')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for fetching user's prepared dishes
CREATE INDEX IF NOT EXISTS idx_prepared_dishes_user_id ON prepared_dishes(user_id);

-- Index for fetching user's prepared dishes sorted by date
CREATE INDEX IF NOT EXISTS idx_prepared_dishes_user_prepared ON prepared_dishes(user_id, prepared_at DESC);

-- Index for filtering available dishes (portions > 0)
CREATE INDEX IF NOT EXISTS idx_prepared_dishes_available ON prepared_dishes(user_id, portions_available) WHERE portions_available > 0;

COMMENT ON TABLE prepared_dishes IS 'Готовые блюда пользователя после приготовления рецептов';
COMMENT ON COLUMN prepared_dishes.portions_available IS 'Сколько порций осталось сейчас';
COMMENT ON COLUMN prepared_dishes.portions_initial IS 'Сколько порций было изначально приготовлено';
COMMENT ON COLUMN prepared_dishes.source IS 'Источник: cook (из рецепта) или manual (добавлено вручную)';

-- +goose Down
DROP TABLE IF EXISTS prepared_dishes;
