package database

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"regexp"
	"testing"
	"time"

	"germinaStack/model"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestPostgresCatalogRepositoryListsYearsDeterministically(t *testing.T) {
	t.Parallel()
	db, mock := newCatalogMock(t)
	created := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)
	const query = `SELECT id, year, created_at FROM years ORDER BY year, id`
	mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(
		sqlmock.NewRows([]string{"id", "year", "created_at"}).
			AddRow(int64(2), "1st year", created).
			AddRow(int64(7), "2nd year", nil),
	)

	got, err := NewPostgresCatalogRepository(db).ListYears(context.Background())
	if err != nil {
		t.Fatalf("ListYears() error = %v", err)
	}
	want := []model.Year{{ID: 2, Year: "1st year", CreatedAt: &created}, {ID: 7, Year: "2nd year"}}
	if len(got) != len(want) || got[0].ID != want[0].ID || got[0].Year != want[0].Year || got[0].CreatedAt == nil || !got[0].CreatedAt.Equal(created) || got[1] != want[1] {
		t.Fatalf("ListYears() = %#v, want %#v", got, want)
	}
	assertCatalogExpectations(t, mock)
}

func TestPostgresCatalogRepositoryReturnsNonNilEmptyLists(t *testing.T) {
	t.Parallel()
	db, mock := newCatalogMock(t)
	mock.ExpectQuery("SELECT id, year, created_at FROM years").WillReturnRows(sqlmock.NewRows([]string{"id", "year", "created_at"}))

	got, err := NewPostgresCatalogRepository(db).ListYears(context.Background())
	if err != nil || got == nil || len(got) != 0 {
		t.Fatalf("ListYears() = (%#v, %v), want non-nil empty slice", got, err)
	}
}

func TestPostgresCatalogRepositoryListsSubjectsWithOptionalYearFilter(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		yearID *int64
		query  string
		args   []driver.Value
	}{
		{name: "all", query: `SELECT id, id_year, subject, created_at FROM subjects ORDER BY subject, id`},
		{name: "filtered", yearID: int64Pointer(7), query: `SELECT id, id_year, subject, created_at FROM subjects WHERE id_year = $1 ORDER BY subject, id`, args: []driver.Value{int64(7)}},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			db, mock := newCatalogMock(t)
			expectation := mock.ExpectQuery(regexp.QuoteMeta(tt.query))
			if len(tt.args) > 0 {
				expectation.WithArgs(tt.args...)
			}
			expectation.WillReturnRows(sqlmock.NewRows([]string{"id", "id_year", "subject", "created_at"}).AddRow(int64(9), int64(7), "Biology", nil))

			got, err := NewPostgresCatalogRepository(db).ListSubjects(context.Background(), tt.yearID)
			if err != nil || len(got) != 1 || got[0].ID != 9 || got[0].YearID == nil || *got[0].YearID != 7 || got[0].Subject != "Biology" {
				t.Fatalf("ListSubjects() = (%#v, %v)", got, err)
			}
			assertCatalogExpectations(t, mock)
		})
	}
}

func TestPostgresCatalogRepositoryYearMutationsScanAndMapErrors(t *testing.T) {
	t.Parallel()
	t.Run("create scans row", func(t *testing.T) {
		db, mock := newCatalogMock(t)
		mock.ExpectQuery("INSERT INTO years").WithArgs("2026").WillReturnRows(sqlmock.NewRows([]string{"id", "year", "created_at"}).AddRow(int64(3), "2026", nil))
		got, err := NewPostgresCatalogRepository(db).CreateYear(context.Background(), "2026")
		if err != nil || got.ID != 3 || got.Year != "2026" {
			t.Fatalf("CreateYear() = (%#v, %v)", got, err)
		}
	})

	for _, tt := range []struct {
		name string
		code string
		want error
	}{
		{name: "duplicate create", code: "23505", want: ErrCatalogConflict},
		{name: "referenced delete", code: "23503", want: ErrCatalogReferenced},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			db, mock := newCatalogMock(t)
			pgErr := &pgconn.PgError{Code: tt.code, Message: "private SQL detail"}
			var err error
			if tt.code == "23505" {
				mock.ExpectQuery("INSERT INTO years").WillReturnError(pgErr)
				_, err = NewPostgresCatalogRepository(db).CreateYear(context.Background(), "2026")
			} else {
				mock.ExpectExec("DELETE FROM years").WithArgs(int64(3)).WillReturnError(pgErr)
				err = NewPostgresCatalogRepository(db).DeleteYear(context.Background(), 3)
			}
			if !errors.Is(err, tt.want) || err.Error() != tt.want.Error() {
				t.Fatalf("error = %v, want safe %v", err, tt.want)
			}
		})
	}

	t.Run("missing update", func(t *testing.T) {
		db, mock := newCatalogMock(t)
		mock.ExpectQuery("UPDATE years").WithArgs("2026", int64(404)).WillReturnError(sql.ErrNoRows)
		_, err := NewPostgresCatalogRepository(db).UpdateYear(context.Background(), 404, "2026")
		if !errors.Is(err, ErrYearNotFound) {
			t.Fatalf("UpdateYear() error = %v, want ErrYearNotFound", err)
		}
	})

	t.Run("missing delete", func(t *testing.T) {
		db, mock := newCatalogMock(t)
		mock.ExpectExec("DELETE FROM years").WithArgs(int64(404)).WillReturnResult(sqlmock.NewResult(0, 0))
		err := NewPostgresCatalogRepository(db).DeleteYear(context.Background(), 404)
		if !errors.Is(err, ErrYearNotFound) {
			t.Fatalf("DeleteYear() error = %v, want ErrYearNotFound", err)
		}
	})
}

