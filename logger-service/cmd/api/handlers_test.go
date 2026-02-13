// Package main contains handler tests for logger-service.
package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"log-service/data"
)

func TestHandleWriteLogRejectsInvalidJSON(t *testing.T) {
	app := Config{}

	req := httptest.NewRequest(http.MethodPost, "/log", bytes.NewBufferString("{"))
	rr := httptest.NewRecorder()

	app.handleWriteLog(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}

	response := decodeJSONResponse(t, rr)
	if !response.Error {
		t.Fatalf("expected error=true, got false")
	}
}

func TestHandleWriteLogReturnsServerErrorWhenStoreUnavailable(t *testing.T) {
	app := Config{Models: data.NewModels(nil)}

	body := []byte(`{"name":"event","data":"test"}`)
	req := httptest.NewRequest(http.MethodPost, "/log", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	app.handleWriteLog(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}

	response := decodeJSONResponse(t, rr)
	if !response.Error {
		t.Fatalf("expected error=true, got false")
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
