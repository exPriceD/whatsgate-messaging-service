ALTER TABLE retailcrm_settings DROP COLUMN base_url;
ALTER TABLE retailcrm_settings ADD COLUMN api_url TEXT NOT NULL default 'https://pristavkin.retailcrm.ru';