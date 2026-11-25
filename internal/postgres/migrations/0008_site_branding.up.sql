ALTER TABLE paste_site_settings
    ADD COLUMN IF NOT EXISTS logo_url text NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS favicon_url text NOT NULL DEFAULT '';
