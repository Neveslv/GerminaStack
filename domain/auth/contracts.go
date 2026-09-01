package auth

import (
	"context"
	"errors"
	"time"
)

var (
	ErrCredentialNotFound = errors.New("credential not found")
	ErrChallengeNotFound  = errors.New("challenge not found")
	ErrChallengeUsed      = errors.New("challenge already used")
	ErrChallengeExpired   = errors.New("challenge expired")
	ErrInvalidCode        = errors.New("invalid authentication code")
	ErrTooManyAttempts    = errors.New("too many authentication attempts")
)

type Credential struct {
	ID           int64
	Username     string
	Email        string
	PasswordHash string
	IsAdmin      bool
	IsBanned     bool
}

type Challenge struct {
	ID          string
	UserID      int64
	CodeHash    []byte
	ExpiresAt   time.Time
	Attempts    int
	MaxAttempts int
	UsedAt      *time.Time
	CreatedAt   time.Time
}

type CredentialRepository interface {
	FindByEmail(context.Context, string) (Credential, error)
	FindByID(context.Context, int64) (Credential, error)
}

type ChallengeRepository interface {
	Create(context.Context, Challenge) error
	VerifyAndConsume(context.Context, string, []byte, time.Time) (int64, error)
	Invalidate(context.Context, string, time.Time) error
}
