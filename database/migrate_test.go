package database

import (
	"context"
	"regexp"
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

	corePattern := regexp.QuoteMeta("CREATE TABLE IF NOT EXISTS years")
	twoFactorPattern := regexp.QuoteMeta("CREATE TABLE IF NOT EXISTS two_factor_challenges")
	mock.ExpectExec(corePattern).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(twoFactorPattern).WillReturnResult(sqlmock.NewResult(0, 0))

	if err := Migrate(context.Background(), db); err != nil {
		t.Fatalf("Migrate() error = %v", err)
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
