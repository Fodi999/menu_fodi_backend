-- Добавляем недостающие столбцы в таблицу semi_finished
ALTER TABLE semi_finished
ADD COLUMN IF NOT EXISTS total_cost DOUBLE PRECISION DEFAULT 0,
ADD COLUMN IF NOT EXISTS is_visible BOOLEAN DEFAULT TRUE,
ADD COLUMN IF NOT EXISTS is_archived BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP NULL;

-- Обновляем существующие записи: пересчитываем total_cost
UPDATE semi_finished
SET total_cost = cost_per_unit * output_quantity
WHERE total_cost = 0 OR total_cost IS NULL;
