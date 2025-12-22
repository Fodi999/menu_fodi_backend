-- +goose Up
ALTER TABLE prepared_dishes ADD COLUMN IF NOT EXISTS cost_per_portion DECIMAL(10, 2);
ALTER TABLE prepared_dishes ADD COLUMN IF NOT EXISTS total_cost DECIMAL(10, 2);

COMMENT ON COLUMN prepared_dishes.cost_per_portion IS 'Cost per single portion (calculated from cook cost / portions)';
COMMENT ON COLUMN prepared_dishes.total_cost IS 'Total cost of the dish when prepared (from cook log)';

-- +goose Down
ALTER TABLE prepared_dishes DROP COLUMN IF EXISTS cost_per_portion;
ALTER TABLE prepared_dishes DROP COLUMN IF EXISTS total_cost;
