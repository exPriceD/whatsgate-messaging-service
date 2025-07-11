ALTER TABLE campaigns ADD COLUMN IF NOT EXISTS category_name TEXT;
CREATE INDEX IF NOT EXISTS idx_campaigns_category_name ON campaigns(category_name);