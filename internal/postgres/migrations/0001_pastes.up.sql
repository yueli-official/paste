CREATE TABLE IF NOT EXISTS pastes (
    id uuid PRIMARY KEY,
    code char(8) NOT NULL,
    owner_user_key varchar(64),
    title varchar(120) NOT NULL,
    description varchar(2000) NOT NULL DEFAULT '',
    tags text[] NOT NULL DEFAULT '{}',
    visibility varchar(16) NOT NULL,
    password_hash bytea,
    state varchar(16) NOT NULL,
    revision bigint NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    expires_at timestamptz,
    deleted_at timestamptz,
    CONSTRAINT pastes_code_unique UNIQUE (code),
    CONSTRAINT pastes_visibility_check CHECK (visibility IN ('unlisted', 'private')),
    CONSTRAINT pastes_private_owner_check CHECK (visibility <> 'private' OR owner_user_key IS NOT NULL),
    CONSTRAINT pastes_state_check CHECK (state IN ('active', 'deleted')),
    CONSTRAINT pastes_revision_check CHECK (revision > 0),
    CONSTRAINT pastes_deleted_at_check CHECK (
        (state = 'active' AND deleted_at IS NULL) OR
        (state = 'deleted' AND deleted_at IS NOT NULL)
    )
);

CREATE TABLE IF NOT EXISTS paste_files (
    paste_id uuid NOT NULL REFERENCES pastes(id) ON DELETE CASCADE,
    ordinal smallint NOT NULL,
    path varchar(180) NOT NULL,
    language varchar(64) NOT NULL,
    content text NOT NULL,
    PRIMARY KEY (paste_id, ordinal),
    CONSTRAINT paste_files_ordinal_check CHECK (ordinal >= 0 AND ordinal < 20)
);

CREATE UNIQUE INDEX IF NOT EXISTS paste_files_path_unique
    ON paste_files (paste_id, lower(path));

CREATE INDEX IF NOT EXISTS pastes_owner_created_idx
    ON pastes (owner_user_key, created_at DESC)
    WHERE owner_user_key IS NOT NULL;

CREATE INDEX IF NOT EXISTS pastes_expiry_idx
    ON pastes (expires_at)
    WHERE state = 'active' AND expires_at IS NOT NULL;

