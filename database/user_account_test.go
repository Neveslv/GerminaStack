package database

import (
	"context"
	"regexp"
	"testing"

	"germinaStack/model"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestFindUserAccountLoadsNullableYearAndPreferencesWithoutPassword(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New(): %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	mock.ExpectQuery("SELECT u.id").WithArgs(int64(42)).WillReturnRows(sqlmock.NewRows([]string{
		"id", "id_year", "name", "username", "email", "is_admin", "contrast_theme", "font_family", "font_spacing", "font_size",
	}).AddRow(int64(42), nil, "Admin", "admin", "admin@example.test", true, "dark", nil, nil, nil))

	account, err := NewPostgresCredentialRepository(db).FindUserAccount(context.Background(), 42)
	if err != nil {
		t.Fatalf("FindUserAccount() error = %v", err)
	}
	if account.ID != 42 || account.YearID != nil || !account.IsAdmin || account.Preferences.ContrastTheme == nil || *account.Preferences.ContrastTheme != model.ContrastThemeDark {
		t.Fatalf("account = %#v", account)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("SQL expectations: %v", err)
	}
}

func TestUpdateAccessibilityPreferencesUsesUpsertAndReturnsValidatedValues(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New(): %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	dark := model.ContrastThemeDark
	preferences := AccessibilityPreferences{ContrastTheme: &dark}
	query := `INSERT INTO preferences (id_user, contrast_theme, font_family, font_spacing, font_size)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (id_user) DO UPDATE SET
    contrast_theme = EXCLUDED.contrast_theme,
    font_family = EXCLUDED.font_family,
    font_spacing = EXCLUDED.font_spacing,
    font_size = EXCLUDED.font_size
RETURNING contrast_theme, font_family, font_spacing, font_size`
	mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(int64(42), &dark, nil, nil, nil).WillReturnRows(sqlmock.NewRows([]string{
		"contrast_theme", "font_family", "font_spacing", "font_size",
	}).AddRow("dark", nil, nil, nil))

	updated, err := NewPostgresCredentialRepository(db).UpdateAccessibilityPreferences(context.Background(), 42, preferences)
	if err != nil || updated.ContrastTheme == nil || *updated.ContrastTheme != dark {
		t.Fatalf("updated preferences = %#v, error = %v", updated, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("SQL expectations: %v", err)
	}
}
