-- Добавление новых полей в таблицу semi_finished

-- Добавляем total_cost (общая себестоимость партии)
ALTER TABLE semi_finished 
ADD COLUMN IF NOT EXISTS total_cost DECIMAL(10,2) DEFAULT 0;

-- Добавляем is_visible (видимость в меню)
ALTER TABLE semi_finished 
ADD COLUMN IF NOT EXISTS is_visible BOOLEAN DEFAULT true;

-- Добавляем is_archived (архивирование)
ALTER TABLE semi_finished 
ADD COLUMN IF NOT EXISTS is_archived BOOLEAN DEFAULT false;

-- Добавляем deleted_at для soft delete
ALTER TABLE semi_finished 
ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP;

-- Обновляем существующие записи: рассчитываем total_cost
UPDATE semi_finished 
SET total_cost = cost_per_unit * output_quantity
WHERE total_cost = 0 OR total_cost IS NULL;

-- Создаём индексы для быстрого поиска
CREATE INDEX IF NOT EXISTS idx_semi_finished_visible 
ON semi_finished(is_visible) 
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_semi_finished_archived 
ON semi_finished(is_archived) 
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_semi_finished_deleted 
ON semi_finished(deleted_at);

-- Комментарии к новым полям
COMMENT ON COLUMN semi_finished.total_cost IS 'Общая себестоимость партии полуфабриката';
COMMENT ON COLUMN semi_finished.is_visible IS 'Видимость в меню и отчётах';
COMMENT ON COLUMN semi_finished.is_archived IS 'Флаг архивирования (скрыто, но не удалено)';
COMMENT ON COLUMN semi_finished.deleted_at IS 'Дата мягкого удаления (soft delete)';
