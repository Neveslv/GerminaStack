package database

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestPostgresCatalogRepositoryListsOnlyYearTwoDeterministically(t *testing.T) {
	t.Parallel()
	db, mock := newCatalogMock(t)
	created := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)
	const query = `SELECT id, year, created_at FROM years WHERE year = $1 ORDER BY id`
	mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs("2").WillReturnRows(
		sqlmock.NewRows([]string{"id", "year", "created_at"}).
			AddRow(int64(7), "2", created).
			AddRow(int64(9), "2", nil),
	)

	got, err := NewPostgresCatalogRepository(db).ListYears(context.Background())
	if err != nil || len(got) != 2 || got[0].ID != 7 || got[0].Year != "2" || got[0].CreatedAt == nil || !got[0].CreatedAt.Equal(created) {
		t.Fatalf("ListYears() = (%#v, %v), want only year 2", got, err)
	}
	assertCatalogExpectations(t, mock)
}

func TestPostgresCatalogRepositoryReturnsNonNilEmptyLists(t *testing.T) {
	t.Parallel()
	db, mock := newCatalogMock(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, year, created_at FROM years WHERE year = $1 ORDER BY id`)).
		WithArgs("2").
		WillReturnRows(sqlmock.NewRows([]string{"id", "year", "created_at"}))

	got, err := NewPostgresCatalogRepository(db).ListYears(context.Background())
	if err != nil || got == nil || len(got) != 0 {
		t.Fatalf("ListYears() = (%#v, %v), want non-nil empty slice", got, err)
	}
	assertCatalogExpectations(t, mock)
}

func TestPostgresCatalogRepositoryListsUsersYearAndGeneral(t *testing.T) {
	t.Parallel()
	db, mock := newCatalogMock(t)
	const query = `SELECT id, id_year, subject, created_at FROM subjects WHERE id_year = (SELECT id_year FROM users WHERE id = $1) OR (id_year IS NULL AND subject = 'Geral') ORDER BY subject, id`
	mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(int64(42)).WillReturnRows(
		sqlmock.NewRows([]string{"id", "id_year", "subject", "created_at"}).
			AddRow(int64(1), int64(7), "Biology", nil).
			AddRow(int64(2), nil, "Geral", nil),
	)

	got, err := NewPostgresCatalogRepository(db).ListSubjects(context.Background(), 42)
	if err != nil || len(got) != 2 || got[0].YearID == nil || *got[0].YearID != 7 || got[1].YearID != nil || got[1].Subject != "Geral" {
		t.Fatalf("ListSubjects() = (%#v, %v), want user's year plus Geral", got, err)
	}
	assertCatalogExpectations(t, mock)
}

func TestPostgresCatalogRepositoryWrapsUnexpectedDatabaseErrors(t *testing.T) {
	t.Parallel()
	db, mock := newCatalogMock(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, year, created_at FROM years WHERE year = $1 ORDER BY id`)).
		WithArgs("2").
		WillReturnError(context.DeadlineExceeded)

	_, err := NewPostgresCatalogRepository(db).ListYears(context.Background())
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("ListYears() error = %v, want wrapped deadline", err)
	}
	assertCatalogExpectations(t, mock)
}

func newCatalogMock(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New(): %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db, mock
}

func assertCatalogExpectations(t *testing.T, mock sqlmock.Sqlmock) {
	t.Helper()
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("SQL expectations: %v", err)
	}
}

func int64Pointer(value int64) *int64 { return &value }
