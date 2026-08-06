package config

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestLoadReadsRequiredConfiguration(t *testing.T) {
	setValidEnvironment(t)
	t.Setenv("COOKIE_SECURE", "false")
	t.Setenv("HTTP_ADDR", "127.0.0.1:9090")
	t.Setenv("AUTH_OPERATION_TIMEOUT", "3s")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.DatabaseURL != "postgres://app:password@db.example.com/germina" {
		t.Fatalf("DatabaseURL = %q", cfg.DatabaseURL)
	}
	if cfg.JWTSecret != testJWTSecret || cfg.TwoFactorSecret != testTwoFactorSecret {
		t.Fatal("authentication secrets were not loaded")
	}
	if cfg.AuthOperationTimeout != 3*time.Second {
		t.Fatalf("AuthOperationTimeout = %v, want 3s", cfg.AuthOperationTimeout)
	}
	if cfg.CookieSecure {
		t.Fatal("CookieSecure = true, want false")
	}
	if cfg.HTTPAddr != "127.0.0.1:9090" {
		t.Fatalf("HTTPAddr = %q", cfg.HTTPAddr)
	}
	if cfg.SMTP.Host != "smtp.example.com" || cfg.SMTP.Port != 587 || cfg.SMTP.FromName != "GerminaStack" {
		t.Fatalf("SMTP = %#v", cfg.SMTP)
	}
}

func TestLoadDefaultsCookieSecureAndHTTPAddress(t *testing.T) {
	setValidEnvironment(t)
	t.Setenv("COOKIE_SECURE", "")
	t.Setenv("HTTP_ADDR", "")
	t.Setenv("AUTH_OPERATION_TIMEOUT", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if !cfg.CookieSecure {
		t.Fatal("CookieSecure = false, want secure default")
	}
	if cfg.HTTPAddr != ":8080" {
		t.Fatalf("HTTPAddr = %q, want :8080", cfg.HTTPAddr)
	}
	if cfg.AuthOperationTimeout != 15*time.Second {
		t.Fatalf("AuthOperationTimeout = %v, want 15s", cfg.AuthOperationTimeout)
	}
}

func TestLoadRejectsEveryMissingRequiredVariableWithoutLeakingValue(t *testing.T) {
	required := []string{
		"DATABASE_URL",
		"JWT_SECRET",
		"TWO_FACTOR_SECRET",
		"SMTP_HOST",
		"SMTP_PORT",
		"SMTP_USERNAME",
		"SMTP_PASSWORD",
		"SMTP_FROM_ADDRESS",
		"SMTP_FROM_NAME",
	}
	for _, variable := range required {
		variable := variable
		t.Run(variable, func(t *testing.T) {
			setValidEnvironment(t)
			t.Setenv(variable, "")

			_, err := Load()
			if err == nil {
				t.Fatal("Load() error = nil, want missing variable")
			}
			if strings.Contains(err.Error(), "smtp-test-password") ||
				strings.Contains(err.Error(), testJWTSecret) ||
				strings.Contains(err.Error(), testTwoFactorSecret) {
				t.Fatalf("Load() error leaked secret: %v", err)
			}
		})
	}
}

func TestLoadRejectsInvalidPortAndCookieBoolean(t *testing.T) {
	t.Run("port", func(t *testing.T) {
		setValidEnvironment(t)
		t.Setenv("SMTP_PORT", "not-a-port")
		if _, err := Load(); err == nil {
			t.Fatal("Load() error = nil, want invalid port")
		}
	})
	t.Run("cookie", func(t *testing.T) {
		setValidEnvironment(t)
		t.Setenv("COOKIE_SECURE", "sometimes")
		if _, err := Load(); err == nil {
			t.Fatal("Load() error = nil, want invalid cookie boolean")
		}
	})
	for _, value := range []string{"not-a-duration", "0s", "-1s"} {
		value := value
		t.Run("operation timeout "+value, func(t *testing.T) {
			setValidEnvironment(t)
			t.Setenv("AUTH_OPERATION_TIMEOUT", value)
			if _, err := Load(); err == nil {
				t.Fatal("Load() error = nil, want invalid operation timeout")
			}
		})
	}
}

func TestLoadRejectsWeakOrReusedAuthenticationSecrets(t *testing.T) {
	tests := []struct {
		name      string
		jwtSecret string
		twoFactor string
	}{
		{name: "short JWT secret", jwtSecret: strings.Repeat("j", 31), twoFactor: testTwoFactorSecret},
		{name: "short two factor secret", jwtSecret: testJWTSecret, twoFactor: strings.Repeat("t", 31)},
		{name: "same secret", jwtSecret: testJWTSecret, twoFactor: testJWTSecret},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			setValidEnvironment(t)
			t.Setenv("JWT_SECRET", tt.jwtSecret)
			t.Setenv("TWO_FACTOR_SECRET", tt.twoFactor)

			_, err := Load()
			if err == nil {
				t.Fatal("Load() error = nil, want rejected authentication secrets")
			}
			if strings.Contains(err.Error(), tt.jwtSecret) || strings.Contains(err.Error(), tt.twoFactor) {
				t.Fatalf("Load() error leaked a secret: %v", err)
			}
		})
	}
}

func TestExampleAuthenticationSecretsCannotStartApplication(t *testing.T) {
	example := readExampleEnvironment(t)
	setValidEnvironment(t)
	t.Setenv("JWT_SECRET", example["JWT_SECRET"])
	t.Setenv("TWO_FACTOR_SECRET", example["TWO_FACTOR_SECRET"])

	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, .env.example authentication secrets must require replacement")
	}
}

const (
	testJWTSecret       = "jwt-secret-with-at-least-32-bytes-a"
	testTwoFactorSecret = "two-factor-secret-with-32-bytes-b"
)

func setValidEnvironment(t *testing.T) {
	t.Helper()
	t.Setenv("DATABASE_URL", "postgres://app:password@db.example.com/germina")
	t.Setenv("JWT_SECRET", testJWTSecret)
	t.Setenv("TWO_FACTOR_SECRET", testTwoFactorSecret)
	t.Setenv("SMTP_HOST", "smtp.example.com")
	t.Setenv("SMTP_PORT", "587")
	t.Setenv("SMTP_USERNAME", "mailer")
	t.Setenv("SMTP_PASSWORD", "smtp-test-password")
	t.Setenv("SMTP_FROM_ADDRESS", "no-reply@example.com")
	t.Setenv("SMTP_FROM_NAME", "GerminaStack")
	t.Setenv("COOKIE_SECURE", "true")
	t.Setenv("HTTP_ADDR", ":8080")
	t.Setenv("AUTH_OPERATION_TIMEOUT", "15s")
}

func readExampleEnvironment(t *testing.T) map[string]string {
	t.Helper()
	content, err := os.ReadFile("../.env.example")
	if err != nil {
		t.Fatalf("read .env.example: %v", err)
	}

	values := make(map[string]string)
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if ok {
			values[key] = value
		}
	}
	return values
}
