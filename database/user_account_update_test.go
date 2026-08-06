package database

import (
	"context"
	"errors"
	"germinaStack/model"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgx/v5/pgconn"
	"regexp"
	"testing"
)

func TestUpdateAccessibilityPreferencesTouchesOnlyPreferenceColumns(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	const query = `INSERT INTO preferences (id_user, contrast_theme, font_family, font_spacing, font_size)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (id_user) DO UPDATE SET
    contrast_theme = EXCLUDED.contrast_theme,
    font_family = EXCLUDED.font_family,
    font_spacing = EXCLUDED.font_spacing,
    font_size = EXCLUDED.font_size
RETURNING contrast_theme, font_family, font_spacing, font_size`
	dark := model.ContrastThemeDark
	arial := model.FontFamilyArial
	spacing := model.FontSpacingGrande
	size := model.FontSizeGrande
	want := AccessibilityPreferences{ContrastTheme: &dark, FontFamily: &arial, FontSpacing: &spacing, FontSize: &size}
	mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(int64(42), &dark, &arial, &spacing, &size).WillReturnRows(sqlmock.NewRows([]string{"contrast_theme", "font_family", "font_spacing", "font_size"}).AddRow("dark", "arial", "grande", "grande"))
	got, err := NewPostgresCredentialRepository(db).UpdateAccessibilityPreferences(context.Background(), 42, want)
	if err != nil || got.ContrastTheme == nil || *got.ContrastTheme != dark || got.FontFamily == nil || *got.FontFamily != arial || got.FontSpacing == nil || *got.FontSpacing != spacing || got.FontSize == nil || *got.FontSize != size {
		t.Fatalf("updated preferences = %#v, err=%v", got, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUpdateAccessibilityPreferencesMapsMissingUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	mock.ExpectQuery("INSERT INTO preferences").WillReturnError(&pgconn.PgError{Code: "23503", Message: "private detail"})
	_, err = NewPostgresCredentialRepository(db).UpdateAccessibilityPreferences(context.Background(), 404, AccessibilityPreferences{})
	if !errors.Is(err, ErrCredentialNotFound) || err.Error() != ErrCredentialNotFound.Error() {
		t.Fatalf("error=%v", err)
	}
}
