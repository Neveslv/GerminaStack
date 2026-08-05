package database

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestUserRegistrationYearIDIsRequired(t *testing.T) {
	t.Parallel()

	registrationYearID, found := reflect.TypeOf(UserRegistration{}).FieldByName("YearID")
	if !found {
		t.Fatal("UserRegistration.YearID field not found")
	}
	if got, want := registrationYearID.Type, reflect.TypeOf(int64(0)); got != want {
		t.Fatalf("UserRegistration.YearID type = %s, want %s", got, want)
	}
}

const credentialQuery = `SELECT id, username, email, password, is_admin
FROM users
WHERE username = $1`

func TestPostgresCredentialRepositoryFindByUsername(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New(): %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	mock.ExpectQuery(regexp.QuoteMeta(credentialQuery)).
		WithArgs("ana").
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "email", "password", "is_admin"}).
			AddRow(int64(42), "ana", "ana@example.com", "$2a$12$hash", true))

	repo := NewPostgresCredentialRepository(db)
	got, err := repo.FindByUsername(context.Background(), "ana")
	if err != nil {
		t.Fatalf("FindByUsername() error = %v", err)
	}

	want := (Credential{ID: 42, Username: "ana", Email: "ana@example.com", PasswordHash: "$2a$12$hash", IsAdmin: true})
	if got != want {
		t.Fatalf("FindByUsername() = %#v, want %#v", got, want)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("SQL expectations: %v", err)
	}
}

func TestPostgresCredentialRepositoryHidesMissingUsername(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New(): %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	mock.ExpectQuery(regexp.QuoteMeta(credentialQuery)).
		WithArgs("nobody").
		WillReturnError(sql.ErrNoRows)

	repo := NewPostgresCredentialRepository(db)
	_, err = repo.FindByUsername(context.Background(), "nobody")
	if !errors.Is(err, ErrCredentialNotFound) {
		t.Fatalf("FindByUsername() error = %v, want ErrCredentialNotFound", err)
	}
	if err.Error() != ErrCredentialNotFound.Error() {
		t.Fatalf("error text = %q, want %q", err.Error(), ErrCredentialNotFound.Error())
	}
}

func TestPostgresCredentialRepositoryReturnsDatabaseFailure(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New(): %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	mock.ExpectQuery(regexp.QuoteMeta(credentialQuery)).
		WithArgs("ana").
		WillReturnError(context.DeadlineExceeded)

	repo := NewPostgresCredentialRepository(db)
	_, err = repo.FindByUsername(context.Background(), "ana")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("FindByUsername() error = %v, want context deadline exceeded", err)
	}
}

func TestPostgresCredentialRepositoryFindByEmailIncludesAdmin(t *testing.T) {
	t.Parallel()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New(): %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	const query = `SELECT id, username, email, password, is_admin
FROM users
WHERE email = $1`
	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs("ana.silva@institutojef.org.br").
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "email", "password", "is_admin"}).
			AddRow(int64(42), "ana.silva", "ana.silva@institutojef.org.br", "$2a$12$hash", true))

	got, err := NewPostgresCredentialRepository(db).FindByEmail(context.Background(), "ana.silva@institutojef.org.br")
	if err != nil {
		t.Fatalf("FindByEmail() error = %v", err)
	}
	want := Credential{ID: 42, Username: "ana.silva", Email: "ana.silva@institutojef.org.br", PasswordHash: "$2a$12$hash", IsAdmin: true}
	if got != want {
		t.Fatalf("FindByEmail() = %#v, want %#v", got, want)
	}
}

func TestPostgresCredentialRepositoryFindByIDReturnsCurrentPrincipalData(t *testing.T) {
	t.Parallel()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New(): %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	const query = `SELECT id, username, email, password, is_admin
FROM users
WHERE id = $1`
	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "email", "password", "is_admin"}).
			AddRow(int64(42), "ana.silva", "ana.silva@institutojef.org.br", "$2a$12$hash", true))

	got, err := NewPostgresCredentialRepository(db).FindByID(context.Background(), 42)
	if err != nil || got.ID != 42 || !got.IsAdmin {
		t.Fatalf("FindByID() = (%#v, %v)", got, err)
	}
}

func TestPostgresCredentialRepositoryCreateUserForcesNonAdmin(t *testing.T) {
	t.Parallel()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New(): %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	const query = `INSERT INTO users (id_year, name, username, email, password, is_admin)
VALUES ($1, $2, $3, $4, $5, FALSE)
RETURNING id, id_year, name, username, email, is_admin`
	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(int64(7), "Ana Silva", "ana.silva", "ana.silva@institutojef.org.br", "$2a$12$hash").
		WillReturnRows(sqlmock.NewRows([]string{"id", "id_year", "name", "username", "email", "is_admin"}).
			AddRow(int64(42), int64(7), "Ana Silva", "ana.silva", "ana.silva@institutojef.org.br", false))

	got, err := NewPostgresCredentialRepository(db).CreateUser(context.Background(), UserRegistration{
		YearID: 7, Name: "Ana Silva", Username: "ana.silva", Email: "ana.silva@institutojef.org.br", PasswordHash: "$2a$12$hash",
	})
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}
	want := User{ID: 42, YearID: 7, Name: "Ana Silva", Username: "ana.silva", Email: "ana.silva@institutojef.org.br", IsAdmin: false}
	if got != want {
		t.Fatalf("CreateUser() = %#v, want %#v", got, want)
	}
}

func TestPostgresCredentialRepositoryCreateUserMapsSafeDomainErrors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		code string
		want error
	}{
		{name: "duplicate", code: "23505", want: ErrCredentialConflict},
		{name: "unknown year", code: "23503", want: ErrYearNotFound},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("sqlmock.New(): %v", err)
			}
			t.Cleanup(func() { _ = db.Close() })
			mock.ExpectQuery("INSERT INTO users").WillReturnError(&pgconn.PgError{Code: tt.code, Message: "private detail"})

			_, err = NewPostgresCredentialRepository(db).CreateUser(context.Background(), UserRegistration{YearID: 7})
			if !errors.Is(err, tt.want) || err.Error() != tt.want.Error() {
				t.Fatalf("CreateUser() error = %v, want safe %v", err, tt.want)
			}
		})
	}
}
