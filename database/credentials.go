package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	domainauth "germinaStack/domain/auth"
	"germinaStack/domain/users"

	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrCredentialNotFound = domainauth.ErrCredentialNotFound
	ErrCredentialConflict = users.ErrCredentialConflict
	ErrYearNotFound       = users.ErrYearNotFound
)

type Credential = domainauth.Credential
type UserRegistration = users.Registration
type User = users.Credential
type PostgresCredentialRepository struct {
	db *sql.DB
}

func NewPostgresCredentialRepository(db *sql.DB) *PostgresCredentialRepository {
	return &PostgresCredentialRepository{db: db}
}

func (r *PostgresCredentialRepository) FindByUsername(ctx context.Context, username string) (Credential, error) {
	const query = `SELECT id, username, email, password, is_admin, is_banned
FROM users
WHERE username = $1`
	return r.findCredential(ctx, query, username)
}

func (r *PostgresCredentialRepository) FindByEmail(ctx context.Context, email string) (Credential, error) {
	const query = `SELECT id, username, email, password, is_admin, is_banned
FROM users
WHERE email = $1`
	return r.findCredential(ctx, query, email)
}

func (r *PostgresCredentialRepository) FindByID(ctx context.Context, userID int64) (Credential, error) {
	const query = `SELECT id, username, email, password, is_admin, is_banned
FROM users
WHERE id = $1`
	return r.findCredential(ctx, query, userID)
}

func (r *PostgresCredentialRepository) findCredential(ctx context.Context, query string, argument any) (Credential, error) {
	var credential Credential
	err := r.db.QueryRowContext(ctx, query, argument).Scan(
		&credential.ID,
		&credential.Username,
		&credential.Email,
		&credential.PasswordHash,
		&credential.IsAdmin,
		&credential.IsBanned,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Credential{}, ErrCredentialNotFound
	}
	if err != nil {
		return Credential{}, fmt.Errorf("find credential: %w", err)
	}
	return credential, nil
}

func (r *PostgresCredentialRepository) CreateUser(ctx context.Context, registration UserRegistration) (User, error) {
	const query = `INSERT INTO users (id_year, name, username, email, password, is_admin)
VALUES ($1, $2, $3, $4, $5, FALSE)
RETURNING id, id_year, name, username, email, is_admin`

	var user User
	err := r.db.QueryRowContext(
		ctx,
		query,
		registration.YearID,
		registration.Name,
		registration.Username,
		registration.Email,
		registration.PasswordHash,
	).Scan(
		&user.ID,
		&user.YearID,
		&user.Name,
		&user.Username,
		&user.Email,
		&user.IsAdmin,
	)
	if err == nil {
		return user, nil
	}

	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) {
		switch postgresError.Code {
		case "23505":
			return User{}, ErrCredentialConflict
		case "23503":
			return User{}, ErrYearNotFound
		}
	}
	return User{}, fmt.Errorf("create user: %w", err)
}
