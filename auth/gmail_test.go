package auth

import (
	"strings"
	"testing"
)

func TestWireGmailMessageDoesNotRequireSMTPConfiguration(t *testing.T) {
	t.Parallel()
	raw, err := wireGmailMessage("no-reply@example.com", Message{To: "ana@example.com", Subject: "Código", Body: "123456"})
	if err != nil {
		t.Fatalf("wireGmailMessage() error = %v", err)
	}
	if !strings.Contains(string(raw), "From:") || !strings.Contains(string(raw), "<no-reply@example.com>") {
		t.Fatalf("wireGmailMessage() = %q", raw)
	}
}
