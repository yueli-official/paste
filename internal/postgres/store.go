package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/yueli-official/paste/internal/governance"
	"github.com/yueli-official/paste/internal/paste"
)

type Store struct {
	database *sql.DB
}

func New(database *sql.DB) (*Store, error) {
	if database == nil {
		return nil, errors.New("paste/postgres: DB is required")
	}
	return &Store{database: database}, nil
}

func (store *Store) Insert(ctx context.Context, value paste.Paste) (paste.Paste, error) {
	transaction, err := store.database.BeginTx(ctx, nil)
	if err != nil {
		return paste.Paste{}, fmt.Errorf("begin Paste insert: %w", err)
	}
	defer transaction.Rollback()
	if err := claimCreation(ctx, transaction, value); err != nil {
		return paste.Paste{}, err
	}

	_, err = transaction.ExecContext(ctx, `
INSERT INTO pastes (
    id, code, owner_user_key, title, description, tags, visibility,
    password_hash, state, revision, created_at, updated_at, expires_at, deleted_at
) VALUES (
    $1::uuid, $2, $3, $4, $5, $6, $7,
    $8, $9, $10, $11, $12, $13, $14
)`, value.ID, value.Code, nullableString(value.OwnerUserKey), value.Title, value.Description,
		pq.Array(append([]string{}, value.Tags...)), value.Visibility, nullableBytes(value.PasswordHash), value.State, value.Revision,
		value.CreatedAt, value.UpdatedAt, value.ExpiresAt, value.DeletedAt)
	if err != nil {
		return paste.Paste{}, mapWriteError(err)
	}
	if err := replaceFiles(ctx, transaction, value.ID, value.Files); err != nil {
		return paste.Paste{}, err
	}
	if err := transaction.Commit(); err != nil {
		return paste.Paste{}, fmt.Errorf("commit Paste insert: %w", err)
	}
	return clone(value), nil
}

func claimCreation(ctx context.Context, transaction *sql.Tx, value paste.Paste) error {
	settings, err := scanGovernanceSettings(transaction.QueryRowContext(ctx, `
SELECT user_daily_limit, anonymous_daily_limit, revision, updated_at, updated_by
FROM paste_governance_settings
WHERE singleton = true
FOR SHARE`))
	if err != nil {
		return fmt.Errorf("read creation governance settings: %w", err)
	}
	actorKind := "anonymous"
	actorKey := "global"
	limit := settings.AnonymousDailyLimit
	if value.OwnerUserKey != "" {
		actorKind = "user"
		actorKey = value.OwnerUserKey
		policy := governance.UserPolicy{UserKey: actorKey, State: governance.UserStateActive}
		policyRow := transaction.QueryRowContext(ctx, `
SELECT user_key, state, daily_limit_override, reason, revision, updated_at, updated_by
FROM paste_user_policies
WHERE user_key = $1
FOR SHARE`, actorKey)
		policy, err = scanGovernancePolicy(policyRow)
		if errors.Is(err, sql.ErrNoRows) {
			policy = governance.UserPolicy{UserKey: actorKey, State: governance.UserStateActive}
		} else if err != nil {
			return fmt.Errorf("read creation user policy: %w", err)
		}
		if policy.State == governance.UserStateSuspended {
			return governance.ErrCreationSuspended
		}
		limit = governance.EffectiveDailyLimit(settings, policy)
	} else if limit == 0 {
		return governance.ErrAnonymousCreationOff
	}
	var used int
	err = transaction.QueryRowContext(ctx, `
INSERT INTO paste_daily_creation_usage (usage_day, actor_kind, actor_key, used, updated_at)
VALUES ($1::date, $2, $3, 1, $4)
ON CONFLICT (usage_day, actor_kind, actor_key) DO UPDATE SET
    used = paste_daily_creation_usage.used + 1,
    updated_at = EXCLUDED.updated_at
WHERE paste_daily_creation_usage.used < $5
RETURNING used`, value.CreatedAt.UTC().Format("2006-01-02"), actorKind, actorKey, value.CreatedAt, limit).Scan(&used)
	if errors.Is(err, sql.ErrNoRows) {
		return governance.ErrDailyLimitReached
	}
	if err != nil {
		return fmt.Errorf("claim daily creation usage: %w", err)
	}
	return nil
}

func (store *Store) GetByCode(ctx context.Context, code string) (paste.Paste, error) {
	value, err := scanPaste(store.database.QueryRowContext(ctx, `
SELECT `+pasteColumns+`
FROM pastes
WHERE code = $1`, code))
	if errors.Is(err, sql.ErrNoRows) {
		return paste.Paste{}, paste.ErrNotFound
	}
	if err != nil {
		return paste.Paste{}, fmt.Errorf("get Paste: %w", err)
	}
	value.Files, err = store.files(ctx, value.ID)
	if err != nil {
		return paste.Paste{}, err
	}
	return value, nil
}

func (store *Store) ListByOwner(ctx context.Context, userKey string) ([]paste.Paste, error) {
	rows, err := store.database.QueryContext(ctx, `
SELECT `+pasteColumns+`
FROM pastes
WHERE owner_user_key = $1
ORDER BY created_at DESC`, userKey)
	if err != nil {
		return nil, fmt.Errorf("list owner Pastes: %w", err)
	}
	defer rows.Close()

	values := make([]paste.Paste, 0)
	for rows.Next() {
		value, scanErr := scanPaste(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("scan owner Paste: %w", scanErr)
		}
		values = append(values, value)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate owner Pastes: %w", err)
	}
	for index := range values {
		values[index].Files, err = store.files(ctx, values[index].ID)
		if err != nil {
			return nil, err
		}
	}
	return values, nil
}

