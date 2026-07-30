package auth

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestNewSMTPMailerValidatesRequiredConfiguration(t *testing.T) {
	t.Parallel()

	valid := SMTPConfig{
		Host:        "smtp.example.com",
		Port:        587,
		Username:    "mailer",
		Password:    "secret",
		FromAddress: "no-reply@example.com",
		FromName:    "GerminaStack",
		Timeout:     5 * time.Second,
	}
	tests := []struct {
		name   string
		mutate func(*SMTPConfig)
	}{
		{name: "host", mutate: func(c *SMTPConfig) { c.Host = "" }},
		{name: "port", mutate: func(c *SMTPConfig) { c.Port = 0 }},
		{name: "username", mutate: func(c *SMTPConfig) { c.Username = "" }},
		{name: "password", mutate: func(c *SMTPConfig) { c.Password = "" }},
		{name: "from address", mutate: func(c *SMTPConfig) { c.FromAddress = "" }},
		{name: "from name", mutate: func(c *SMTPConfig) { c.FromName = "" }},
		{name: "timeout", mutate: func(c *SMTPConfig) { c.Timeout = 0 }},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := valid
			tt.mutate(&cfg)
			if _, err := NewSMTPMailer(cfg); err == nil {
				t.Fatal("NewSMTPMailer() error = nil, want invalid configuration")
			}
		})
	}
}

func TestSMTPMailerBuildsUTF8MIMEMessageWithoutExposingPassword(t *testing.T) {
	t.Parallel()

	mailer, err := NewSMTPMailer(SMTPConfig{
		Host:        "smtp.example.com",
		Port:        587,
		Username:    "mailer",
		Password:    "smtp-password",
		FromAddress: "no-reply@example.com",
		FromName:    "Equipe GerminaStack",
		Timeout:     5 * time.Second,
	})
	if err != nil {
		t.Fatalf("NewSMTPMailer() error = %v", err)
	}
	raw, err := mailer.wireMessage(Message{
		To:      "ana@example.com",
		Subject: "Seu código de autenticação — GerminaStack",
		Body:    "Código: 123456",
	})
	if err != nil {
		t.Fatalf("wireMessage() error = %v", err)
	}
	text := string(raw)
	for _, want := range []string{"From:", "To: ana@example.com", "Subject:", "MIME-Version: 1.0", "Content-Type: text/plain; charset=UTF-8", "Código: 123456"} {
		if !strings.Contains(text, want) {
			t.Fatalf("wire message missing %q:\n%s", want, text)
		}
	}
	if strings.Contains(text, "smtp-password") {
		t.Fatal("wire message contains SMTP password")
	}
}

func TestSMTPMailerHonorsCanceledContextBeforeDial(t *testing.T) {
	t.Parallel()

	mailer, err := NewSMTPMailer(SMTPConfig{
		Host:        "smtp.example.com",
		Port:        587,
		Username:    "mailer",
		Password:    "smtp-password",
		FromAddress: "no-reply@example.com",
		FromName:    "GerminaStack",
		Timeout:     5 * time.Second,
	})
	if err != nil {
		t.Fatalf("NewSMTPMailer() error = %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err = mailer.Send(ctx, Message{To: "ana@example.com", Subject: "subject", Body: "body"})
	if err == nil {
		t.Fatal("Send() error = nil, want canceled context")
	}
}
