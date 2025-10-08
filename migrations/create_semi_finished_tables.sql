-- Таблица полуфабрикатов
CREATE TABLE IF NOT EXISTS semi_finished (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    output_quantity DECIMAL(10, 3) NOT NULL,
    output_unit VARCHAR(10) NOT NULL,
    cost_per_unit DECIMAL(10, 2) NOT NULL DEFAULT 0,
    category VARCHAR(100),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Таблица ингредиентов в составе полуфабрикатов
CREATE TABLE IF NOT EXISTS semi_finished_ingredients (
    id VARCHAR(36) PRIMARY KEY,
    semi_finished_id VARCHAR(36) NOT NULL,
    ingredient_id VARCHAR(36) NOT NULL,
    ingredient_name VARCHAR(255) NOT NULL,
    quantity DECIMAL(10, 3) NOT NULL,
    unit VARCHAR(10) NOT NULL,
    price_per_unit DECIMAL(10, 2) NOT NULL,
    total_price DECIMAL(10, 2) NOT NULL,
    FOREIGN KEY (semi_finished_id) REFERENCES semi_finished(id) ON DELETE CASCADE,
    FOREIGN KEY (ingredient_id) REFERENCES ingredients(id) ON DELETE RESTRICT
);

-- Индексы для производительности
CREATE INDEX IF NOT EXISTS idx_semi_finished_name ON semi_finished(name);
CREATE INDEX IF NOT EXISTS idx_semi_finished_category ON semi_finished(category);
CREATE INDEX IF NOT EXISTS idx_semi_finished_ingredients_sf_id ON semi_finished_ingredients(semi_finished_id);
CREATE INDEX IF NOT EXISTS idx_semi_finished_ingredients_ing_id ON semi_finished_ingredients(ingredient_id);
