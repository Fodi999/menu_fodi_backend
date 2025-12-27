-- Add settings JSONB field to User table
-- Settings include: language, timeFormat, units, aiStyle

ALTER TABLE "User"
ADD COLUMN IF NOT EXISTS settings JSONB DEFAULT '{
  "language": "pl",
  "timeFormat": "24h",
  "units": "metric",
  "aiStyle": "mentor"
}'::jsonb NOT NULL;

-- Add index for faster queries on settings->language
CREATE INDEX IF NOT EXISTS idx_user_settings_language ON "User" USING gin ((settings->'language'));

-- Add comment for documentation
COMMENT ON COLUMN "User".settings IS 'User preferences: language (pl|en|ru), timeFormat (12h|24h), units (metric|kitchen), aiStyle (mentor|direct)';
