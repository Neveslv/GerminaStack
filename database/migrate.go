package database

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
)

//go:embed migrations/0001_two_factor_challenges.sql
var twoFactorChallengesMigration string

func Migrate(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, twoFactorChallengesMigration); err != nil {
		return fmt.Errorf("apply authentication migration: %w", err)
	}
	return nil
}
