package config

import (
	"strings"
	"testing"
)

func TestLoadReadsRequiredConfiguration(t *testing.T) {
	setValidEnvironment(t)
	t.Setenv("COOKIE_SECURE", "false")
	t.Setenv("HTTP_ADDR", "127.0.0.1:9090")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.DatabaseURL != "postgres://app:password@db.example.com/germina" {
		t.Fatalf("DatabaseURL = %q", cfg.DatabaseURL)
	}
	if cfg.JWTSecret != "jwt-test-secret" || cfg.TwoFactorSecret != "two-factor-test-secret" {
		t.Fatal("authentication secrets were not loaded")
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
				strings.Contains(err.Error(), "jwt-test-secret") ||
				strings.Contains(err.Error(), "two-factor-test-secret") {
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
}

func setValidEnvironment(t *testing.T) {
	t.Helper()
	t.Setenv("DATABASE_URL", "postgres://app:password@db.example.com/germina")
	t.Setenv("JWT_SECRET", "jwt-test-secret")
	t.Setenv("TWO_FACTOR_SECRET", "two-factor-test-secret")
	t.Setenv("SMTP_HOST", "smtp.example.com")
	t.Setenv("SMTP_PORT", "587")
	t.Setenv("SMTP_USERNAME", "mailer")
	t.Setenv("SMTP_PASSWORD", "smtp-test-password")
	t.Setenv("SMTP_FROM_ADDRESS", "no-reply@example.com")
	t.Setenv("SMTP_FROM_NAME", "GerminaStack")
	t.Setenv("COOKIE_SECURE", "true")
	t.Setenv("HTTP_ADDR", ":8080")
}
