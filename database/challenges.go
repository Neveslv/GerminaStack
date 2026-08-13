package database

import (
	"context"
	"crypto/hmac"
	"database/sql"
	"errors"
	"fmt"
	"time"

	domainauth "germinaStack/domain/auth"
)

var (
	ErrChallengeNotFound = domainauth.ErrChallengeNotFound
	ErrChallengeUsed     = domainauth.ErrChallengeUsed
	ErrChallengeExpired  = domainauth.ErrChallengeExpired
	ErrInvalidCode       = domainauth.ErrInvalidCode
	ErrTooManyAttempts   = domainauth.ErrTooManyAttempts
)

type Challenge = domainauth.Challenge

type PostgresChallengeRepository struct {
	db *sql.DB
}

func NewPostgresChallengeRepository(db *sql.DB) *PostgresChallengeRepository {
	return &PostgresChallengeRepository{db: db}
}

func (r *PostgresChallengeRepository) Create(ctx context.Context, challenge Challenge) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("begin challenge creation: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	const invalidatePrevious = `UPDATE two_factor_challenges
SET used_at = $2
WHERE user_id = $1 AND used_at IS NULL`
	if _, err := tx.ExecContext(ctx, invalidatePrevious, challenge.UserID, challenge.CreatedAt); err != nil {
		return fmt.Errorf("invalidate previous challenges: %w", err)
	}

	const insert = `INSERT INTO two_factor_challenges
    (id, user_id, code_hash, expires_at, attempts, max_attempts, used_at, created_at)
VALUES ($1, $2, $3, $4, $5, $6, NULL, $7)`
	if _, err := tx.ExecContext(
		ctx,
		insert,
		challenge.ID,
		challenge.UserID,
		challenge.CodeHash,
		challenge.ExpiresAt,
		challenge.Attempts,
		challenge.MaxAttempts,
		challenge.CreatedAt,
	); err != nil {
		return fmt.Errorf("insert challenge: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit challenge creation: %w", err)
	}
	return nil
}

func (r *PostgresChallengeRepository) Invalidate(ctx context.Context, challengeID string, at time.Time) error {
	const query = `UPDATE two_factor_challenges
SET used_at = $2
WHERE id = $1 AND used_at IS NULL`
	if _, err := r.db.ExecContext(ctx, query, challengeID, at); err != nil {
		return fmt.Errorf("invalidate challenge: %w", err)
	}
	return nil
}

func (r *PostgresChallengeRepository) VerifyAndConsume(ctx context.Context, challengeID string, presentedHash []byte, now time.Time) (int64, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return 0, fmt.Errorf("begin challenge verification: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	const selectForUpdate = `SELECT id, user_id, code_hash, expires_at, attempts, max_attempts, used_at, created_at
FROM two_factor_challenges
WHERE id = $1
FOR UPDATE`

	var challenge Challenge
	err = tx.QueryRowContext(ctx, selectForUpdate, challengeID).Scan(
		&challenge.ID,
		&challenge.UserID,
		&challenge.CodeHash,
		&challenge.ExpiresAt,
		&challenge.Attempts,
		&challenge.MaxAttempts,
		&challenge.UsedAt,
		&challenge.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrChallengeNotFound
	}
	if err != nil {
		return 0, fmt.Errorf("load challenge: %w", err)
	}

	if challenge.UsedAt != nil {
		return 0, ErrChallengeUsed
	}
	if challenge.Attempts >= challenge.MaxAttempts {
		return 0, ErrTooManyAttempts
	}
	if !now.Before(challenge.ExpiresAt) {
		if err := markChallengeUsed(ctx, tx, challengeID, now); err != nil {
			return 0, err
		}
		return 0, commitDomainError(tx, ErrChallengeExpired)
	}
	if !hmac.Equal(challenge.CodeHash, presentedHash) {
		if challenge.Attempts+1 >= challenge.MaxAttempts {
			const exhaust = `UPDATE two_factor_challenges
SET attempts = attempts + 1, used_at = $2
WHERE id = $1`
			if _, err := tx.ExecContext(ctx, exhaust, challengeID, now); err != nil {
				return 0, fmt.Errorf("exhaust challenge attempts: %w", err)
			}
			return 0, commitDomainError(tx, ErrTooManyAttempts)
		}

		const increment = `UPDATE two_factor_challenges
SET attempts = attempts + 1
WHERE id = $1`
		if _, err := tx.ExecContext(ctx, increment, challengeID); err != nil {
			return 0, fmt.Errorf("increment challenge attempts: %w", err)
		}
		return 0, commitDomainError(tx, ErrInvalidCode)
	}

	if err := markChallengeUsed(ctx, tx, challengeID, now); err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit challenge consumption: %w", err)
	}
	return challenge.UserID, nil
}

func markChallengeUsed(ctx context.Context, tx *sql.Tx, challengeID string, at time.Time) error {
	const query = `UPDATE two_factor_challenges
SET used_at = $2
WHERE id = $1 AND used_at IS NULL`
	result, err := tx.ExecContext(ctx, query, challengeID, at)
	if err != nil {
		return fmt.Errorf("mark challenge used: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read challenge update result: %w", err)
	}
	if affected != 1 {
		return ErrChallengeUsed
	}
	return nil
}

func commitDomainError(tx *sql.Tx, domainErr error) error {
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit challenge state: %w", err)
	}
	return domainErr
}
