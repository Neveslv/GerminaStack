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

//go:embed migrations/0005_academic_seed.sql
var academicSeedMigration string

//go:embed migrations/0006_users_admin.sql
var usersAdminMigration string

//go:embed migrations/0007_notify_mentions.sql
var notifyMentionsMigration string

//go:embed migrations/0005_message_functions_triggers.sql
var messageFunctionsMigration string

//go:embed migrations/0008_frok_user.sql
var frokUserMigration string

//go:embed migrations/0009_admin_moderation.sql
var adminModerationMigration string

//go:embed migrations/0010_notification_visibility.sql
var notificationVisibilityMigration string

func Migrate(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, coreSchemaMigration); err != nil {
		return fmt.Errorf("apply core schema migration: %w", err)
	}
	if _, err := db.ExecContext(ctx, stablePaginationIndexesMigration); err != nil {
		return fmt.Errorf("apply stable pagination index migration: %w", err)
	}
	if _, err := db.ExecContext(ctx, twoFactorChallengesMigration); err != nil {
		return fmt.Errorf("apply authentication migration: %w", err)
	}
	if _, err := db.ExecContext(ctx, usersAdminMigration); err != nil {
		return fmt.Errorf("apply users migration: %w", err)
	}
	if _, err := db.ExecContext(ctx, academicSeedMigration); err != nil {
		return fmt.Errorf("apply academic seed migration: %w", err)
	}
	if _, err := db.ExecContext(ctx, messageFunctionsMigration); err != nil {
		return fmt.Errorf("apply message functions migration: %w", err)
	}
	if _, err := db.ExecContext(ctx, notifyMentionsMigration); err != nil {
		return fmt.Errorf("apply notifications migration: %w", err)
	}
	if _, err := db.ExecContext(ctx, frokUserMigration); err != nil {
		return fmt.Errorf("apply Frok migration: %w", err)
	}
	if _, err := db.ExecContext(ctx, adminModerationMigration); err != nil {
		return fmt.Errorf("apply moderation migration: %w", err)
	}
	if _, err := db.ExecContext(ctx, notificationVisibilityMigration); err != nil {
		return fmt.Errorf("apply notification visibility migration: %w", err)
	}
	return nil
}
