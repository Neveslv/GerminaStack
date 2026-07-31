package auth

import (
	"strings"
	"testing"
)

func TestAuthenticationMessageContainsRequiredSecurityGuidance(t *testing.T) {
	t.Parallel()

	message, err := AuthenticationMessage("ana", "ana@example.com", "123456")
	if err != nil {
		t.Fatalf("AuthenticationMessage() error = %v", err)
	}

	if message.To != "ana@example.com" {
		t.Fatalf("To = %q, want ana@example.com", message.To)
	}
	if message.Subject != "Seu código de autenticação — GerminaStack" {
		t.Fatalf("Subject = %q", message.Subject)
	}
	for _, want := range []string{
		"Olá, ana!",
		"123456",
		"10 minutos",
		"uso único",
		"ignore este e-mail",
		"altere sua senha",
		"Equipe GerminaStack",
	} {
		if !strings.Contains(message.Body, want) {
			t.Fatalf("Body does not contain %q:\n%s", want, message.Body)
		}
	}
}

func TestAuthenticationMessageRejectsHeaderInjection(t *testing.T) {
	t.Parallel()

	if _, err := AuthenticationMessage("ana", "ana@example.com\r\nBcc: attacker@example.com", "123456"); err == nil {
		t.Fatal("AuthenticationMessage() error = nil, want invalid recipient")
	}
}

func TestAuthenticationMessageRejectsMalformedCode(t *testing.T) {
	t.Parallel()

	if _, err := AuthenticationMessage("ana", "ana@example.com", "12345"); err == nil {
		t.Fatal("AuthenticationMessage() error = nil, want invalid code")
	}
}
