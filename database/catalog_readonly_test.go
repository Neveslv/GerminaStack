package database

import (
	"context"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestCatalogReadOnlyListsOnlyYearTwo(t *testing.T) {
	db, mock := newCatalogMock(t)
	const query = `SELECT id, year, created_at FROM years WHERE year = $1 ORDER BY id`
	mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs("2").WillReturnRows(
		sqlmock.NewRows([]string{"id", "year", "created_at"}).AddRow(int64(7), "2", nil),
	)

	years, err := NewPostgresCatalogRepository(db).ListYears(context.Background())
	if err != nil || len(years) != 1 || years[0].Year != "2" {
		t.Fatalf("ListYears() = (%#v, %v), want only year 2", years, err)
	}
	assertCatalogExpectations(t, mock)
}

func TestCatalogReadOnlyListsUsersYearAndGeneral(t *testing.T) {
	db, mock := newCatalogMock(t)
	const query = `SELECT s.id, s.id_year, s.subject, s.created_at, COUNT(p.id) AS posts_count FROM subjects s LEFT JOIN posts p ON p.id_subject = s.id WHERE s.id_year = (SELECT id_year FROM users WHERE id = $1) OR (s.id_year IS NULL AND s.subject = 'Geral') GROUP BY s.id, s.id_year, s.subject, s.created_at ORDER BY s.subject, s.id`
	mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(int64(42)).WillReturnRows(
		sqlmock.NewRows([]string{"id", "id_year", "subject", "created_at", "posts_count"}).
			AddRow(int64(1), int64(7), "Biologia ESG", nil, int64(1)).
			AddRow(int64(2), nil, "Geral", nil, int64(0)),
	)

	subjects, err := NewPostgresCatalogRepository(db).ListSubjects(context.Background(), 42)
	if err != nil || len(subjects) != 2 || subjects[0].YearID == nil || *subjects[0].YearID != 7 || subjects[1].YearID != nil || subjects[1].Subject != "Geral" {
		t.Fatalf("ListSubjects() = (%#v, %v), want user's year plus Geral", subjects, err)
	}
	assertCatalogExpectations(t, mock)
}

func TestAcademicCatalogMigrationIsIdempotentByContract(t *testing.T) {
	migration, err := os.ReadFile("migrations/0004_admin_year_and_academic_seed.sql")
	if err != nil {
		t.Fatalf("read academic catalog migration: %v", err)
	}
	sql := strings.Join(strings.Fields(string(migration)), " ")
	for _, fragment := range []string{
		"INSERT INTO years (year) VALUES ('2') ON CONFLICT (year) DO NOTHING",
		"ON CONFLICT (id_year, subject) DO NOTHING",
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_subjects_null_year_subject ON subjects (subject) WHERE id_year IS NULL",
		"WHERE NOT EXISTS (",
		"SELECT NULL, 'Geral'",
		"('Biologia ESG')",
		"('Engenharia e Qualidade de Software')",
	} {
		if !strings.Contains(sql, fragment) {
			t.Errorf("migration is missing idempotent seed fragment %q", fragment)
		}
	}
}
