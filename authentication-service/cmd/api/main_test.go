// Package main tests process-level configuration helpers for authentication-service.
package main

import "testing"

func TestGetenvReturnsFallbackWhenUnset(t *testing.T) {
	t.Setenv("AUTH_TEST_ENV", "")

	value := getenv("AUTH_TEST_ENV", "fallback")
	if value != "fallback" {
		t.Fatalf("expected fallback value, got %q", value)
	}
}

func TestGetenvReturnsEnvironmentValue(t *testing.T) {
	t.Setenv("AUTH_TEST_ENV", "configured")

	value := getenv("AUTH_TEST_ENV", "fallback")
	if value != "configured" {
		t.Fatalf("expected configured value, got %q", value)
	}
}
