-- ============================================================================
-- Migration 008: Drop Old Tables After Data Migration
-- ============================================================================

-- Drop old indexes first
DROP INDEX IF EXISTS idx_bulk_campaign_statuses_campaign_id;

-- Drop old tables (in order due to foreign key constraints)
DROP TABLE IF EXISTS bulk_campaign_statuses;
DROP TABLE IF EXISTS bulk_campaigns;

-- Note: whatsgate_settings table is kept and will be used as-is 