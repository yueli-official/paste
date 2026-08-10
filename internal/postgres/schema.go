package postgres

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
)

//go:embed migrations/0001_pastes.up.sql
var schemaV1 string

func ApplySchema(ctx context.Context, database *sql.DB) error {
	if database == nil {
		return errors.New("paste/postgres: DB is required")
	}
	_, err := database.ExecContext(ctx, schemaV1)
	return err
}
