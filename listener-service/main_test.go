// Package main tests environment helpers for listener-service startup.
package main

import "testing"

func TestGetenvReturnsFallbackWhenUnset(t *testing.T) {
	t.Setenv("LISTENER_TEST_ENV", "")

	value := getenv("LISTENER_TEST_ENV", "fallback")
	if value != "fallback" {
		t.Fatalf("expected fallback value, got %q", value)
	}
}

func TestGetenvReturnsEnvironmentValue(t *testing.T) {
	t.Setenv("LISTENER_TEST_ENV", "configured")

	value := getenv("LISTENER_TEST_ENV", "fallback")
	if value != "configured" {
		t.Fatalf("expected configured value, got %q", value)
	}
}
