-- Create ingredient_categories reference table
-- Categories are a data catalog, not hardcoded enum
-- ✅ Stable key, localized labels, icon, sort order

CREATE TABLE IF NOT EXISTS ingredient_categories (
    key TEXT PRIMARY KEY,
    icon TEXT NOT NULL,
    sort_order INTEGER NOT NULL,
    label_pl TEXT NOT NULL,
    label_en TEXT NOT NULL,
    label_ru TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert category data
-- Icons use emoji for universal support
INSERT INTO ingredient_categories (key, icon, sort_order, label_pl, label_en, label_ru) VALUES
('all', '🧊', 0, 'Wszystko', 'All', 'Все'),
('fish', '🐟', 1, 'Ryby', 'Fish', 'Рыба'),
('meat', '🥩', 2, 'Mięso', 'Meat', 'Мясо'),
('egg', '🥚', 3, 'Jajka', 'Eggs', 'Яйца'),
('dairy', '🥛', 4, 'Nabiał', 'Dairy', 'Молочные'),
('vegetable', '🥕', 5, 'Warzywa', 'Vegetables', 'Овощи'),
('fruit', '🍎', 6, 'Owoce', 'Fruits', 'Фрукты'),
('grain', '🌾', 7, 'Zboża', 'Grains', 'Крупы'),
('condiment', '🧂', 8, 'Przyprawy', 'Condiments', 'Приправы'),
('other', '📦', 9, 'Inne', 'Other', 'Другое')
ON CONFLICT (key) DO NOTHING;

-- Create index for faster queries
CREATE INDEX IF NOT EXISTS idx_ingredient_categories_sort_order ON ingredient_categories(sort_order);

-- Verify data
SELECT * FROM ingredient_categories ORDER BY sort_order;
