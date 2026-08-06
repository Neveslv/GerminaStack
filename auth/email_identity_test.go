package auth

import "testing"

func TestParseInstitutionalEmailDerivesIdentity(t *testing.T) {
	t.Parallel()

	identity, err := ParseInstitutionalEmail("maria.silva@institutojef.org.br")
	if err != nil {
		t.Fatalf("ParseInstitutionalEmail() error = %v", err)
	}
	if identity.Name != "Maria Silva" || identity.Username != "maria.silva" {
		t.Fatalf("identity = %#v", identity)
	}
}

func TestParseInstitutionalEmailRejectsNonInstitutionalIdentity(t *testing.T) {
	t.Parallel()

	invalid := []string{
		"Maria.Silva@institutojef.org.br",
		"maria.silva@example.com",
		"maria@institutojef.org.br",
		"maria.silva.souza@institutojef.org.br",
		"maria-silva@institutojef.org.br",
		" maria.silva@institutojef.org.br",
		"maria.silva@institutojef.org.br ",
	}
	for _, email := range invalid {
		email := email
		t.Run(email, func(t *testing.T) {
			t.Parallel()
			if _, err := ParseInstitutionalEmail(email); err == nil {
				t.Fatalf("ParseInstitutionalEmail(%q) error = nil, want rejection", email)
			}
		})
	}
}
