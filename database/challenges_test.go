package database

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

const challengeSelectForUpdate = `SELECT id, user_id, code_hash, expires_at, attempts, max_attempts, used_at, created_at
FROM two_factor_challenges
WHERE id = $1
FOR UPDATE`

func TestPostgresChallengeRepositoryCreateInvalidatesPreviousAndPersists(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New(): %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	now := time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC)
	challenge := Challenge{
		ID:          "challenge-id",
		UserID:      42,
		CodeHash:    []byte{1, 2, 3},
		ExpiresAt:   now.Add(10 * time.Minute),
		Attempts:    0,
		MaxAttempts: 5,
		CreatedAt:   now,
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE two_factor_challenges
SET used_at = $2
WHERE user_id = $1 AND used_at IS NULL`)).
		WithArgs(int64(42), now).
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO two_factor_challenges
    (id, user_id, code_hash, expires_at, attempts, max_attempts, used_at, created_at)
VALUES ($1, $2, $3, $4, $5, $6, NULL, $7)`)).
		WithArgs("challenge-id", int64(42), []byte{1, 2, 3}, now.Add(10*time.Minute), 0, 5, now).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	repo := NewPostgresChallengeRepository(db)
	if err := repo.Create(context.Background(), challenge); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("SQL expectations: %v", err)
	}
}

func TestPostgresChallengeRepositoryInvalidate(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New(): %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	now := time.Date(2026, 7, 30, 12, 1, 0, 0, time.UTC)
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE two_factor_challenges
SET used_at = $2
WHERE id = $1 AND used_at IS NULL`)).
		WithArgs("challenge-id", now).
		WillReturnResult(sqlmock.NewResult(0, 1))

	repo := NewPostgresChallengeRepository(db)
	if err := repo.Invalidate(context.Background(), "challenge-id", now); err != nil {
		t.Fatalf("Invalidate() error = %v", err)
	}
}

func TestPostgresChallengeRepositoryConsumesCorrectCode(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New(): %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	now := time.Date(2026, 7, 30, 12, 2, 0, 0, time.UTC)
	hash := []byte{9, 8, 7}
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(challengeSelectForUpdate)).
		WithArgs("challenge-id").
		WillReturnRows(challengeRow("challenge-id", 42, hash, now.Add(time.Minute), 0, 5, nil, now.Add(-time.Minute)))
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE two_factor_challenges
SET used_at = $2
WHERE id = $1 AND used_at IS NULL`)).
		WithArgs("challenge-id", now).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	repo := NewPostgresChallengeRepository(db)
	userID, err := repo.VerifyAndConsume(context.Background(), "challenge-id", hash, now)
	if err != nil {
		t.Fatalf("VerifyAndConsume() error = %v", err)
	}
	if userID != 42 {
		t.Fatalf("VerifyAndConsume() userID = %d, want 42", userID)
	}
}

func TestPostgresChallengeRepositoryIncrementsIncorrectAttempt(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New(): %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	now := time.Date(2026, 7, 30, 12, 3, 0, 0, time.UTC)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(challengeSelectForUpdate)).
		WithArgs("challenge-id").
		WillReturnRows(challengeRow("challenge-id", 42, []byte{1}, now.Add(time.Minute), 2, 5, nil, now))
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE two_factor_challenges
SET attempts = attempts + 1
WHERE id = $1`)).
		WithArgs("challenge-id").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	repo := NewPostgresChallengeRepository(db)
	_, err = repo.VerifyAndConsume(context.Background(), "challenge-id", []byte{2}, now)
	if !errors.Is(err, ErrInvalidCode) {
		t.Fatalf("VerifyAndConsume() error = %v, want ErrInvalidCode", err)
	}
}

func TestPostgresChallengeRepositoryIncorrectFinalAttemptExhaustsLimit(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New(): %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	now := time.Date(2026, 7, 30, 12, 4, 0, 0, time.UTC)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(challengeSelectForUpdate)).
		WithArgs("challenge-id").
		WillReturnRows(challengeRow("challenge-id", 42, []byte{1}, now.Add(time.Minute), 4, 5, nil, now))
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE two_factor_challenges
SET attempts = attempts + 1, used_at = $2
WHERE id = $1`)).
		WithArgs("challenge-id", now).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	repo := NewPostgresChallengeRepository(db)
	_, err = repo.VerifyAndConsume(context.Background(), "challenge-id", []byte{2}, now)
	if !errors.Is(err, ErrTooManyAttempts) {
		t.Fatalf("VerifyAndConsume() error = %v, want ErrTooManyAttempts", err)
	}
}

