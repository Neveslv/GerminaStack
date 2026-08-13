package auth

import (
	"context"
	"errors"
	"log"
	"os"
	"strings"
	"time"

	domainauth "germinaStack/domain/auth"
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

type Principal struct {
	ID      int64
	IsAdmin bool
}

type MailSender interface {
	Send(context.Context, Message) error
}

type Clock interface {
	Now() time.Time
}

type Service struct {
	credentials domainauth.CredentialRepository
	challenges  domainauth.ChallengeRepository
	mailer      MailSender
	secret      []byte
	clock       Clock
}

func NewService(credentials domainauth.CredentialRepository, challenges domainauth.ChallengeRepository, mailer MailSender, secret []byte, clock Clock) *Service {
	return &Service{
		credentials: credentials,
		challenges:  challenges,
		mailer:      mailer,
		secret:      append([]byte(nil), secret...),
		clock:       clock,
	}
}

func (s *Service) StartLogin(ctx context.Context, email, password string) (string, error) {
	if email == "" || password == "" || len(s.secret) == 0 {
		return "", ErrInvalidCredentials
	}

	credential, err := s.credentials.FindByEmail(ctx, email)
	if errors.Is(err, domainauth.ErrCredentialNotFound) {
		_ = CheckPassword(dummyPasswordHash(), password)
		return "", ErrInvalidCredentials
	}
	if err != nil {
		return "", ErrUnavailable
	}
	if credential.IsBanned || CheckPassword(credential.PasswordHash, password) != nil {
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
	challenge := domainauth.Challenge{
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
	if recipient := strings.TrimSpace(os.Getenv("SMTP_DEBUG_RECIPIENT")); recipient != "" {
		message, err = AuthenticationMessage(credential.Username, recipient, code)
		log.Printf("SMTP_DEBUG_RECIPIENT active: sending authentication email to %q", recipient)
	}
	if err != nil {
		s.invalidateAfterSendFailure(ctx, challengeID, now)
		return "", ErrUnavailable
	}
	if err := s.mailer.Send(ctx, message); err != nil {
		log.Printf("authentication email delivery failed for domain %q: %v", emailDomain(credential.Email), err)
		s.invalidateAfterSendFailure(ctx, challengeID, s.clock.Now().UTC())
		return "", ErrUnavailable
	}
	return challengeID, nil
}

func emailDomain(email string) string {
	if _, domain, ok := strings.Cut(email, "@"); ok {
		return domain
	}
	return "invalid"
}

func (s *Service) CompleteLogin(ctx context.Context, challengeID, code string) (Principal, error) {
	if !sixDigitCode.MatchString(code) {
		return Principal{}, ErrMalformedCode
	}
	if challengeID == "" || len(challengeID) > 128 || len(s.secret) == 0 {
		return Principal{}, ErrInvalidCode
	}

	userID, err := s.challenges.VerifyAndConsume(
		ctx,
		challengeID,
		HashCode(s.secret, challengeID, code),
		s.clock.Now().UTC(),
	)
	switch {
	case err == nil:
		credential, findErr := s.credentials.FindByID(ctx, userID)
		if findErr != nil {
			return Principal{}, ErrUnavailable
		}
		if credential.IsBanned {
			return Principal{}, ErrInvalidCredentials
		}
		return Principal{ID: credential.ID, IsAdmin: credential.IsAdmin}, nil
	case errors.Is(err, domainauth.ErrInvalidCode), errors.Is(err, domainauth.ErrChallengeNotFound):
		return Principal{}, ErrInvalidCode
	case errors.Is(err, domainauth.ErrChallengeExpired):
		return Principal{}, ErrChallengeExpired
	case errors.Is(err, domainauth.ErrChallengeUsed):
		return Principal{}, ErrChallengeUsed
	case errors.Is(err, domainauth.ErrTooManyAttempts):
		return Principal{}, ErrTooManyAttempts
	default:
		return Principal{}, ErrUnavailable
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
