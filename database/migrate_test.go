package database

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestMigrateExecutesFiveStagesInOrderOnEveryRun(t *testing.T) {
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
	mock.ExpectExec(corePattern).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(upgradePattern).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(twoFactorPattern).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("(?s)INSERT INTO years").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("(?s)CREATE OR REPLACE FUNCTION create_message").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(corePattern).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(upgradePattern).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(twoFactorPattern).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("(?s)INSERT INTO years").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("(?s)CREATE OR REPLACE FUNCTION create_message").WillReturnResult(sqlmock.NewResult(0, 0))

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
