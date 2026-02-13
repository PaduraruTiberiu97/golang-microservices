// Package main tests environment-driven mail-service configuration helpers.
package main

import "testing"

func TestMailerFromEnvUsesConfiguredValues(t *testing.T) {
	t.Setenv("MAIL_PORT", "2525")
	t.Setenv("MAIL_DOMAIN", "example.com")
	t.Setenv("MAIL_HOST", "smtp.example.com")
	t.Setenv("MAIL_USERNAME", "user")
	t.Setenv("MAIL_PASSWORD", "pass")
	t.Setenv("MAIL_ENCRYPTION", "tls")
	t.Setenv("MAIL_NAME", "Example")
	t.Setenv("MAIL_ADDRESS", "noreply@example.com")

	mailer := mailerFromEnv()

	if mailer.Port != 2525 {
		t.Fatalf("expected port 2525, got %d", mailer.Port)
	}
	if mailer.Domain != "example.com" {
		t.Fatalf("expected domain example.com, got %q", mailer.Domain)
	}
	if mailer.Host != "smtp.example.com" {
		t.Fatalf("expected host smtp.example.com, got %q", mailer.Host)
	}
}

func TestMailerFromEnvFallsBackToDefaultPort(t *testing.T) {
	t.Setenv("MAIL_PORT", "invalid")

	mailer := mailerFromEnv()
	if mailer.Port != 25 {
		t.Fatalf("expected default port 25, got %d", mailer.Port)
	}
}
