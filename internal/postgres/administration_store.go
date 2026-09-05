package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/lib/pq"
	"github.com/yueli-official/paste/internal/paste"
)

const administrationWhere = `
FROM pastes AS value
WHERE (
    $1::text = '' OR
    value.code ILIKE '%' || $1::text || '%' OR
    value.title ILIKE '%' || $1::text || '%' OR
    COALESCE(value.owner_user_key, '') ILIKE '%' || $1::text || '%' OR
    array_to_string(value.tags, ' ') ILIKE '%' || $1::text || '%' OR
    EXISTS (SELECT 1 FROM paste_files AS search_file WHERE search_file.paste_id = value.id AND search_file.language ILIKE '%' || $1::text || '%')
)
AND ($2::text = '' OR value.visibility = $2::text)
AND ($3::text = '' OR value.state = $3::text)
AND (
    $4::text = '' OR
    ($4::text = 'anonymous' AND value.owner_user_key IS NULL) OR
    ($4::text = 'owned' AND value.owner_user_key IS NOT NULL)
)
AND ($5::text = '' OR value.owner_user_key = $5::text)`

func (store *Store) ListForAdministration(ctx context.Context, query paste.AdministrationQuery) (paste.AdministrationPage, error) {
	arguments := []any{strings.TrimSpace(query.Query), string(query.Visibility), string(query.State), query.Ownership, query.OwnerUserKey}
	var total int
	if err := store.database.QueryRowContext(ctx, `SELECT count(*) `+administrationWhere, arguments...).Scan(&total); err != nil {
		return paste.AdministrationPage{}, fmt.Errorf("count administrative Pastes: %w", err)
	}
	rows, err := store.database.QueryContext(ctx, `
SELECT
    value.code,
    value.owner_user_key,
    value.title,
    value.tags,
    (SELECT count(*) FROM paste_files AS file WHERE file.paste_id = value.id),
    COALESCE((SELECT file.language FROM paste_files AS file WHERE file.paste_id = value.id ORDER BY file.ordinal LIMIT 1), 'text'),
    value.visibility,
    value.password_hash IS NOT NULL,
    value.state,
    value.revision,
    value.created_at,
    value.updated_at,
    value.expires_at
`+administrationWhere+`
ORDER BY value.created_at DESC, value.code DESC
LIMIT $6 OFFSET $7`, append(arguments, query.Limit, query.Offset)...)
	if err != nil {
		return paste.AdministrationPage{}, fmt.Errorf("list administrative Pastes: %w", err)
	}
	defer rows.Close()

	items := make([]paste.AdministrationItem, 0)
	for rows.Next() {
		var item paste.AdministrationItem
		var owner sql.NullString
		var tags pq.StringArray
		var expires sql.NullTime
		if err := rows.Scan(
			&item.Code, &owner, &item.Title, &tags, &item.FileCount, &item.PrimaryLanguage,
			&item.Visibility, &item.PasswordProtected, &item.State, &item.Revision,
			&item.CreatedAt, &item.UpdatedAt, &expires,
		); err != nil {
			return paste.AdministrationPage{}, fmt.Errorf("scan administrative Paste: %w", err)
		}
		if owner.Valid {
			item.OwnerUserKey = owner.String
		}
		item.Tags = append([]string{}, tags...)
		if expires.Valid {
			value := expires.Time
			item.ExpiresAt = &value
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return paste.AdministrationPage{}, fmt.Errorf("iterate administrative Pastes: %w", err)
	}
	return paste.AdministrationPage{Items: items, Total: total, Limit: query.Limit, Offset: query.Offset}, nil
}
