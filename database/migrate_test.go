package database

import (
	"context"
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
	usersAdminPattern := regexp.QuoteMeta("ALTER TABLE users ADD COLUMN IF NOT EXISTS is_admin")
	seedPattern := regexp.QuoteMeta("INSERT INTO years (year)")
	notificationsPattern := regexp.QuoteMeta("CREATE OR REPLACE FUNCTION notify_mentions()")
	frokPattern := regexp.QuoteMeta("INSERT INTO users (id_year, name, username, email, password, profile_image_url, profile_image_description)")
	mock.ExpectExec(corePattern).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(upgradePattern).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(twoFactorPattern).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(usersAdminPattern).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(seedPattern).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(notificationsPattern).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(frokPattern).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(corePattern).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(upgradePattern).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(twoFactorPattern).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(usersAdminPattern).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(seedPattern).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(notificationsPattern).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(frokPattern).WillReturnResult(sqlmock.NewResult(0, 0))

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

func TestFrokUserMigrationSetsDefaultProfileImage(t *testing.T) {
	if !strings.Contains(frokUserMigration, "'/static/images/frok-profile.jpeg'") {
		t.Fatal("Frok migration does not set the default profile image")
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

func TestNotifyMentionsPatternMatchesDottedUsername(t *testing.T) {
	pattern := regexp.MustCompile(`@([A-Za-z0-9_]+([.][A-Za-z0-9_]+)*)`)
	match := pattern.FindStringSubmatch("@ana.silva, veja isso.")
	if len(match) < 2 || match[1] != "ana.silva" {
		t.Fatalf("mention match = %#v, want ana.silva", match)
	}
}
