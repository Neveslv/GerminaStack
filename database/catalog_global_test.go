package database

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestCatalogRepositoryIncludesOnlyTheGeneralGlobalSubject(t *testing.T) {
	db, mock := newCatalogMock(t)
	const query = `SELECT id, id_year, subject, created_at FROM subjects WHERE id_year = (SELECT id_year FROM users WHERE id = $1) OR (id_year IS NULL AND subject = 'Geral') ORDER BY subject, id`
	mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(int64(42)).WillReturnRows(
		sqlmock.NewRows([]string{"id", "id_year", "subject", "created_at"}).AddRow(int64(2), nil, "Geral", nil),
	)

	subjects, err := NewPostgresCatalogRepository(db).ListSubjects(context.Background(), 42)
	if err != nil || len(subjects) != 1 || subjects[0].Subject != "Geral" {
		t.Fatalf("ListSubjects() = (%#v, %v), want only Geral globally", subjects, err)
	}
	assertCatalogExpectations(t, mock)
}
