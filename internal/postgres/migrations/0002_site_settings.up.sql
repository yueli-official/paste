CREATE TABLE IF NOT EXISTS paste_site_settings (
    singleton boolean PRIMARY KEY DEFAULT true,
    site_name varchar(40) NOT NULL,
    site_description varchar(160) NOT NULL DEFAULT '',
    revision bigint NOT NULL,
    updated_at timestamptz NOT NULL,
    updated_by varchar(64),
    CONSTRAINT paste_site_settings_singleton_check CHECK (singleton),
    CONSTRAINT paste_site_settings_name_check CHECK (length(trim(site_name)) > 0),
    CONSTRAINT paste_site_settings_revision_check CHECK (revision > 0)
);

INSERT INTO paste_site_settings (
    singleton, site_name, site_description, revision, updated_at, updated_by
) VALUES (
    true, '代码片段', '轻量、专注的多文件代码分享。', 1, CURRENT_TIMESTAMP, NULL
) ON CONFLICT (singleton) DO NOTHING;
