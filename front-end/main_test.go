// Package main tests environment helpers for front-end template data.
package main

import "testing"

func TestBrokerURLFromEnvUsesDefaultWhenUnset(t *testing.T) {
	t.Setenv("BROKER_URL", "")

	url := brokerURLFromEnv()
	if url != "http://localhost:8000" {
		t.Fatalf("expected default broker URL, got %q", url)
	}
}

func TestBrokerURLFromEnvUsesEnvironmentValue(t *testing.T) {
	t.Setenv("BROKER_URL", "http://example.com")

	url := brokerURLFromEnv()
	if url != "http://example.com" {
		t.Fatalf("expected env broker URL, got %q", url)
	}
}
