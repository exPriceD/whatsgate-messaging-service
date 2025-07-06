-- ============================================================================
-- Migration 006: Create Improved Database Schema
-- ============================================================================

-- 1. Settings Tables
-- ============================================================================

-- RetailCRM Settings
CREATE TABLE IF NOT EXISTS retailcrm_settings (
    id SERIAL PRIMARY KEY,
    api_url TEXT NOT NULL,
    api_key TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

-- 2. Media Files Table (normalized)
-- ============================================================================
CREATE TABLE IF NOT EXISTS media_files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    filename TEXT NOT NULL,
    mime_type TEXT NOT NULL,
    message_type TEXT NOT NULL, -- 'image', 'doc', 'voice', etc.
    file_size BIGINT NOT NULL,
    storage_path TEXT, -- путь к файлу на диске или NULL если в БД
    file_data TEXT, -- Base64 данные или NULL если на диске
    checksum_md5 TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

-- 3. Improved Campaigns Table
-- ============================================================================
CREATE TABLE IF NOT EXISTS campaigns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    message TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending', -- pending, started, finished, cancelled, failed
    media_file_id UUID REFERENCES media_files(id) ON DELETE SET NULL,
    messages_per_hour INT NOT NULL DEFAULT 20,
    total_count INT NOT NULL DEFAULT 0,
    processed_count INT NOT NULL DEFAULT 0,
    error_count INT NOT NULL DEFAULT 0,
    success_count INT NOT NULL DEFAULT 0,
    initiator TEXT,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

-- 4. Improved Phone Numbers Table
-- ============================================================================
CREATE TABLE IF NOT EXISTS campaign_phone_numbers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    phone_number TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending', -- pending, sent, delivered, read, failed
    error_message TEXT,
    whatsapp_message_id TEXT, -- ID сообщения из WhatsApp API
    sent_at TIMESTAMP WITH TIME ZONE,
    delivered_at TIMESTAMP WITH TIME ZONE,
    read_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

-- 5. Campaign Statistics Table (for analytics)
-- ============================================================================
CREATE TABLE IF NOT EXISTS campaign_stats (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    stat_date DATE NOT NULL DEFAULT CURRENT_DATE,
    messages_sent INT NOT NULL DEFAULT 0,
    messages_delivered INT NOT NULL DEFAULT 0,
    messages_read INT NOT NULL DEFAULT 0,
    messages_failed INT NOT NULL DEFAULT 0,
    delivery_rate DECIMAL(5,2) DEFAULT 0.00, -- процент доставки
    read_rate DECIMAL(5,2) DEFAULT 0.00, -- процент прочтения
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

-- 6. Indexes for Performance
-- ============================================================================

-- Media files indexes
CREATE INDEX IF NOT EXISTS idx_media_files_mime_type ON media_files(mime_type);
CREATE INDEX IF NOT EXISTS idx_media_files_message_type ON media_files(message_type);
CREATE INDEX IF NOT EXISTS idx_media_files_created_at ON media_files(created_at DESC);

-- Campaigns indexes
CREATE INDEX IF NOT EXISTS idx_campaigns_status ON campaigns(status);
CREATE INDEX IF NOT EXISTS idx_campaigns_created_at ON campaigns(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_campaigns_initiator ON campaigns(initiator);
CREATE INDEX IF NOT EXISTS idx_campaigns_media_file_id ON campaigns(media_file_id);

-- Phone numbers indexes
CREATE INDEX IF NOT EXISTS idx_campaign_phone_numbers_campaign_id ON campaign_phone_numbers(campaign_id);
CREATE INDEX IF NOT EXISTS idx_campaign_phone_numbers_status ON campaign_phone_numbers(status);
CREATE INDEX IF NOT EXISTS idx_campaign_phone_numbers_phone ON campaign_phone_numbers(phone_number);
CREATE INDEX IF NOT EXISTS idx_campaign_phone_numbers_sent_at ON campaign_phone_numbers(sent_at DESC);

-- Statistics indexes
CREATE INDEX IF NOT EXISTS idx_campaign_stats_campaign_id ON campaign_stats(campaign_id);
CREATE INDEX IF NOT EXISTS idx_campaign_stats_date ON campaign_stats(stat_date DESC);
CREATE UNIQUE INDEX IF NOT EXISTS idx_campaign_stats_unique ON campaign_stats(campaign_id, stat_date);

-- 7. Add updated_at trigger function
-- ============================================================================
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Add triggers for updated_at
CREATE TRIGGER update_retailcrm_settings_updated_at BEFORE UPDATE ON retailcrm_settings FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_whatsgate_settings_updated_at BEFORE UPDATE ON whatsgate_settings FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_media_files_updated_at BEFORE UPDATE ON media_files FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_campaigns_updated_at BEFORE UPDATE ON campaigns FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_campaign_phone_numbers_updated_at BEFORE UPDATE ON campaign_phone_numbers FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_campaign_stats_updated_at BEFORE UPDATE ON campaign_stats FOR EACH ROW EXECUTE FUNCTION update_updated_at_column(); 