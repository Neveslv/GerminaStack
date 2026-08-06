package database

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
)

//go:embed migrations/0002_core_schema.sql
var coreSchemaMigration string

//go:embed migrations/0003_stable_pagination_indexes.sql
var stablePaginationIndexesMigration string

//go:embed migrations/0001_two_factor_challenges.sql
var twoFactorChallengesMigration string

//go:embed migrations/0004_admin_year_and_academic_seed.sql
var academicCatalogMigration string

//go:embed migrations/0005_message_functions_triggers.sql
var messageFunctionsMigration string

func Migrate(ctx context.Context, db *sql.DB) error {
	migrations := []struct {
		name string
		sql  string
	}{
		{name: "core schema", sql: coreSchemaMigration},
		{name: "stable pagination index", sql: stablePaginationIndexesMigration},
		{name: "authentication", sql: twoFactorChallengesMigration},
		{name: "academic catalog", sql: academicCatalogMigration},
		{name: "message functions", sql: messageFunctionsMigration},
	}
	for _, migration := range migrations {
		if _, err := db.ExecContext(ctx, migration.sql); err != nil {
			return fmt.Errorf("apply %s migration: %w", migration.name, err)
		}
	}
	return nil
}
