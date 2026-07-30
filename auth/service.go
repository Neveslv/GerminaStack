package auth

import (
	"context"
	"errors"
	"time"

	"germinaStack/database"
)

const (
	ChallengeTTL                 = 10 * time.Minute
	ChallengeMaxAttempts         = 5
	ChallengeInvalidationTimeout = 3 * time.Second
)

var (
	ErrInvalidCredentials = errors.New("credenciais inválidas")
	ErrInvalidCode        = errors.New("código inválido")
	ErrMalformedCode      = errors.New("código malformado")
	ErrChallengeExpired   = errors.New("desafio expirado")
	ErrChallengeUsed      = errors.New("desafio já utilizado")
	ErrTooManyAttempts    = errors.New("limite de tentativas excedido")
	ErrUnavailable        = errors.New("serviço temporariamente indisponível")
)

type CredentialRepository interface {
	FindByUsername(context.Context, string) (database.Credential, error)
}

type ChallengeRepository interface {
	Create(context.Context, database.Challenge) error
	VerifyAndConsume(context.Context, string, []byte, time.Time) (int64, error)
	Invalidate(context.Context, string, time.Time) error
}

type MailSender interface {
	Send(context.Context, Message) error
}

type Clock interface {
	Now() time.Time
}

type Service struct {
	credentials CredentialRepository
	challenges  ChallengeRepository
	mailer      MailSender
	secret      []byte
	clock       Clock
}

func NewService(credentials CredentialRepository, challenges ChallengeRepository, mailer MailSender, secret []byte, clock Clock) *Service {
	return &Service{
		credentials: credentials,
		challenges:  challenges,
		mailer:      mailer,
		secret:      append([]byte(nil), secret...),
		clock:       clock,
	}
}

func (s *Service) StartLogin(ctx context.Context, username, password string) (string, error) {
	if username == "" || password == "" || len(s.secret) == 0 {
		return "", ErrInvalidCredentials
	}

	credential, err := s.credentials.FindByUsername(ctx, username)
	if errors.Is(err, database.ErrCredentialNotFound) {
		_ = CheckPassword(dummyPasswordHash(), password)
		return "", ErrInvalidCredentials
	}
	if err != nil {
		return "", ErrUnavailable
	}
	if err := CheckPassword(credential.PasswordHash, password); err != nil {
		return "", ErrInvalidCredentials
	}

	challengeID, err := GenerateChallengeID()
	if err != nil {
		return "", ErrUnavailable
	}
	code, err := GenerateCode()
	if err != nil {
		return "", ErrUnavailable
	}
	now := s.clock.Now().UTC()
	challenge := database.Challenge{
		ID:          challengeID,
		UserID:      credential.ID,
		CodeHash:    HashCode(s.secret, challengeID, code),
		ExpiresAt:   now.Add(ChallengeTTL),
		Attempts:    0,
		MaxAttempts: ChallengeMaxAttempts,
		CreatedAt:   now,
	}
	if err := s.challenges.Create(ctx, challenge); err != nil {
		return "", ErrUnavailable
	}

	message, err := AuthenticationMessage(credential.Username, credential.Email, code)
	if err != nil {
		s.invalidateAfterSendFailure(ctx, challengeID, now)
		return "", ErrUnavailable
	}
	if err := s.mailer.Send(ctx, message); err != nil {
		s.invalidateAfterSendFailure(ctx, challengeID, s.clock.Now().UTC())
		return "", ErrUnavailable
	}
	return challengeID, nil
}

func (s *Service) CompleteLogin(ctx context.Context, challengeID, code string) (int64, error) {
	if !sixDigitCode.MatchString(code) {
		return 0, ErrMalformedCode
	}
	if challengeID == "" || len(challengeID) > 128 || len(s.secret) == 0 {
		return 0, ErrInvalidCode
	}

	userID, err := s.challenges.VerifyAndConsume(
		ctx,
		challengeID,
		HashCode(s.secret, challengeID, code),
		s.clock.Now().UTC(),
	)
	switch {
	case err == nil:
		return userID, nil
	case errors.Is(err, database.ErrInvalidCode), errors.Is(err, database.ErrChallengeNotFound):
		return 0, ErrInvalidCode
	case errors.Is(err, database.ErrChallengeExpired):
		return 0, ErrChallengeExpired
	case errors.Is(err, database.ErrChallengeUsed):
		return 0, ErrChallengeUsed
	case errors.Is(err, database.ErrTooManyAttempts):
		return 0, ErrTooManyAttempts
	default:
		return 0, ErrUnavailable
	}
}

func (s *Service) invalidateAfterSendFailure(ctx context.Context, challengeID string, at time.Time) {
	invalidationContext, cancel := context.WithTimeout(context.WithoutCancel(ctx), ChallengeInvalidationTimeout)
	defer cancel()
	_ = s.challenges.Invalidate(invalidationContext, challengeID, at)
}

var sentinelHash = func() string {
	hash, err := HashPassword("invalid-credential-sentinel")
	if err != nil {
		panic("bcrypt sentinel initialization failed")
	}
	return hash
}()

func dummyPasswordHash() string {
	return sentinelHash
}
