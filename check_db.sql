SELECT column_name, data_type, is_nullable
FROM information_schema.columns
WHERE table_name = 'OrderItem'
ORDER BY ordinal_position;
