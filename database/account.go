package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"germinaStack/model"

	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrAccountNotFound     = errors.New("account not found")
	ErrPreferencesNotFound = errors.New("preferences not found")
)

type PostgresAccountRepository struct{ db *sql.DB }

func (r *PostgresAccountRepository) ListMentionUsers(ctx context.Context, prefix string) ([]model.User, error) {
	const query = `SELECT id, name, username, profile_image_url, profile_image_description
FROM users
WHERE username ILIKE $1 || '%'
ORDER BY username
LIMIT 8`

	rows, err := r.db.QueryContext(ctx, query, prefix)
	if err != nil {
		return nil, fmt.Errorf("list mention users: %w", err)
	}
	defer rows.Close()

	users := make([]model.User, 0)
	for rows.Next() {
		var user model.User
		if err := rows.Scan(&user.ID, &user.Name, &user.Username, &user.ProfileImageURL, &user.ProfileImageDescription); err != nil {
			return nil, fmt.Errorf("scan mention user: %w", err)
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate mention users: %w", err)
	}
	return users, nil
}

func NewPostgresAccountRepository(db *sql.DB) *PostgresAccountRepository {
	return &PostgresAccountRepository{db: db}
}

func (r *PostgresAccountRepository) GetProfile(ctx context.Context, userID int64) (model.User, error) {
	const query = `SELECT id, id_year, name, profile_image_url, profile_image_description, username, email, is_admin, created_at
FROM users
WHERE id = $1`
	var user model.User
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&user.ID, &user.YearID, &user.Name, &user.ProfileImageURL,
		&user.ProfileImageDescription, &user.Username, &user.Email,
		&user.IsAdmin, &user.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return model.User{}, ErrAccountNotFound
	}
	if err != nil {
		return model.User{}, fmt.Errorf("get profile: %w", err)
	}
	return user, nil
}

func (r *PostgresAccountRepository) GetPublicProfile(ctx context.Context, username string) (model.User, error) {
	const query = `SELECT id, id_year, name, profile_image_url, profile_image_description, username, created_at
FROM users
WHERE username = $1`
	var user model.User
	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID, &user.YearID, &user.Name, &user.ProfileImageURL,
		&user.ProfileImageDescription, &user.Username, &user.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return model.User{}, ErrAccountNotFound
	}
	if err != nil {
		return model.User{}, fmt.Errorf("get public profile: %w", err)
	}
	return user, nil
}

func (r *PostgresAccountRepository) UpdateProfile(ctx context.Context, userID int64, user model.User) (model.User, error) {
	const query = `UPDATE users
SET name = $1, profile_image_url = $2, profile_image_description = $3
WHERE id = $4
RETURNING id, id_year, name, profile_image_url, profile_image_description, username, email, is_admin, created_at`
	var updated model.User
	err := r.db.QueryRowContext(ctx, query, user.Name, user.ProfileImageURL, user.ProfileImageDescription, userID).Scan(
		&updated.ID, &updated.YearID, &updated.Name, &updated.ProfileImageURL,
		&updated.ProfileImageDescription, &updated.Username, &updated.Email,
		&updated.IsAdmin, &updated.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return model.User{}, ErrAccountNotFound
	}
	if err != nil {
		return model.User{}, fmt.Errorf("update profile: %w", err)
	}
	return updated, nil
}

func (r *PostgresAccountRepository) GetPreferences(ctx context.Context, userID int64) (model.Preference, error) {
	const query = `SELECT id, id_user, contrast_theme, font_family, font_spacing, font_size, created_at
FROM preferences
WHERE id_user = $1`
	var preference model.Preference
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&preference.ID, &preference.UserID, &preference.ContrastTheme,
		&preference.FontFamily, &preference.FontSpacing, &preference.FontSize,
		&preference.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Preference{}, ErrPreferencesNotFound
	}
	if err != nil {
		return model.Preference{}, fmt.Errorf("get preferences: %w", err)
	}
	return preference, nil
}

func (r *PostgresAccountRepository) UpsertPreferences(ctx context.Context, userID int64, preference model.Preference) (model.Preference, error) {
	const query = `INSERT INTO preferences (id_user, contrast_theme, font_family, font_spacing, font_size)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (id_user) DO UPDATE SET contrast_theme = EXCLUDED.contrast_theme,
    font_family = EXCLUDED.font_family, font_spacing = EXCLUDED.font_spacing, font_size = EXCLUDED.font_size
RETURNING id, id_user, contrast_theme, font_family, font_spacing, font_size, created_at`
	var updated model.Preference
	err := r.db.QueryRowContext(ctx, query, userID, preference.ContrastTheme, preference.FontFamily, preference.FontSpacing, preference.FontSize).Scan(
		&updated.ID, &updated.UserID, &updated.ContrastTheme, &updated.FontFamily,
		&updated.FontSpacing, &updated.FontSize, &updated.CreatedAt,
	)
	if err != nil {
		var postgresError *pgconn.PgError
		if errors.As(err, &postgresError) && postgresError.Code == "23503" {
			return model.Preference{}, ErrAccountNotFound
		}
		return model.Preference{}, fmt.Errorf("upsert preferences: %w", err)
	}
	return updated, nil
}
