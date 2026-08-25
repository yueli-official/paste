package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/yueli-official/paste/internal/governance"
)

func (store *Store) ReadGovernanceSettings(ctx context.Context) (governance.Settings, error) {
	return scanGovernanceSettings(store.database.QueryRowContext(ctx, `
SELECT user_daily_limit, anonymous_daily_limit, revision, updated_at, updated_by
FROM paste_governance_settings
WHERE singleton = true`))
}

func (store *Store) UpdateGovernanceSettings(ctx context.Context, value governance.Settings, expectedRevision int64) (governance.Settings, error) {
	updated, err := scanGovernanceSettings(store.database.QueryRowContext(ctx, `
UPDATE paste_governance_settings SET
    user_daily_limit = $1,
    anonymous_daily_limit = $2,
    revision = $3,
    updated_at = $4,
    updated_by = $5
WHERE singleton = true AND revision = $6
RETURNING user_daily_limit, anonymous_daily_limit, revision, updated_at, updated_by`,
		value.UserDailyLimit, value.AnonymousDailyLimit, value.Revision, value.UpdatedAt,
		nullableString(value.UpdatedBy), expectedRevision))
	if errors.Is(err, sql.ErrNoRows) {
		return governance.Settings{}, governance.ErrConflict
	}
	if err != nil {
		return governance.Settings{}, fmt.Errorf("update Paste governance settings: %w", err)
	}
	return updated, nil
}

func scanGovernanceSettings(row scanner) (governance.Settings, error) {
	var value governance.Settings
	var updatedBy sql.NullString
	if err := row.Scan(
		&value.UserDailyLimit, &value.AnonymousDailyLimit, &value.Revision, &value.UpdatedAt, &updatedBy,
	); err != nil {
		return governance.Settings{}, err
	}
	if updatedBy.Valid {
		value.UpdatedBy = updatedBy.String
	}
	return value, nil
}

const governanceUserSummaries = `
WITH known_users AS (
    SELECT owner_user_key AS user_key FROM pastes WHERE owner_user_key IS NOT NULL
    UNION
    SELECT user_key FROM paste_user_policies
), summaries AS (
    SELECT
        known.user_key,
        COALESCE(policy.state, 'active') AS state,
        policy.daily_limit_override,
        COALESCE(policy.daily_limit_override, settings.user_daily_limit) AS effective_daily_limit,
        COALESCE(usage.used, 0) AS used_today,
        count(value.id)::integer AS total_pastes,
        count(value.id) FILTER (WHERE value.state = 'active')::integer AS active_pastes,
        max(value.created_at) AS last_created_at,
        COALESCE(policy.reason, '') AS reason,
        COALESCE(policy.revision, 0) AS revision,
        policy.updated_at,
        policy.updated_by
    FROM known_users AS known
    CROSS JOIN paste_governance_settings AS settings
    LEFT JOIN paste_user_policies AS policy ON policy.user_key = known.user_key
    LEFT JOIN paste_daily_creation_usage AS usage
      ON usage.usage_day = $3::date
     AND usage.actor_kind = 'user'
     AND usage.actor_key = known.user_key
    LEFT JOIN pastes AS value ON value.owner_user_key = known.user_key
    WHERE settings.singleton = true
    GROUP BY
        known.user_key, policy.state, policy.daily_limit_override, settings.user_daily_limit,
        usage.used, policy.reason, policy.revision, policy.updated_at, policy.updated_by
)
`

const governanceUserWhere = `
WHERE ($1::text = '' OR summaries.user_key ILIKE '%' || $1::text || '%')
  AND ($2::text = '' OR summaries.state = $2::text)`

