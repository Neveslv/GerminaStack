package auth

import (
	"context"
	"crypto/hmac"
	"errors"
	"regexp"
	"sync"
	"testing"
	"time"

	"germinaStack/database"
)

func TestServiceStartLoginPersistsChallengeAndSendsEmail(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 30, 13, 0, 0, 0, time.UTC)
	passwordHash, err := HashPassword("correct-password")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	credentials := &credentialRepoFake{credential: database.Credential{
		ID:           42,
		Username:     "ana",
		Email:        "ana@example.com",
		PasswordHash: passwordHash,
	}}
	challenges := newChallengeRepoFake()
	mailer := &mailSenderFake{}
	secret := []byte("two-factor-secret")
	service := NewService(credentials, challenges, mailer, secret, fixedClock{now: now})

	challengeID, err := service.StartLogin(context.Background(), "ana", "correct-password")
	if err != nil {
		t.Fatalf("StartLogin() error = %v", err)
	}
	if challengeID == "" {
		t.Fatal("StartLogin() challenge ID is empty")
	}

	challenge, ok := challenges.challenge(challengeID)
	if !ok {
		t.Fatalf("challenge %q was not persisted", challengeID)
	}
	if challenge.UserID != 42 || challenge.Attempts != 0 || challenge.MaxAttempts != 5 {
		t.Fatalf("persisted challenge = %#v", challenge)
	}
	if !challenge.CreatedAt.Equal(now) || !challenge.ExpiresAt.Equal(now.Add(10*time.Minute)) {
		t.Fatalf("challenge times = created %v expires %v", challenge.CreatedAt, challenge.ExpiresAt)
	}
	if len(mailer.messages) != 1 {
		t.Fatalf("sent messages = %d, want 1", len(mailer.messages))
	}
	code := regexp.MustCompile(`[0-9]{6}`).FindString(mailer.messages[0].Body)
	if code == "" {
		t.Fatalf("email body does not contain code: %s", mailer.messages[0].Body)
	}
	if !hmac.Equal(challenge.CodeHash, HashCode(secret, challengeID, code)) {
		t.Fatal("persisted hash is not bound to the emailed code and challenge")
	}
}

func TestServiceStartLoginDoesNotEnumerateCredentials(t *testing.T) {
	t.Parallel()

	passwordHash, err := HashPassword("correct-password")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	now := time.Date(2026, 7, 30, 13, 1, 0, 0, time.UTC)
	missing := NewService(
		&credentialRepoFake{err: database.ErrCredentialNotFound},
		newChallengeRepoFake(),
		&mailSenderFake{},
		[]byte("secret"),
		fixedClock{now: now},
	)
	wrong := NewService(
		&credentialRepoFake{credential: database.Credential{ID: 42, Username: "ana", Email: "ana@example.com", PasswordHash: passwordHash}},
		newChallengeRepoFake(),
		&mailSenderFake{},
		[]byte("secret"),
		fixedClock{now: now},
	)

	_, missingErr := missing.StartLogin(context.Background(), "nobody", "wrong-password")
	_, wrongErr := wrong.StartLogin(context.Background(), "ana", "wrong-password")
	if !errors.Is(missingErr, ErrInvalidCredentials) || !errors.Is(wrongErr, ErrInvalidCredentials) {
		t.Fatalf("errors = (%v, %v), want ErrInvalidCredentials", missingErr, wrongErr)
	}
	if missingErr.Error() != wrongErr.Error() {
		t.Fatalf("error messages differ: %q vs %q", missingErr, wrongErr)
	}
}

func TestServiceStartLoginInvalidatesChallengeWhenSMTPFails(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 30, 13, 2, 0, 0, time.UTC)
	passwordHash, err := HashPassword("correct-password")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	challenges := newChallengeRepoFake()
	service := NewService(
		&credentialRepoFake{credential: database.Credential{ID: 42, Username: "ana", Email: "ana@example.com", PasswordHash: passwordHash}},
		challenges,
		&mailSenderFake{err: errors.New("SMTP unavailable with private detail")},
		[]byte("secret"),
		fixedClock{now: now},
	)

	challengeID, err := service.StartLogin(context.Background(), "ana", "correct-password")
	if challengeID != "" {
		t.Fatalf("StartLogin() challenge ID = %q, want empty", challengeID)
	}
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("StartLogin() error = %v, want ErrUnavailable", err)
	}
	if err.Error() != ErrUnavailable.Error() {
		t.Fatalf("public error leaked detail: %q", err.Error())
	}
	if challenges.invalidations != 1 {
		t.Fatalf("invalidations = %d, want 1", challenges.invalidations)
	}
}

