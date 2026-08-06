package auth

import (
	"encoding/base64"
	"regexp"
	"testing"
)

func TestGenerateChallengeIDUsesThirtyTwoRandomBytes(t *testing.T) {
	t.Parallel()

	first, err := GenerateChallengeID()
	if err != nil {
		t.Fatalf("GenerateChallengeID() error = %v", err)
	}
	second, err := GenerateChallengeID()
	if err != nil {
		t.Fatalf("GenerateChallengeID() second error = %v", err)
	}
	if first == second {
		t.Fatal("GenerateChallengeID() returned duplicate IDs")
	}
	decoded, err := base64.RawURLEncoding.DecodeString(first)
	if err != nil {
		t.Fatalf("DecodeString() error = %v", err)
	}
	if len(decoded) != 32 {
		t.Fatalf("decoded ID length = %d, want 32", len(decoded))
	}
}

func TestGenerateCodeAlwaysReturnsSixDecimalDigits(t *testing.T) {
	t.Parallel()

	pattern := regexp.MustCompile(`^[0-9]{6}$`)
	for range 100 {
		code, err := GenerateCode()
		if err != nil {
			t.Fatalf("GenerateCode() error = %v", err)
		}
		if !pattern.MatchString(code) {
			t.Fatalf("GenerateCode() = %q, want six decimal digits", code)
		}
	}
}

func TestHashCodeBindsChallengeAndCode(t *testing.T) {
	t.Parallel()

	secret := []byte("two-factor-test-secret")
	base := HashCode(secret, "challenge-a", "123456")
	if string(base) == "123456" {
		t.Fatal("HashCode() returned plaintext code")
	}
	if got := HashCode(secret, "challenge-a", "123456"); string(got) != string(base) {
		t.Fatal("HashCode() is not deterministic")
	}
	if got := HashCode(secret, "challenge-b", "123456"); string(got) == string(base) {
		t.Fatal("HashCode() did not bind challenge ID")
	}
	if got := HashCode(secret, "challenge-a", "654321"); string(got) == string(base) {
		t.Fatal("HashCode() did not bind code")
	}
}
