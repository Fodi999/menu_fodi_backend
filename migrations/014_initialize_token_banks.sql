-- Initialize token_bank records for all existing users
INSERT INTO token_bank (user_id, balance, total_allocated, total_used, created_at, updated_at)
SELECT id, 0, 0, 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
FROM "User"
WHERE id NOT IN (SELECT user_id FROM token_bank)
ON CONFLICT (user_id) DO NOTHING;

-- Add a trigger to automatically create token_bank entry for new users
CREATE OR REPLACE FUNCTION create_token_bank_for_new_user()
RETURNS TRIGGER AS $$
BEGIN
  INSERT INTO token_bank (user_id, balance, total_allocated, total_used)
  VALUES (NEW.id, 0, 0, 0)
  ON CONFLICT (user_id) DO NOTHING;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Drop the trigger if it already exists
DROP TRIGGER IF EXISTS trigger_create_token_bank ON "User";

-- Create the trigger
CREATE TRIGGER trigger_create_token_bank
AFTER INSERT ON "User"
FOR EACH ROW
EXECUTE FUNCTION create_token_bank_for_new_user();
