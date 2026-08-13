package database

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestMigrateExecutesAcademicSeedAndMessageFunctionsAfterAuthentication(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New(): %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	mock.ExpectExec("(?s)CREATE TABLE IF NOT EXISTS years").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("DO $$")).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("(?s)CREATE TABLE IF NOT EXISTS two_factor_challenges").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("(?s)ALTER TABLE users ADD COLUMN IF NOT EXISTS is_admin").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("(?s)INSERT INTO years").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("(?s)CREATE OR REPLACE FUNCTION create_message").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("(?s)CREATE OR REPLACE FUNCTION notify_mentions").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("(?s)INSERT INTO users").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("(?s)ALTER TABLE users ADD COLUMN IF NOT EXISTS is_banned").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("(?s)ALTER TABLE notifications ADD COLUMN IF NOT EXISTS is_hidden").WillReturnResult(sqlmock.NewResult(0, 0))

	if err := Migrate(context.Background(), db); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("migration order expectations: %v", err)
	}
}
