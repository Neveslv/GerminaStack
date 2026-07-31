package auth

import "testing"

func TestHashPasswordProducesVerifiableBcryptHash(t *testing.T) {
	t.Parallel()

	hash, err := HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	if hash == "correct horse battery staple" {
		t.Fatal("HashPassword() returned plaintext")
	}
	if err := CheckPassword(hash, "correct horse battery staple"); err != nil {
		t.Fatalf("CheckPassword() error = %v", err)
	}
}

func TestCheckPasswordRejectsWrongPassword(t *testing.T) {
	t.Parallel()

	hash, err := HashPassword("right-password")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	if err := CheckPassword(hash, "wrong-password"); err == nil {
		t.Fatal("CheckPassword() error = nil, want mismatch")
	}
}
