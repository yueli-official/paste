package postgres

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
)

//go:embed migrations/0001_pastes.up.sql
var schemaV1 string

//go:embed migrations/0002_site_settings.up.sql
var schemaV2 string

//go:embed migrations/0003_governance.up.sql
var schemaV3 string

//go:embed migrations/0004_authorization.up.sql
var schemaV4 string

//go:embed migrations/0005_audit.up.sql
var schemaV5 string

//go:embed migrations/0006_authorization_catalog_v2.up.sql
var schemaV6 string

func ApplySchema(ctx context.Context, database *sql.DB) error {
	if database == nil {
		return errors.New("paste/postgres: DB is required")
	}
	_, err := database.ExecContext(ctx, schemaV1+"\n"+schemaV2+"\n"+schemaV3+"\n"+schemaV4+"\n"+schemaV5+"\n"+schemaV6)
	return err
}
