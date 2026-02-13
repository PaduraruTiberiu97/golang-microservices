// Package main tests environment helpers for front-end template data.
package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

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

func TestFrontendPortFromEnvUsesDefaultWhenUnset(t *testing.T) {
	t.Setenv("FRONTEND_PORT", "")

	port := frontendPortFromEnv()
	if port != "8081" {
		t.Fatalf("expected default frontend port, got %q", port)
	}
}

func TestFrontendPortFromEnvUsesEnvironmentValue(t *testing.T) {
	t.Setenv("FRONTEND_PORT", "8099")

	port := frontendPortFromEnv()
	if port != "8099" {
		t.Fatalf("expected env frontend port, got %q", port)
	}
}

func TestRenderTemplateRendersDashboard(t *testing.T) {
	rr := httptest.NewRecorder()

	renderTemplate(rr, "test.page.gohtml")

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	if !strings.Contains(rr.Body.String(), "Microservice Control Room") {
		t.Fatalf("expected dashboard headline to be rendered")
	}
}