func TestPostgresCatalogRepositorySubjectMutationsScanAndMapErrors(t *testing.T) {
	t.Parallel()
	t.Run("create nullable subject", func(t *testing.T) {
		db, mock := newCatalogMock(t)
		mock.ExpectQuery("INSERT INTO subjects").WithArgs(nil, "General").WillReturnRows(sqlmock.NewRows([]string{"id", "id_year", "subject", "created_at"}).AddRow(int64(5), nil, "General", nil))
		got, err := NewPostgresCatalogRepository(db).CreateSubject(context.Background(), "General", nil)
		if err != nil || got.ID != 5 || got.YearID != nil || got.Subject != "General" {
			t.Fatalf("CreateSubject() = (%#v, %v)", got, err)
		}
	})

	for _, tt := range []struct {
		name string
		code string
		want error
	}{
		{name: "duplicate", code: "23505", want: ErrCatalogConflict},
		{name: "unknown related year", code: "23503", want: ErrYearNotFound},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			db, mock := newCatalogMock(t)
			mock.ExpectQuery("INSERT INTO subjects").WillReturnError(&pgconn.PgError{Code: tt.code, Message: "private SQL detail"})
			_, err := NewPostgresCatalogRepository(db).CreateSubject(context.Background(), "Biology", int64Pointer(7))
			if !errors.Is(err, tt.want) || err.Error() != tt.want.Error() {
				t.Fatalf("CreateSubject() error = %v, want safe %v", err, tt.want)
			}
		})
	}

	t.Run("missing update", func(t *testing.T) {
		db, mock := newCatalogMock(t)
		mock.ExpectQuery("UPDATE subjects").WithArgs(int64(7), "Biology", int64(404)).WillReturnError(sql.ErrNoRows)
		_, err := NewPostgresCatalogRepository(db).UpdateSubject(context.Background(), 404, "Biology", int64Pointer(7))
		if !errors.Is(err, ErrSubjectNotFound) {
			t.Fatalf("UpdateSubject() error = %v, want ErrSubjectNotFound", err)
		}
	})

	t.Run("referenced delete", func(t *testing.T) {
		db, mock := newCatalogMock(t)
		mock.ExpectExec("DELETE FROM subjects").WithArgs(int64(5)).WillReturnError(&pgconn.PgError{Code: "23503", Message: "private SQL detail"})
		err := NewPostgresCatalogRepository(db).DeleteSubject(context.Background(), 5)
		if !errors.Is(err, ErrCatalogReferenced) || err.Error() != ErrCatalogReferenced.Error() {
			t.Fatalf("DeleteSubject() error = %v, want safe ErrCatalogReferenced", err)
		}
	})
}

func TestPostgresCatalogRepositoryWrapsUnexpectedDatabaseErrors(t *testing.T) {
	t.Parallel()
	db, mock := newCatalogMock(t)
	mock.ExpectQuery("SELECT id, year, created_at FROM years").WillReturnError(context.DeadlineExceeded)
	_, err := NewPostgresCatalogRepository(db).ListYears(context.Background())
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("ListYears() error = %v, want wrapped deadline", err)
	}
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
