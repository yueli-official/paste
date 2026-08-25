package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/yueli-official/paste/internal/site"
)

func (store *Store) ReadSettings(ctx context.Context) (site.Settings, error) {
	return scanSettings(store.database.QueryRowContext(ctx, `
SELECT site_name, site_description, revision, updated_at, updated_by
FROM paste_site_settings
WHERE singleton = true`))
}

func (store *Store) UpdateSettings(ctx context.Context, value site.Settings, expectedRevision int64) (site.Settings, error) {
	updated, err := scanSettings(store.database.QueryRowContext(ctx, `
UPDATE paste_site_settings SET
    site_name = $1,
    site_description = $2,
    revision = $3,
    updated_at = $4,
    updated_by = $5
WHERE singleton = true AND revision = $6
RETURNING site_name, site_description, revision, updated_at, updated_by`,
		value.Name, value.Description, value.Revision, value.UpdatedAt, nullableString(value.UpdatedBy), expectedRevision))
	if errors.Is(err, sql.ErrNoRows) {
		return site.Settings{}, site.ErrConflict
	}
	if err != nil {
		return site.Settings{}, fmt.Errorf("update Paste site settings: %w", err)
	}
	return updated, nil
}

func scanSettings(row scanner) (site.Settings, error) {
	var value site.Settings
	var updatedBy sql.NullString
	if err := row.Scan(&value.Name, &value.Description, &value.Revision, &value.UpdatedAt, &updatedBy); err != nil {
		return site.Settings{}, err
	}
	if updatedBy.Valid {
		value.UpdatedBy = updatedBy.String
	}
	return value, nil
}
