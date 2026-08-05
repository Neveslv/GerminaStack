package database

import (
	"context"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestMigrateExecutesAcademicMigrationAfterExistingMigrations(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New(): %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	corePattern := "(?s)" +
		regexp.QuoteMeta("CREATE INDEX IF NOT EXISTS idx_posts_created ON posts (created_at DESC, id DESC);") + ".*" +
		regexp.QuoteMeta("CREATE INDEX IF NOT EXISTS idx_posts_subject_created ON posts (id_subject, created_at DESC, id DESC);") + ".*" +
		regexp.QuoteMeta("CREATE INDEX IF NOT EXISTS idx_posts_user_created ON posts (id_user, created_at DESC, id DESC);") + ".*" +
		regexp.QuoteMeta("CREATE INDEX IF NOT EXISTS idx_comments_post_created ON comments (id_post, created_at DESC, id DESC);") + ".*" +
		regexp.QuoteMeta("CREATE INDEX IF NOT EXISTS idx_comments_on_comments_comment_created ON comments_on_comments (id_comment, created_at DESC, id DESC);") + ".*" +
		regexp.QuoteMeta("CREATE INDEX IF NOT EXISTS idx_notifications_user_created ON notifications (id_user, created_at DESC, id DESC);") + ".*" +
		regexp.QuoteMeta("CREATE INDEX IF NOT EXISTS idx_notifications_user_unread_created ON notifications (id_user, created_at DESC, id DESC) WHERE is_read = FALSE;")
	upgradePattern := "(?s)" +
		regexp.QuoteMeta("DO $$") + ".*" +
		regexp.QuoteMeta("CREATE INDEX IF NOT EXISTS idx_posts_subject_created ON posts (id_subject, created_at DESC, id DESC);") + ".*" +
		regexp.QuoteMeta("CREATE INDEX IF NOT EXISTS idx_posts_user_created ON posts (id_user, created_at DESC, id DESC);") + ".*" +
		regexp.QuoteMeta("CREATE INDEX IF NOT EXISTS idx_comments_post_created ON comments (id_post, created_at DESC, id DESC);") + ".*" +
		regexp.QuoteMeta("CREATE INDEX IF NOT EXISTS idx_comments_on_comments_comment_created ON comments_on_comments (id_comment, created_at DESC, id DESC);")
	twoFactorPattern := regexp.QuoteMeta("CREATE TABLE IF NOT EXISTS two_factor_challenges")
	academicSeedPattern := "(?s)" +
		regexp.QuoteMeta("ALTER TABLE users ADD COLUMN IF NOT EXISTS is_admin BOOLEAN NOT NULL DEFAULT FALSE;") + ".*" +
		regexp.QuoteMeta("ADD CONSTRAINT users_admin_year_check") + ".*" +
		regexp.QuoteMeta("CREATE UNIQUE INDEX IF NOT EXISTS idx_subjects_null_year_subject")
	mock.ExpectExec(corePattern).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(upgradePattern).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(twoFactorPattern).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(academicSeedPattern).WillReturnResult(sqlmock.NewResult(0, 0))

	if err := Migrate(context.Background(), db); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("SQL expectations: %v", err)
	}
}

