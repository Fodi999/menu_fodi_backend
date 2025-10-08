-- Добавляем колонку pricePerUnit в таблицу StockItem
-- Это позволит сохранять цену за единицу (кг/л/шт) для корректного отображения

ALTER TABLE "StockItem"
ADD COLUMN IF NOT EXISTS "pricePerUnit" DOUBLE PRECISION;

COMMENT ON COLUMN "StockItem"."pricePerUnit" IS 'Цена за единицу измерения (кг/л/шт) - для корректного отображения и расчетов';