func (store *Store) ListGovernanceUsers(ctx context.Context, query governance.UserQuery, now time.Time) (governance.UserPage, error) {
	day := now.UTC().Format("2006-01-02")
	arguments := []any{strings.TrimSpace(query.Query), string(query.State), day}
	var total int
	if err := store.database.QueryRowContext(ctx, governanceUserSummaries+`
SELECT count(*) FROM summaries `+governanceUserWhere, arguments...).Scan(&total); err != nil {
		return governance.UserPage{}, fmt.Errorf("count Paste governance users: %w", err)
	}
	rows, err := store.database.QueryContext(ctx, governanceUserSummaries+`
SELECT
    user_key, state, daily_limit_override, effective_daily_limit, used_today,
    total_pastes, active_pastes, last_created_at, reason, revision, updated_at, updated_by
FROM summaries
`+governanceUserWhere+`
ORDER BY used_today DESC, last_created_at DESC NULLS LAST, user_key ASC
LIMIT $4 OFFSET $5`, append(arguments, query.Limit, query.Offset)...)
	if err != nil {
		return governance.UserPage{}, fmt.Errorf("list Paste governance users: %w", err)
	}
	defer rows.Close()
	users := make([]governance.User, 0)
	for rows.Next() {
		var user governance.User
		var override sql.NullInt64
		var lastCreated sql.NullTime
		var updatedAt sql.NullTime
		var updatedBy sql.NullString
		if err := rows.Scan(
			&user.UserKey, &user.State, &override, &user.EffectiveDailyLimit, &user.UsedToday,
			&user.TotalPastes, &user.ActivePastes, &lastCreated, &user.Reason, &user.Revision, &updatedAt, &updatedBy,
		); err != nil {
			return governance.UserPage{}, fmt.Errorf("scan Paste governance user: %w", err)
		}
		if override.Valid {
			value := int(override.Int64)
			user.DailyLimitOverride = &value
		}
		if lastCreated.Valid {
			user.LastCreatedAt = timePointer(lastCreated.Time)
		}
		if updatedAt.Valid {
			user.UpdatedAt = timePointer(updatedAt.Time)
		}
		if updatedBy.Valid {
			user.UpdatedBy = updatedBy.String
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return governance.UserPage{}, fmt.Errorf("iterate Paste governance users: %w", err)
	}
	return governance.UserPage{Users: users, Total: total, Limit: query.Limit, Offset: query.Offset}, nil
}

func (store *Store) UpdateGovernanceUser(ctx context.Context, value governance.UserPolicy, expectedRevision int64) (governance.UserPolicy, error) {
	var row *sql.Row
	if expectedRevision == 0 {
		row = store.database.QueryRowContext(ctx, `
INSERT INTO paste_user_policies (
    user_key, state, daily_limit_override, reason, revision, updated_at, updated_by
) VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (user_key) DO NOTHING
RETURNING user_key, state, daily_limit_override, reason, revision, updated_at, updated_by`,
			value.UserKey, value.State, nullableInt(value.DailyLimitOverride), value.Reason,
			value.Revision, value.UpdatedAt, nullableString(value.UpdatedBy))
	} else {
		row = store.database.QueryRowContext(ctx, `
UPDATE paste_user_policies SET
    state = $1,
    daily_limit_override = $2,
    reason = $3,
    revision = $4,
    updated_at = $5,
    updated_by = $6
WHERE user_key = $7 AND revision = $8
RETURNING user_key, state, daily_limit_override, reason, revision, updated_at, updated_by`,
			value.State, nullableInt(value.DailyLimitOverride), value.Reason, value.Revision,
			value.UpdatedAt, nullableString(value.UpdatedBy), value.UserKey, expectedRevision)
	}
	updated, err := scanGovernancePolicy(row)
	if errors.Is(err, sql.ErrNoRows) {
		return governance.UserPolicy{}, governance.ErrConflict
	}
	if err != nil {
		return governance.UserPolicy{}, fmt.Errorf("update Paste governance user: %w", err)
	}
	return updated, nil
}

func scanGovernancePolicy(row scanner) (governance.UserPolicy, error) {
	var value governance.UserPolicy
	var override sql.NullInt64
	var updatedBy sql.NullString
	if err := row.Scan(
		&value.UserKey, &value.State, &override, &value.Reason, &value.Revision, &value.UpdatedAt, &updatedBy,
	); err != nil {
		return governance.UserPolicy{}, err
	}
	if override.Valid {
		limit := int(override.Int64)
		value.DailyLimitOverride = &limit
	}
	if updatedBy.Valid {
		value.UpdatedBy = updatedBy.String
	}
	return value, nil
}

func nullableInt(value *int) any {
	if value == nil {
		return nil
	}
	return *value
}
