package database

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"germinaStack/model"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestPostgresAccountRepositoryReadsAndUpdatesProfileWithoutPassword(t *testing.T) {
	t.Parallel()
	db, mock := newCatalogMock(t)
	const selectQuery = `SELECT id, id_year, name, profile_image_url, profile_image_description, username, email, is_admin, created_at
FROM users
WHERE id = $1`
	mock.ExpectQuery(regexp.QuoteMeta(selectQuery)).WithArgs(int64(42)).WillReturnRows(
		sqlmock.NewRows([]string{"id", "id_year", "name", "profile_image_url", "profile_image_description", "username", "email", "is_admin", "created_at"}).
			AddRow(int64(42), int64(2), "Ana", nil, nil, "ana.silva", "ana@example.org", false, nil),
	)
	const updateQuery = `UPDATE users
SET name = $1, profile_image_url = $2, profile_image_description = $3
WHERE id = $4
RETURNING id, id_year, name, profile_image_url, profile_image_description, username, email, is_admin, created_at`
	mock.ExpectQuery(regexp.QuoteMeta(updateQuery)).WithArgs("Ana Silva", nil, nil, int64(42)).WillReturnRows(
		sqlmock.NewRows([]string{"id", "id_year", "name", "profile_image_url", "profile_image_description", "username", "email", "is_admin", "created_at"}).
			AddRow(int64(42), int64(2), "Ana Silva", nil, nil, "ana.silva", "ana@example.org", false, nil),
	)
	repository := NewPostgresAccountRepository(db)
	profile, err := repository.GetProfile(context.Background(), 42)
	if err != nil || profile.Name != "Ana" || profile.Password != "" {
		t.Fatalf("GetProfile() = (%#v, %v)", profile, err)
	}
	profile.Name = "Ana Silva"
	updated, err := repository.UpdateProfile(context.Background(), 42, profile)
	if err != nil || updated.Name != "Ana Silva" {
		t.Fatalf("UpdateProfile() = (%#v, %v)", updated, err)
	}
	assertCatalogExpectations(t, mock)
}

func TestPostgresAccountRepositoryReadsPublicProfileWithoutEmail(t *testing.T) {
	t.Parallel()
	db, mock := newCatalogMock(t)
	const query = `SELECT id, id_year, name, profile_image_url, profile_image_description, username, created_at
FROM users
WHERE username = $1`
	mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs("ana.silva").WillReturnRows(
		sqlmock.NewRows([]string{"id", "id_year", "name", "profile_image_url", "profile_image_description", "username", "created_at"}).
			AddRow(int64(42), int64(7), "Ana Silva", nil, nil, "ana.silva", nil),
	)
	profile, err := NewPostgresAccountRepository(db).GetPublicProfile(context.Background(), "ana.silva")
	if err != nil || profile.ID != 42 || profile.Username != "ana.silva" || profile.Email != "" {
		t.Fatalf("GetPublicProfile() = (%#v, %v)", profile, err)
	}
	assertCatalogExpectations(t, mock)
}

func TestPostgresAccountRepositoryUpsertsPreferencesAndMapsMissingUser(t *testing.T) {
	t.Parallel()
	db, mock := newCatalogMock(t)
	fontSize := model.FontSizeGrande
	const query = `INSERT INTO preferences (id_user, contrast_theme, font_family, font_spacing, font_size)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (id_user) DO UPDATE SET contrast_theme = EXCLUDED.contrast_theme,
    font_family = EXCLUDED.font_family, font_spacing = EXCLUDED.font_spacing, font_size = EXCLUDED.font_size
RETURNING id, id_user, contrast_theme, font_family, font_spacing, font_size, created_at`
	mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(int64(42), nil, nil, nil, string(fontSize)).WillReturnRows(
		sqlmock.NewRows([]string{"id", "id_user", "contrast_theme", "font_family", "font_spacing", "font_size", "created_at"}).
			AddRow(int64(1), int64(42), nil, nil, nil, string(fontSize), nil),
	)
	if _, err := NewPostgresAccountRepository(db).UpsertPreferences(context.Background(), 42, model.Preference{FontSize: &fontSize}); err != nil {
		t.Fatalf("UpsertPreferences() error = %v", err)
	}
	assertCatalogExpectations(t, mock)

	db, mock = newCatalogMock(t)
	mock.ExpectQuery("INSERT INTO preferences").WillReturnError(context.DeadlineExceeded)
	if _, err := NewPostgresAccountRepository(db).UpsertPreferences(context.Background(), 42, model.Preference{}); err == nil || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("UpsertPreferences() error = %v", err)
	}
}
