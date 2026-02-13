// Package event tests logger forwarding helpers used by broker-service consumers.
package event

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoggerServiceURLUsesDefaultWhenUnset(t *testing.T) {
	t.Setenv("LOGGER_SERVICE_URL", "")

	url := loggerServiceURL()
	if url != "http://logger-service/log" {
		t.Fatalf("expected default URL, got %q", url)
	}
}

func TestLoggerServiceURLUsesEnvironmentOverride(t *testing.T) {
	t.Setenv("LOGGER_SERVICE_URL", "http://example.com/log")

	url := loggerServiceURL()
	if url != "http://example.com/log" {
		t.Fatalf("expected env URL, got %q", url)
	}
}

func TestForwardLogEventSuccess(t *testing.T) {
	logger := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected method POST, got %s", r.Method)
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer logger.Close()

	t.Setenv("LOGGER_SERVICE_URL", logger.URL)

	err := forwardLogEvent(Payload{Name: "event", Data: "payload"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestForwardLogEventReturnsErrorOnNon2xx(t *testing.T) {
	logger := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer logger.Close()

	t.Setenv("LOGGER_SERVICE_URL", logger.URL)

	err := forwardLogEvent(Payload{Name: "event", Data: "payload"})
	if err == nil {
		t.Fatalf("expected an error when logger service returns non-2xx")
	}
}
