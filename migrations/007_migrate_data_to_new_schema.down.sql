-- ============================================================================
-- Migration 007 Rollback: Clear Data from New Schema
-- ============================================================================

-- Clear data from new tables (in reverse order due to foreign keys)
TRUNCATE TABLE campaign_stats;
TRUNCATE TABLE campaign_phone_numbers;
TRUNCATE TABLE campaigns CASCADE;
TRUNCATE TABLE media_files CASCADE;
TRUNCATE TABLE retailcrm_settings;

-- Note: Data will be preserved in old tables if migration 008 hasn't run yet 