func TestPostgresChallengeRepositoryExpiresChallenge(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New(): %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	now := time.Date(2026, 7, 30, 12, 5, 0, 0, time.UTC)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(challengeSelectForUpdate)).
		WithArgs("challenge-id").
		WillReturnRows(challengeRow("challenge-id", 42, []byte{1}, now, 0, 5, nil, now.Add(-10*time.Minute)))
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE two_factor_challenges
SET used_at = $2
WHERE id = $1 AND used_at IS NULL`)).
		WithArgs("challenge-id", now).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	repo := NewPostgresChallengeRepository(db)
	_, err = repo.VerifyAndConsume(context.Background(), "challenge-id", []byte{1}, now)
	if !errors.Is(err, ErrChallengeExpired) {
		t.Fatalf("VerifyAndConsume() error = %v, want ErrChallengeExpired", err)
	}
}

func TestPostgresChallengeRepositoryRejectsUsedChallenge(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New(): %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	now := time.Date(2026, 7, 30, 12, 6, 0, 0, time.UTC)
	usedAt := now.Add(-time.Second)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(challengeSelectForUpdate)).
		WithArgs("challenge-id").
		WillReturnRows(challengeRow("challenge-id", 42, []byte{1}, now.Add(time.Minute), 0, 5, &usedAt, now))
	mock.ExpectRollback()

	repo := NewPostgresChallengeRepository(db)
	_, err = repo.VerifyAndConsume(context.Background(), "challenge-id", []byte{1}, now)
	if !errors.Is(err, ErrChallengeUsed) {
		t.Fatalf("VerifyAndConsume() error = %v, want ErrChallengeUsed", err)
	}
}

func TestPostgresChallengeRepositoryRejectsAlreadyExhaustedChallenge(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New(): %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	now := time.Date(2026, 7, 30, 12, 7, 0, 0, time.UTC)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(challengeSelectForUpdate)).
		WithArgs("challenge-id").
		WillReturnRows(challengeRow("challenge-id", 42, []byte{1}, now.Add(time.Minute), 5, 5, nil, now))
	mock.ExpectRollback()

	repo := NewPostgresChallengeRepository(db)
	_, err = repo.VerifyAndConsume(context.Background(), "challenge-id", []byte{1}, now)
	if !errors.Is(err, ErrTooManyAttempts) {
		t.Fatalf("VerifyAndConsume() error = %v, want ErrTooManyAttempts", err)
	}
}

func TestPostgresChallengeRepositoryHidesUnknownChallenge(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New(): %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(challengeSelectForUpdate)).
		WithArgs("missing").
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "code_hash", "expires_at", "attempts", "max_attempts", "used_at", "created_at"}))
	mock.ExpectRollback()

	repo := NewPostgresChallengeRepository(db)
	_, err = repo.VerifyAndConsume(context.Background(), "missing", []byte{1}, time.Now())
	if !errors.Is(err, ErrChallengeNotFound) {
		t.Fatalf("VerifyAndConsume() error = %v, want ErrChallengeNotFound", err)
	}
}

func challengeRow(id string, userID int64, hash []byte, expiresAt time.Time, attempts, maxAttempts int, usedAt *time.Time, createdAt time.Time) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "user_id", "code_hash", "expires_at", "attempts", "max_attempts", "used_at", "created_at"}).
		AddRow(id, userID, hash, expiresAt, attempts, maxAttempts, usedAt, createdAt)
}

func TestPostgresChallengeRepositoryCreatePreservesSerializableIsolation(t *testing.T) {
	t.Parallel()

	state, db := openTransactionCaptureDB(t)
	now := time.Now().UTC()
	repo := NewPostgresChallengeRepository(db)
	err := repo.Create(context.Background(), Challenge{
		ID:          "challenge-id",
		UserID:      42,
		CodeHash:    []byte{1, 2, 3},
		ExpiresAt:   now.Add(10 * time.Minute),
		MaxAttempts: 5,
		CreatedAt:   now,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	options, _ := state.snapshot()
	if options.Isolation != driver.IsolationLevel(sql.LevelSerializable) {
		t.Fatalf("transaction isolation = %d, want Serializable (%d)", options.Isolation, sql.LevelSerializable)
	}
}

func TestPostgresChallengeRepositoryVerifyUsesReadCommittedAndRowLock(t *testing.T) {
	t.Parallel()

	state, db := openTransactionCaptureDB(t)
	now := time.Now().UTC()
	state.challenge = Challenge{
		ID:          "challenge-id",
		UserID:      42,
		CodeHash:    []byte{1, 2, 3},
		ExpiresAt:   now.Add(time.Minute),
		MaxAttempts: 5,
		CreatedAt:   now,
	}
	repo := NewPostgresChallengeRepository(db)
	if _, err := repo.VerifyAndConsume(context.Background(), "challenge-id", []byte{1, 2, 3}, now); err != nil {
		t.Fatalf("VerifyAndConsume() error = %v", err)
	}

	options, query := state.snapshot()
	if options.Isolation != driver.IsolationLevel(sql.LevelReadCommitted) {
		t.Fatalf("transaction isolation = %d, want Read Committed (%d)", options.Isolation, sql.LevelReadCommitted)
	}
	if !strings.Contains(query, "FOR UPDATE") {
		t.Fatalf("verification query does not lock the row: %q", query)
	}
}

var transactionCaptureSequence atomic.Uint64

type transactionCaptureState struct {
	mu        sync.Mutex
	options   driver.TxOptions
	query     string
	challenge Challenge
}

func (s *transactionCaptureState) snapshot() (driver.TxOptions, string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.options, s.query
}

type transactionCaptureDriver struct {
	state *transactionCaptureState
}

func (d transactionCaptureDriver) Open(string) (driver.Conn, error) {
	return &transactionCaptureConn{state: d.state}, nil
}

type transactionCaptureConn struct {
	state *transactionCaptureState
}

func (c *transactionCaptureConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("prepared statements are not supported by the capture driver")
}

func (c *transactionCaptureConn) Close() error {
	return nil
}

func (c *transactionCaptureConn) Begin() (driver.Tx, error) {
	return nil, errors.New("Begin without context is not supported by the capture driver")
}

func (c *transactionCaptureConn) BeginTx(_ context.Context, options driver.TxOptions) (driver.Tx, error) {
	c.state.mu.Lock()
	c.state.options = options
	c.state.mu.Unlock()
	return transactionCaptureTx{}, nil
}

func (c *transactionCaptureConn) ExecContext(_ context.Context, _ string, _ []driver.NamedValue) (driver.Result, error) {
	return driver.RowsAffected(1), nil
}

func (c *transactionCaptureConn) QueryContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Rows, error) {
	c.state.mu.Lock()
	c.state.query = query
	challenge := c.state.challenge
	c.state.mu.Unlock()
	return &challengeCaptureRows{challenge: challenge}, nil
}

type transactionCaptureTx struct{}

func (transactionCaptureTx) Commit() error {
	return nil
}

func (transactionCaptureTx) Rollback() error {
	return nil
}

type challengeCaptureRows struct {
	challenge Challenge
	read      bool
}

func (r *challengeCaptureRows) Columns() []string {
	return []string{"id", "user_id", "code_hash", "expires_at", "attempts", "max_attempts", "used_at", "created_at"}
}

func (r *challengeCaptureRows) Close() error {
	return nil
}

func (r *challengeCaptureRows) Next(values []driver.Value) error {
	if r.read {
		return io.EOF
	}
	r.read = true
	values[0] = r.challenge.ID
	values[1] = r.challenge.UserID
	values[2] = r.challenge.CodeHash
	values[3] = r.challenge.ExpiresAt
	values[4] = int64(r.challenge.Attempts)
	values[5] = int64(r.challenge.MaxAttempts)
	values[6] = nil
	values[7] = r.challenge.CreatedAt
	return nil
}

func openTransactionCaptureDB(t *testing.T) (*transactionCaptureState, *sql.DB) {
	t.Helper()
	state := &transactionCaptureState{}
	driverName := fmt.Sprintf("challenge-transaction-capture-%d", transactionCaptureSequence.Add(1))
	sql.Register(driverName, transactionCaptureDriver{state: state})
	db, err := sql.Open(driverName, "")
	if err != nil {
		t.Fatalf("sql.Open(): %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return state, db
}

var (
	_ driver.ConnBeginTx    = (*transactionCaptureConn)(nil)
	_ driver.ExecerContext  = (*transactionCaptureConn)(nil)
	_ driver.QueryerContext = (*transactionCaptureConn)(nil)
)
