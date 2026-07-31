package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

var ErrCredentialNotFound = errors.New("credential not found")

type Credential struct {
	ID           int64
	Username     string
	Email        string
	PasswordHash string
}

type PostgresCredentialRepository struct {
	db *sql.DB
}

func NewPostgresCredentialRepository(db *sql.DB) *PostgresCredentialRepository {
	return &PostgresCredentialRepository{db: db}
}

func (r *PostgresCredentialRepository) FindByUsername(ctx context.Context, username string) (Credential, error) {
	const query = `SELECT id, username, email, password
FROM users
WHERE username = $1`

	var credential Credential
	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&credential.ID,
		&credential.Username,
		&credential.Email,
		&credential.PasswordHash,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Credential{}, ErrCredentialNotFound
	}
	if err != nil {
		return Credential{}, fmt.Errorf("find credential: %w", err)
	}
	return credential, nil
}
