package users

import (
	"context"
	"errors"
)

var (
	ErrCredentialConflict = errors.New("credential already exists")
	ErrYearNotFound       = errors.New("year not found")
)

type Credential struct {
	ID       int64
	YearID   int64
	Name     string
	Username string
	Email    string
	IsAdmin  bool
}

type Registration struct {
	YearID       int64
	Name         string
	Username     string
	Email        string
	PasswordHash string
}

type Repository interface {
	CreateUser(context.Context, Registration) (Credential, error)
}
