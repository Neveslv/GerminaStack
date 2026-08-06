package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"germinaStack/model"

	"github.com/jackc/pgx/v5/pgconn"
)

type AccessibilityPreferences struct {
	ContrastTheme *model.ContrastTheme
	FontFamily    *model.FontFamily
	FontSpacing   *model.FontSpacing
	FontSize      *model.FontSize
}

type UserAccount struct {
	ID          int64
	YearID      *int64
	Name        string
	Username    string
	Email       string
	IsAdmin     bool
	Preferences AccessibilityPreferences
}

func (r *PostgresCredentialRepository) FindUserAccount(ctx context.Context, userID int64) (UserAccount, error) {
	const query = `SELECT u.id, u.id_year, u.name, u.username, u.email, u.is_admin,
       p.contrast_theme, p.font_family, p.font_spacing, p.font_size
FROM users u
LEFT JOIN preferences p ON p.id_user = u.id
WHERE u.id = $1`

	var account UserAccount
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&account.ID, &account.YearID, &account.Name, &account.Username, &account.Email, &account.IsAdmin, &account.Preferences.ContrastTheme, &account.Preferences.FontFamily, &account.Preferences.FontSpacing, &account.Preferences.FontSize)
	if errors.Is(err, sql.ErrNoRows) {
		return UserAccount{}, ErrCredentialNotFound
	}
	if err != nil {
		return UserAccount{}, fmt.Errorf("find user account: %w", err)
	}
	return account, nil
}

func (r *PostgresCredentialRepository) UpdateAccessibilityPreferences(ctx context.Context, userID int64, preferences AccessibilityPreferences) (AccessibilityPreferences, error) {
	const query = `INSERT INTO preferences (id_user, contrast_theme, font_family, font_spacing, font_size)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (id_user) DO UPDATE SET
    contrast_theme = EXCLUDED.contrast_theme,
    font_family = EXCLUDED.font_family,
    font_spacing = EXCLUDED.font_spacing,
    font_size = EXCLUDED.font_size
RETURNING contrast_theme, font_family, font_spacing, font_size`

	var updated AccessibilityPreferences
	err := r.db.QueryRowContext(ctx, query, userID, preferences.ContrastTheme, preferences.FontFamily, preferences.FontSpacing, preferences.FontSize).Scan(&updated.ContrastTheme, &updated.FontFamily, &updated.FontSpacing, &updated.FontSize)
	if err == nil {
		return updated, nil
	}
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) && postgresError.Code == "23503" {
		return AccessibilityPreferences{}, ErrCredentialNotFound
	}
	return AccessibilityPreferences{}, fmt.Errorf("update accessibility preferences: %w", err)
}
