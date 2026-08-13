package database

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestCatalogRepositoryIncludesOnlyTheGeneralGlobalSubject(t *testing.T) {
	db, mock := newCatalogMock(t)
	const query = `SELECT s.id, s.id_year, s.subject, s.created_at, COUNT(p.id) AS posts_count FROM subjects s LEFT JOIN posts p ON p.id_subject = s.id WHERE s.id_year = (SELECT id_year FROM users WHERE id = $1) OR (s.id_year IS NULL AND s.subject = 'Geral') GROUP BY s.id, s.id_year, s.subject, s.created_at ORDER BY s.subject, s.id`
	mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(int64(42)).WillReturnRows(
		sqlmock.NewRows([]string{"id", "id_year", "subject", "created_at", "posts_count"}).AddRow(int64(2), nil, "Geral", nil, int64(0)),
	)

	subjects, err := NewPostgresCatalogRepository(db).ListSubjects(context.Background(), 42)
	if err != nil || len(subjects) != 1 || subjects[0].Subject != "Geral" {
		t.Fatalf("ListSubjects() = (%#v, %v), want only Geral globally", subjects, err)
	}
	assertCatalogExpectations(t, mock)
}
