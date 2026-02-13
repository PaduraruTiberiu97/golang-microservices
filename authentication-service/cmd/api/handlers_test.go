// Package main contains handler-level tests for the authentication API.
package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type roundTripFunc func(req *http.Request) *http.Response

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req), nil
}

func newTestHTTPClient(fn roundTripFunc) *http.Client {
	return &http.Client{
		Transport: fn,
	}
}

func TestHandleAuthenticate(t *testing.T) {
	loggerResponseBody := `{"error": false, "message": "some message"}`

	client := newTestHTTPClient(func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(loggerResponseBody)),
			Header:     make(http.Header),
		}
	})

	testApp.HTTPClient = client

	postBody := map[string]interface{}{
		"email":    "me@here.com",
		"password": "verysecretpassword",
	}

	body, _ := json.Marshal(postBody)

	req, _ := http.NewRequest("POST", "/authenticate", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(testApp.handleAuthenticate)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Errorf("expected http.StatusAccepted, got: %d", rr.Code)
	}
}