func (store *Store) Update(ctx context.Context, value paste.Paste, expectedRevision int64) (paste.Paste, error) {
	transaction, err := store.database.BeginTx(ctx, nil)
	if err != nil {
		return paste.Paste{}, fmt.Errorf("begin Paste update: %w", err)
	}
	defer transaction.Rollback()

	result, err := transaction.ExecContext(ctx, `
UPDATE pastes SET
    owner_user_key = $1,
    title = $2,
    description = $3,
    tags = $4,
    visibility = $5,
    password_hash = $6,
    state = $7,
    revision = $8,
    updated_at = $9,
    expires_at = $10,
    deleted_at = $11
WHERE code = $12 AND revision = $13`,
		nullableString(value.OwnerUserKey), value.Title, value.Description, pq.Array(append([]string{}, value.Tags...)), value.Visibility,
		nullableBytes(value.PasswordHash), value.State, value.Revision, value.UpdatedAt, value.ExpiresAt, value.DeletedAt,
		value.Code, expectedRevision)
	if err != nil {
		return paste.Paste{}, mapWriteError(err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return paste.Paste{}, fmt.Errorf("read Paste update result: %w", err)
	}
	if affected == 0 {
		return paste.Paste{}, paste.ErrConflict
	}
	if err := replaceFiles(ctx, transaction, value.ID, value.Files); err != nil {
		return paste.Paste{}, err
	}
	if err := transaction.Commit(); err != nil {
		return paste.Paste{}, fmt.Errorf("commit Paste update: %w", err)
	}
	return clone(value), nil
}

const pasteColumns = `
id::text, code, owner_user_key, title, description, tags, visibility,
password_hash, state, revision, created_at, updated_at, expires_at, deleted_at`

type scanner interface {
	Scan(...any) error
}

func scanPaste(row scanner) (paste.Paste, error) {
	var (
		value     paste.Paste
		owner     sql.NullString
		tags      pq.StringArray
		expiresAt sql.NullTime
		deletedAt sql.NullTime
	)
	err := row.Scan(
		&value.ID, &value.Code, &owner, &value.Title, &value.Description, &tags, &value.Visibility,
		&value.PasswordHash, &value.State, &value.Revision, &value.CreatedAt, &value.UpdatedAt, &expiresAt, &deletedAt,
	)
	if err != nil {
		return paste.Paste{}, err
	}
	if owner.Valid {
		value.OwnerUserKey = owner.String
	}
	value.Tags = append([]string{}, tags...)
	if expiresAt.Valid {
		value.ExpiresAt = timePointer(expiresAt.Time)
	}
	if deletedAt.Valid {
		value.DeletedAt = timePointer(deletedAt.Time)
	}
	value.PasswordGuard = len(value.PasswordHash) > 0
	return value, nil
}

func (store *Store) files(ctx context.Context, pasteID string) ([]paste.File, error) {
	rows, err := store.database.QueryContext(ctx, `
SELECT path, language, content, ordinal
FROM paste_files
WHERE paste_id = $1::uuid
ORDER BY ordinal`, pasteID)
	if err != nil {
		return nil, fmt.Errorf("list Paste files: %w", err)
	}
	defer rows.Close()
	files := make([]paste.File, 0)
	for rows.Next() {
		var file paste.File
		if err := rows.Scan(&file.Path, &file.Language, &file.Content, &file.Order); err != nil {
			return nil, fmt.Errorf("scan Paste file: %w", err)
		}
		files = append(files, file)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate Paste files: %w", err)
	}
	return files, nil
}

func replaceFiles(ctx context.Context, transaction *sql.Tx, pasteID string, files []paste.File) error {
	if _, err := transaction.ExecContext(ctx, `DELETE FROM paste_files WHERE paste_id = $1::uuid`, pasteID); err != nil {
		return fmt.Errorf("clear Paste files: %w", err)
	}
	for _, file := range files {
		if _, err := transaction.ExecContext(ctx, `
INSERT INTO paste_files (paste_id, ordinal, path, language, content)
VALUES ($1::uuid, $2, $3, $4, $5)`, pasteID, file.Order, file.Path, file.Language, file.Content); err != nil {
			return fmt.Errorf("insert Paste file: %w", err)
		}
	}
	return nil
}

func mapWriteError(err error) error {
	var postgresError *pq.Error
	if errors.As(err, &postgresError) && postgresError.Code == "23505" && postgresError.Constraint == "pastes_code_unique" {
		return paste.ErrCodeCollision
	}
	return fmt.Errorf("write Paste: %w", err)
}

func nullableString(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}

func nullableBytes(value []byte) any {
	if len(value) == 0 {
		return nil
	}
	return value
}

func timePointer(value time.Time) *time.Time {
	copy := value
	return &copy
}

func clone(value paste.Paste) paste.Paste {
	value.Tags = append([]string(nil), value.Tags...)
	value.Files = append([]paste.File(nil), value.Files...)
	value.PasswordHash = append([]byte(nil), value.PasswordHash...)
	return value
}
