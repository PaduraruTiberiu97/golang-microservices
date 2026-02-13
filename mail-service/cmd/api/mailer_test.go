// Package main tests small pure mailer helpers.
package main

import (
	"testing"

	mail "github.com/xhit/go-simple-mail/v2"
)

func TestResolveEncryption(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected mail.Encryption
	}{
		{name: "tls", input: "tls", expected: mail.EncryptionSTARTTLS},
		{name: "ssl", input: "ssl", expected: mail.EncryptionSSLTLS},
		{name: "none", input: "none", expected: mail.EncryptionNone},
		{name: "default", input: "unexpected", expected: mail.EncryptionSTARTTLS},
	}

	mailer := Mail{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mailer.resolveEncryption(tt.input); got != tt.expected {
				t.Fatalf("expected %v, got %v", tt.expected, got)
			}
		})
	}
}