func TestServiceCompleteLoginHandlesChallengeStates(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		repoErr error
		wantErr error
	}{
		{name: "incorrect", repoErr: database.ErrInvalidCode, wantErr: ErrInvalidCode},
		{name: "expired", repoErr: database.ErrChallengeExpired, wantErr: ErrChallengeExpired},
		{name: "used", repoErr: database.ErrChallengeUsed, wantErr: ErrChallengeUsed},
		{name: "missing", repoErr: database.ErrChallengeNotFound, wantErr: ErrInvalidCode},
		{name: "limit", repoErr: database.ErrTooManyAttempts, wantErr: ErrTooManyAttempts},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			challenges := newChallengeRepoFake()
			challenges.verifyErr = tt.repoErr
			service := NewService(
				&credentialRepoFake{},
				challenges,
				&mailSenderFake{},
				[]byte("secret"),
				fixedClock{now: time.Date(2026, 7, 30, 13, 3, 0, 0, time.UTC)},
			)

			_, err := service.CompleteLogin(context.Background(), "challenge-id", "123456")
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("CompleteLogin() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestServiceCompleteLoginRejectsMalformedCodeBeforeRepository(t *testing.T) {
	t.Parallel()

	challenges := newChallengeRepoFake()
	service := NewService(
		&credentialRepoFake{},
		challenges,
		&mailSenderFake{},
		[]byte("secret"),
		fixedClock{now: time.Now()},
	)
	for _, code := range []string{"12345", "1234567", "12a456", "１２３４５６"} {
		if _, err := service.CompleteLogin(context.Background(), "challenge-id", code); !errors.Is(err, ErrMalformedCode) {
			t.Fatalf("CompleteLogin(%q) error = %v, want ErrMalformedCode", code, err)
		}
	}
	if challenges.verifyCalls != 0 {
		t.Fatalf("repository verify calls = %d, want 0", challenges.verifyCalls)
	}
}

func TestServiceCompleteLoginReturnsUserForCorrectCode(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 30, 13, 4, 0, 0, time.UTC)
	secret := []byte("secret")
	challenges := newChallengeRepoFake()
	challenges.items["challenge-id"] = database.Challenge{
		ID:          "challenge-id",
		UserID:      42,
		CodeHash:    HashCode(secret, "challenge-id", "123456"),
		ExpiresAt:   now.Add(time.Minute),
		MaxAttempts: 5,
		CreatedAt:   now,
	}
	service := NewService(&credentialRepoFake{}, challenges, &mailSenderFake{}, secret, fixedClock{now: now})

	userID, err := service.CompleteLogin(context.Background(), "challenge-id", "123456")
	if err != nil {
		t.Fatalf("CompleteLogin() error = %v", err)
	}
	if userID != 42 {
		t.Fatalf("CompleteLogin() userID = %d, want 42", userID)
	}
}

func TestServiceCompleteLoginConcurrentConsumesOnlyOnce(t *testing.T) {
	now := time.Date(2026, 7, 30, 13, 5, 0, 0, time.UTC)
	secret := []byte("secret")
	challenges := newChallengeRepoFake()
	challenges.items["challenge-id"] = database.Challenge{
		ID:          "challenge-id",
		UserID:      42,
		CodeHash:    HashCode(secret, "challenge-id", "123456"),
		ExpiresAt:   now.Add(time.Minute),
		MaxAttempts: 5,
		CreatedAt:   now,
	}
	service := NewService(&credentialRepoFake{}, challenges, &mailSenderFake{}, secret, fixedClock{now: now})

	start := make(chan struct{})
	results := make(chan error, 2)
	for range 2 {
		go func() {
			<-start
			_, err := service.CompleteLogin(context.Background(), "challenge-id", "123456")
			results <- err
		}()
	}
	close(start)

	successes := 0
	used := 0
	for range 2 {
		err := <-results
		switch {
		case err == nil:
			successes++
		case errors.Is(err, ErrChallengeUsed):
			used++
		default:
			t.Fatalf("unexpected concurrent result: %v", err)
		}
	}
	if successes != 1 || used != 1 {
		t.Fatalf("concurrent results: successes=%d used=%d, want 1 and 1", successes, used)
	}
}

type fixedClock struct {
	now time.Time
}

func (c fixedClock) Now() time.Time {
	return c.now
}

type credentialRepoFake struct {
	credential database.Credential
	err        error
}

func (f *credentialRepoFake) FindByUsername(context.Context, string) (database.Credential, error) {
	return f.credential, f.err
}

type mailSenderFake struct {
	messages []Message
	err      error
}

func (f *mailSenderFake) Send(_ context.Context, message Message) error {
	f.messages = append(f.messages, message)
	return f.err
}

type challengeRepoFake struct {
	mu            sync.Mutex
	items         map[string]database.Challenge
	verifyErr     error
	verifyCalls   int
	invalidations int
}

func newChallengeRepoFake() *challengeRepoFake {
	return &challengeRepoFake{items: make(map[string]database.Challenge)}
}

func (f *challengeRepoFake) Create(_ context.Context, challenge database.Challenge) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	for id, previous := range f.items {
		if previous.UserID == challenge.UserID && previous.UsedAt == nil {
			usedAt := challenge.CreatedAt
			previous.UsedAt = &usedAt
			f.items[id] = previous
		}
	}
	f.items[challenge.ID] = challenge
	return nil
}

func (f *challengeRepoFake) VerifyAndConsume(_ context.Context, challengeID string, presentedHash []byte, now time.Time) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.verifyCalls++
	if f.verifyErr != nil {
		return 0, f.verifyErr
	}
	challenge, ok := f.items[challengeID]
	if !ok {
		return 0, database.ErrChallengeNotFound
	}
	if challenge.UsedAt != nil {
		return 0, database.ErrChallengeUsed
	}
	if !hmac.Equal(challenge.CodeHash, presentedHash) {
		return 0, database.ErrInvalidCode
	}
	usedAt := now
	challenge.UsedAt = &usedAt
	f.items[challengeID] = challenge
	return challenge.UserID, nil
}

func (f *challengeRepoFake) Invalidate(_ context.Context, challengeID string, at time.Time) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.invalidations++
	challenge := f.items[challengeID]
	challenge.UsedAt = &at
	f.items[challengeID] = challenge
	return nil
}

func (f *challengeRepoFake) challenge(id string) (database.Challenge, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	challenge, ok := f.items[id]
	return challenge, ok
}
