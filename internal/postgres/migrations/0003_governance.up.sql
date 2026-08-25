CREATE TABLE IF NOT EXISTS paste_governance_settings (
    singleton boolean PRIMARY KEY DEFAULT true,
    user_daily_limit integer NOT NULL,
    anonymous_daily_limit integer NOT NULL,
    revision bigint NOT NULL,
    updated_at timestamptz NOT NULL,
    updated_by varchar(64),
    CONSTRAINT paste_governance_settings_singleton_check CHECK (singleton),
    CONSTRAINT paste_governance_settings_user_limit_check CHECK (user_daily_limit BETWEEN 1 AND 10000),
    CONSTRAINT paste_governance_settings_anonymous_limit_check CHECK (anonymous_daily_limit BETWEEN 0 AND 100000),
    CONSTRAINT paste_governance_settings_revision_check CHECK (revision > 0)
);

INSERT INTO paste_governance_settings (
    singleton, user_daily_limit, anonymous_daily_limit, revision, updated_at, updated_by
) VALUES (
    true, 50, 200, 1, CURRENT_TIMESTAMP, NULL
) ON CONFLICT (singleton) DO NOTHING;

CREATE TABLE IF NOT EXISTS paste_user_policies (
    user_key varchar(64) PRIMARY KEY,
    state varchar(16) NOT NULL,
    daily_limit_override integer,
    reason varchar(240) NOT NULL DEFAULT '',
    revision bigint NOT NULL,
    updated_at timestamptz NOT NULL,
    updated_by varchar(64),
    CONSTRAINT paste_user_policies_user_key_check CHECK (length(trim(user_key)) > 0),
    CONSTRAINT paste_user_policies_state_check CHECK (state IN ('active', 'suspended')),
    CONSTRAINT paste_user_policies_limit_check CHECK (daily_limit_override IS NULL OR daily_limit_override BETWEEN 1 AND 10000),
    CONSTRAINT paste_user_policies_revision_check CHECK (revision > 0)
);

CREATE INDEX IF NOT EXISTS paste_user_policies_state_idx
    ON paste_user_policies (state, updated_at DESC);

CREATE TABLE IF NOT EXISTS paste_daily_creation_usage (
    usage_day date NOT NULL,
    actor_kind varchar(16) NOT NULL,
    actor_key varchar(64) NOT NULL,
    used integer NOT NULL,
    updated_at timestamptz NOT NULL,
    PRIMARY KEY (usage_day, actor_kind, actor_key),
    CONSTRAINT paste_daily_creation_usage_kind_check CHECK (actor_kind IN ('user', 'anonymous')),
    CONSTRAINT paste_daily_creation_usage_actor_check CHECK (length(trim(actor_key)) > 0),
    CONSTRAINT paste_daily_creation_usage_used_check CHECK (used > 0)
);

INSERT INTO paste_daily_creation_usage (usage_day, actor_kind, actor_key, used, updated_at)
SELECT
    (created_at AT TIME ZONE 'UTC')::date,
    'user',
    owner_user_key,
    count(*)::integer,
    max(created_at)
FROM pastes
WHERE owner_user_key IS NOT NULL
  AND (created_at AT TIME ZONE 'UTC')::date = (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')::date
GROUP BY (created_at AT TIME ZONE 'UTC')::date, owner_user_key
ON CONFLICT (usage_day, actor_kind, actor_key) DO NOTHING;

INSERT INTO paste_daily_creation_usage (usage_day, actor_kind, actor_key, used, updated_at)
SELECT
    (created_at AT TIME ZONE 'UTC')::date,
    'anonymous',
    'global',
    count(*)::integer,
    max(created_at)
FROM pastes
WHERE owner_user_key IS NULL
  AND (created_at AT TIME ZONE 'UTC')::date = (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')::date
GROUP BY (created_at AT TIME ZONE 'UTC')::date
ON CONFLICT (usage_day, actor_kind, actor_key) DO NOTHING;
