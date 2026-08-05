package database

import (
	"context"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestMigrateExecutesCoreSchemaBeforeTwoFactorChallengeSchema(t *testing.T) {
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
	mock.ExpectExec(corePattern).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(upgradePattern).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(twoFactorPattern).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(academicSeedPattern).WillReturnResult(sqlmock.NewResult(0, 0))

	if err := Migrate(context.Background(), db); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}
	if err := Migrate(context.Background(), db); err != nil {
		t.Fatalf("second Migrate() error = %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("SQL expectations: %v", err)
	}
}

func TestAcademicSeedMigrationDefinesRequiredCatalogAndAdminInvariant(t *testing.T) {
	t.Parallel()

	migration, err := os.ReadFile("migrations/0004_admin_year_and_academic_seed.sql")
	if err != nil {
		t.Fatalf("read academic migration: %v", err)
	}
	sql := string(migration)

	for _, required := range []string{
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS is_admin BOOLEAN NOT NULL DEFAULT FALSE;",
		"ALTER COLUMN id_year DROP NOT NULL",
		"UPDATE users SET id_year = NULL WHERE is_admin = TRUE;",
		"ADD CONSTRAINT users_admin_year_check",
		"CHECK ((is_admin = TRUE AND id_year IS NULL) OR (is_admin = FALSE AND id_year IS NOT NULL))",
		"INSERT INTO years (year) VALUES ('2') ON CONFLICT (year) DO NOTHING;",
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_subjects_null_year_subject ON subjects (subject) WHERE id_year IS NULL;",
		"INSERT INTO subjects (id_year, subject)",
		"VALUES (NULL, 'Geral')",
	} {
		if !strings.Contains(sql, required) {
			t.Errorf("migration is missing %q", required)
		}
	}

	wantSubjects := []string{
		"Biologia ESG", "Desenvolvimento", "DAD", "Mobile", "DevOps", "Educação Física",
		"Engenharia e Qualidade de Software", "Estatística", "BI", "Língua Inglesa", "Chefia e Liderança",
		"Língua Portuguesa", "Matemática", "Modelagem de Dados", "Inteligência Artificial", "Banco de Dados",
		"Sociologia", "UX",
	}
	for _, subject := range wantSubjects {
		if !strings.Contains(sql, "'"+subject+"'") {
			t.Errorf("migration is missing subject %q", subject)
		}
	}
	if strings.Contains(sql, "Aula Prática–Estágio") {
		t.Error("migration must not seed Aula Prática–Estágio")
	}
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
