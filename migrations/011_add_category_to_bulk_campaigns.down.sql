DROP INDEX IF EXISTS idx_campaigns_category_name;
ALTER TABLE campaigns DROP COLUMN IF EXISTS category_name;