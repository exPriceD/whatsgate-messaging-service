-- ============================================================================
-- Migration 007: Migrate Data from Old Schema to New Schema
-- ============================================================================

-- 1. Migrate media files from bulk_campaigns to media_files
-- ============================================================================
INSERT INTO media_files (id, filename, mime_type, message_type, file_size, file_data, checksum_md5, created_at, updated_at)
SELECT 
    gen_random_uuid() as id,
    COALESCE(media_filename, 'unknown.bin') as filename,
    COALESCE(media_mime, 'application/octet-stream') as mime_type,
    COALESCE(media_type, 'doc') as message_type,
    COALESCE(LENGTH(media_data), 0) as file_size,
    media_data as file_data,
    md5(COALESCE(media_data, '')) as checksum_md5,
    created_at,
    created_at as updated_at
FROM bulk_campaigns 
WHERE media_filename IS NOT NULL 
  AND media_mime IS NOT NULL 
  AND media_type IS NOT NULL
  AND media_data IS NOT NULL;

-- 2. Create temporary mapping table for media files
-- ============================================================================
CREATE TEMP TABLE media_mapping AS
SELECT 
    bc.id as campaign_id,
    mf.id as media_file_id
FROM bulk_campaigns bc
JOIN media_files mf ON (
    mf.filename = COALESCE(bc.media_filename, 'unknown.bin')
    AND mf.mime_type = COALESCE(bc.media_mime, 'application/octet-stream')
    AND mf.message_type = COALESCE(bc.media_type, 'doc')
    AND mf.file_data = bc.media_data
)
WHERE bc.media_filename IS NOT NULL;

-- 3. Migrate campaigns from bulk_campaigns to campaigns
-- ============================================================================
INSERT INTO campaigns (
    id, name, message, status, media_file_id, messages_per_hour, 
    total_count, processed_count, error_count, success_count,
    initiator, created_at, updated_at
)
SELECT 
    bc.id,
    COALESCE(bc.name, 'Imported Campaign') as name,
    bc.message,
    bc.status,
    mm.media_file_id,
    bc.messages_per_hour,
    bc.total,
    COALESCE(bc.processed_count, 0) as processed_count,
    bc.error_count,
    GREATEST(0, COALESCE(bc.processed_count, 0) - bc.error_count) as success_count, -- вычисляем успешные
    bc.initiator,
    bc.created_at,
    bc.created_at as updated_at
FROM bulk_campaigns bc
LEFT JOIN media_mapping mm ON bc.id = mm.campaign_id;

-- 4. Migrate phone numbers from bulk_campaign_statuses to campaign_phone_numbers
-- ============================================================================
INSERT INTO campaign_phone_numbers (
    id, campaign_id, phone_number, status, error_message,
    sent_at, created_at, updated_at
)
SELECT 
    bcs.id,
    bcs.campaign_id,
    bcs.phone_number,
    -- Mapping old status to new status
    CASE 
        WHEN bcs.status = 'pending' THEN 'pending'
        WHEN bcs.status = 'sent' THEN 'sent'
        WHEN bcs.status = 'failed' THEN 'failed'
        ELSE 'pending'
    END as status,
    bcs.error as error_message,
    bcs.sent_at,
    COALESCE(bcs.sent_at, now()) as created_at, -- если sent_at есть, используем его как created_at
    COALESCE(bcs.sent_at, now()) as updated_at
FROM bulk_campaign_statuses bcs;

-- 5. Initialize campaign statistics
-- ============================================================================
INSERT INTO campaign_stats (
    campaign_id, stat_date, messages_sent, messages_delivered, 
    messages_read, messages_failed, delivery_rate, read_rate
)
SELECT 
    c.id as campaign_id,
    CURRENT_DATE as stat_date,
    c.processed_count as messages_sent,
    c.success_count as messages_delivered, -- пока что считаем sent = delivered
    0 as messages_read, -- пока что нет данных о прочтении
    c.error_count as messages_failed,
    CASE 
        WHEN c.total_count > 0 THEN ROUND((c.success_count::DECIMAL / c.total_count::DECIMAL) * 100, 2)
        ELSE 0.00
    END as delivery_rate,
    0.00 as read_rate -- пока что нет данных о прочтении
FROM campaigns c
WHERE c.total_count > 0;

-- 6. Clean up temporary mapping table
-- ============================================================================
DROP TABLE media_mapping; 