func TestAcademicSeedMigrationSeedsExactlyRequiredSubjects(t *testing.T) {
	t.Parallel()

	sql := readAcademicSeedMigration(t)
	for _, required := range []string{
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS is_admin BOOLEAN NOT NULL DEFAULT FALSE;",
		"ALTER COLUMN id_year DROP NOT NULL",
		"UPDATE users SET id_year = NULL WHERE is_admin = TRUE;",
		"ADD CONSTRAINT users_admin_year_check",
		"CHECK ((is_admin = TRUE AND id_year IS NULL) OR (is_admin = FALSE AND id_year IS NOT NULL))",
	} {
		if !strings.Contains(sql, required) {
			t.Errorf("migration is missing %q", required)
		}
	}

	seedStart := strings.Index(sql, "CROSS JOIN (\n    VALUES")
	seedEnd := strings.Index(sql, ") AS academic_subjects(subject)")
	if seedStart < 0 || seedEnd < 0 || seedEnd <= seedStart {
		t.Fatal("academic subject seed block not found")
	}
	seedBlock := sql[seedStart:seedEnd]
	seedRows := regexp.MustCompile(`(?m)^\s*\('([^']+)'\),?$`).FindAllStringSubmatch(seedBlock, -1)

	wantSubjects := map[string]bool{
		"Biologia ESG": true, "Desenvolvimento": true, "DAD": true, "Mobile": true, "DevOps": true,
		"Educação Física": true, "Engenharia e Qualidade de Software": true, "Estatística": true,
		"BI": true, "Língua Inglesa": true, "Chefia e Liderança": true, "Língua Portuguesa": true,
		"Matemática": true, "Modelagem de Dados": true, "Inteligência Artificial": true,
		"Banco de Dados": true, "Sociologia": true, "UX": true,
	}
	if got, want := len(seedRows), len(wantSubjects); got != want {
		t.Fatalf("academic subject seed count = %d, want %d", got, want)
	}

	seen := make(map[string]int, len(seedRows))
	for _, row := range seedRows {
		seen[row[1]]++
		if !wantSubjects[row[1]] {
			t.Errorf("unexpected academic subject %q", row[1])
		}
	}
	for subject := range wantSubjects {
		if got := seen[subject]; got != 1 {
			t.Errorf("academic subject %q appears %d times, want once", subject, got)
		}
	}
	if strings.Contains(sql, "Aula Prática–Estágio") {
		t.Error("migration must not seed Aula Prática–Estágio")
	}
}

func TestAcademicSeedMigrationUsesConflictSafeSeeds(t *testing.T) {
	t.Parallel()

	sql := readAcademicSeedMigration(t)
	if got := strings.Count(sql, "INSERT INTO subjects (id_year, subject)"); got != 2 {
		t.Fatalf("subject seed statement count = %d, want 2", got)
	}
	if got := strings.Count(sql, "INSERT INTO years (year) VALUES ('2') ON CONFLICT (year) DO NOTHING;"); got != 1 {
		t.Fatalf("year 2 conflict-safe seed count = %d, want 1", got)
	}

	subjectSeedStart := strings.Index(sql, "INSERT INTO subjects (id_year, subject)\nSELECT years.id")
	subjectSeedEnd := strings.Index(sql, "CREATE UNIQUE INDEX IF NOT EXISTS idx_subjects_null_year_subject")
	if subjectSeedStart < 0 || subjectSeedEnd < 0 || subjectSeedEnd <= subjectSeedStart {
		t.Fatal("academic subject seed statement not found")
	}
	subjectSeed := sql[subjectSeedStart:subjectSeedEnd]
	if got := strings.Count(subjectSeed, "ON CONFLICT (id_year, subject) DO NOTHING;"); got != 1 {
		t.Fatalf("academic subject conflict clause count = %d, want 1", got)
	}

	generalSeed := regexp.MustCompile(`(?s)INSERT INTO subjects \(id_year, subject\)\s*SELECT NULL, 'Geral'\s*WHERE NOT EXISTS \(\s*SELECT 1\s+FROM subjects\s+WHERE id_year IS NULL AND subject = 'Geral'\s*\);`)
	if got := len(generalSeed.FindAllStringIndex(sql, -1)); got != 1 {
		t.Fatalf("Geral idempotent seed count = %d, want 1", got)
	}
	if got := strings.Count(sql, "CREATE UNIQUE INDEX IF NOT EXISTS idx_subjects_null_year_subject ON subjects (subject) WHERE id_year IS NULL;"); got != 1 {
		t.Fatalf("null-year partial unique index count = %d, want 1", got)
	}
}

func readAcademicSeedMigration(t *testing.T) string {
	t.Helper()

	migration, err := os.ReadFile("migrations/0004_admin_year_and_academic_seed.sql")
	if err != nil {
		t.Fatalf("read academic migration: %v", err)
	}
	return string(migration)
}

func TestMigrateReturnsDatabaseError(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New(): %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	mock.ExpectExec("CREATE TABLE").WillReturnError(context.DeadlineExceeded)

	if err := Migrate(context.Background(), db); err == nil {
		t.Fatal("Migrate() error = nil, want non-nil")
	}
}
