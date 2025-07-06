-- ============================================================================
-- Migration 006 Rollback: Drop Improved Database Schema
-- ============================================================================

-- Drop triggers first
DROP TRIGGER IF EXISTS update_retailcrm_settings_updated_at ON retailcrm_settings;
DROP TRIGGER IF EXISTS update_whatsgate_settings_updated_at ON whatsgate_settings;
DROP TRIGGER IF EXISTS update_media_files_updated_at ON media_files;
DROP TRIGGER IF EXISTS update_campaigns_updated_at ON campaigns;
DROP TRIGGER IF EXISTS update_campaign_phone_numbers_updated_at ON campaign_phone_numbers;
DROP TRIGGER IF EXISTS update_campaign_stats_updated_at ON campaign_stats;

-- Drop trigger function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_media_files_mime_type;
DROP INDEX IF EXISTS idx_media_files_message_type;
DROP INDEX IF EXISTS idx_media_files_created_at;
DROP INDEX IF EXISTS idx_campaigns_status;
DROP INDEX IF EXISTS idx_campaigns_created_at;
DROP INDEX IF EXISTS idx_campaigns_initiator;
DROP INDEX IF EXISTS idx_campaigns_media_file_id;
DROP INDEX IF EXISTS idx_campaign_phone_numbers_campaign_id;
DROP INDEX IF EXISTS idx_campaign_phone_numbers_status;
DROP INDEX IF EXISTS idx_campaign_phone_numbers_phone;
DROP INDEX IF EXISTS idx_campaign_phone_numbers_sent_at;
DROP INDEX IF EXISTS idx_campaign_stats_campaign_id;
DROP INDEX IF EXISTS idx_campaign_stats_date;
DROP INDEX IF EXISTS idx_campaign_stats_unique;

-- Drop tables (in order due to foreign key constraints)
DROP TABLE IF EXISTS campaign_stats;
DROP TABLE IF EXISTS campaign_phone_numbers;
DROP TABLE IF EXISTS campaigns;
DROP TABLE IF EXISTS media_files;
DROP TABLE IF EXISTS retailcrm_settings; 