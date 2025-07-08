ALTER TABLE retailcrm_settings DROP COLUMN api_url;
ALTER TABLE retailcrm_settings ADD COLUMN base_url TEXT NOT NULL default 'https://pristavkin.retailcrm.ru';