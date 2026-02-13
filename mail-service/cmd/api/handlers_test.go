// Package main contains handler tests for mail-service.
package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleSendMailRejectsInvalidJSON(t *testing.T) {
	app := Config{}

	req := httptest.NewRequest(http.MethodPost, "/send", bytes.NewBufferString("{"))
	rr := httptest.NewRecorder()

	app.handleSendMail(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}

	response := decodeJSONResponse(t, rr)
	if !response.Error {
		t.Fatalf("expected error=true, got false")
	}
}

func TestHandleSendMailRejectsMissingFields(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{
			name: "missing to",
			body: `{"subject":"test","message":"hello"}`,
		},
		{
			name: "missing subject",
			body: `{"to":"me@example.com","message":"hello"}`,
		},
		{
			name: "missing message",
			body: `{"to":"me@example.com","subject":"test"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := Config{}

			req := httptest.NewRequest(http.MethodPost, "/send", bytes.NewBufferString(tt.body))
			rr := httptest.NewRecorder()

			app.handleSendMail(rr, req)

			if rr.Code != http.StatusBadRequest {
				t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
			}

			response := decodeJSONResponse(t, rr)
			if !response.Error {
				t.Fatalf("expected error=true, got false")
			}
		})
	}
}

func decodeJSONResponse(t *testing.T, rr *httptest.ResponseRecorder) JsonResponse {
	t.Helper()

	var response JsonResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode JSON response: %v", err)
	}

	return response
}
