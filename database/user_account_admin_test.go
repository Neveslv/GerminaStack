package database

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestFindUserAccountSupportsAdminWithoutYear(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New(): %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	mock.ExpectQuery("SELECT u.id").WithArgs(int64(1)).WillReturnRows(sqlmock.NewRows([]string{
		"id", "id_year", "name", "username", "email", "is_admin", "contrast_theme", "font_family", "font_spacing", "font_size",
	}).AddRow(int64(1), nil, "Admin Jef", "admin.jef", "admin.jef@institutojef.org.br", true, nil, nil, nil, nil))

	got, err := NewPostgresCredentialRepository(db).FindUserAccount(context.Background(), 1)
	if err != nil || got.YearID != nil || !got.IsAdmin {
		t.Fatalf("admin account = (%#v, %v)", got, err)
	}
}
