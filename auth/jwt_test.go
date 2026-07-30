package auth

import (
	"testing"
	"time"
)

func TestGenerateTokenWithTimesIncludesRequiredClaims(t *testing.T) {
	t.Parallel()

	issuedAt := time.Date(2026, 7, 30, 14, 0, 0, 0, time.UTC)
	expiresAt := issuedAt.Add(24 * time.Hour)
	token, err := GenerateTokenWithTimes("42", false, "jwt-test-secret", issuedAt, expiresAt)
	if err != nil {
		t.Fatalf("GenerateTokenWithTimes() error = %v", err)
	}
	claims, err := ParseToken(token, "jwt-test-secret")
	if err != nil {
		t.Fatalf("ParseToken() error = %v", err)
	}
	if claims.Subject != "42" || claims.IsAdmin {
		t.Fatalf("claims = %#v", claims)
	}
	if claims.IssuedAt == nil || !claims.IssuedAt.Time.Equal(issuedAt) {
		t.Fatalf("IssuedAt = %v, want %v", claims.IssuedAt, issuedAt)
	}
	if claims.ExpiresAt == nil || !claims.ExpiresAt.Time.Equal(expiresAt) {
		t.Fatalf("ExpiresAt = %v, want %v", claims.ExpiresAt, expiresAt)
	}
}

func TestGenerateTokenWithTimesRejectsMissingSecretAndInvalidWindow(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 30, 14, 0, 0, 0, time.UTC)
	if _, err := GenerateTokenWithTimes("42", false, "", now, now.Add(time.Hour)); err == nil {
		t.Fatal("GenerateTokenWithTimes() error = nil, want missing secret")
	}
	if _, err := GenerateTokenWithTimes("42", false, "secret", now, now); err == nil {
		t.Fatal("GenerateTokenWithTimes() error = nil, want invalid time window")
	}
}

func TestGenerateTokenCompatibilityWrapperIncludesTimes(t *testing.T) {
	t.Parallel()

	token, err := GenerateToken("42", false, "jwt-test-secret")
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}
	claims, err := ParseToken(token, "jwt-test-secret")
	if err != nil {
		t.Fatalf("ParseToken() error = %v", err)
	}
	if claims.IssuedAt == nil || claims.ExpiresAt == nil {
		t.Fatalf("compatibility token lacks temporal claims: %#v", claims)
	}
	if got := claims.ExpiresAt.Time.Sub(claims.IssuedAt.Time); got != 24*time.Hour {
		t.Fatalf("token lifetime = %v, want 24h", got)
	}
}
