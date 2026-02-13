// Package main contains broker handler and forwarding tests.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"net/rpc"
	"testing"

	"broker/logs"
	"google.golang.org/grpc"
)

func TestHandleBroker(t *testing.T) {
	app := Config{}

	req := httptest.NewRequest(http.MethodPost, "/", http.NoBody)
	rr := httptest.NewRecorder()

	app.handleBroker(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	response := decodeJSONResponse(t, rr)
	if response.Error {
		t.Fatalf("expected error=false, got true")
	}
	if response.Message != "Hit the broker" {
		t.Fatalf("expected broker message, got %q", response.Message)
	}
}

func TestHandleSubmissionInvalidAction(t *testing.T) {
	app := Config{}

	body := `{"action":"unknown"}`
	req := httptest.NewRequest(http.MethodPost, "/handle", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()

	app.handleSubmission(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}

	response := decodeJSONResponse(t, rr)
	if !response.Error {
		t.Fatalf("expected error=true, got false")
	}
	if response.Message != "invalid action" {
		t.Fatalf("expected invalid action message, got %q", response.Message)
	}
}

func TestForwardAuthRequestAccepted(t *testing.T) {
	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected method POST, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(JsonResponse{
			Error:   false,
			Message: "ok",
			Data: map[string]string{
				"id": "123",
			},
		})
	}))
	defer authServer.Close()

	app := Config{AuthServiceURL: authServer.URL}
	rr := httptest.NewRecorder()

	app.forwardAuthRequest(rr, AuthPayload{Email: "me@example.com", Pass: "secret"})

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	response := decodeJSONResponse(t, rr)
	if response.Error {
		t.Fatalf("expected error=false, got true")
	}
	if response.Message != "Authenticated!" {
		t.Fatalf("expected authenticated message, got %q", response.Message)
	}
}

func TestForwardAuthRequestUnauthorized(t *testing.T) {
	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer authServer.Close()

	app := Config{AuthServiceURL: authServer.URL}
	rr := httptest.NewRecorder()

	app.forwardAuthRequest(rr, AuthPayload{Email: "me@example.com", Pass: "wrong"})

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}

	response := decodeJSONResponse(t, rr)
	if !response.Error {
		t.Fatalf("expected error=true, got false")
	}
}

func TestForwardMailRequestSuccess(t *testing.T) {
	mailServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}))
	defer mailServer.Close()

	app := Config{MailServiceURL: mailServer.URL}
	rr := httptest.NewRecorder()

	app.forwardMailRequest(rr, MailPayload{From: "a@b.com", To: "c@d.com", Subject: "s", Message: "m"})

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	response := decodeJSONResponse(t, rr)
	if response.Error {
		t.Fatalf("expected error=false, got true")
	}
	if response.Message != "Mail sent" {
		t.Fatalf("expected mail sent message, got %q", response.Message)
	}
}

func TestForwardMailRequestBadGatewayOnDownstreamFailure(t *testing.T) {
	mailServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer mailServer.Close()

	app := Config{MailServiceURL: mailServer.URL}
	rr := httptest.NewRecorder()

	app.forwardMailRequest(rr, MailPayload{From: "a@b.com", To: "c@d.com", Subject: "s", Message: "m"})

	if rr.Code != http.StatusBadGateway {
		t.Fatalf("expected status %d, got %d", http.StatusBadGateway, rr.Code)
	}

	response := decodeJSONResponse(t, rr)
	if !response.Error {
		t.Fatalf("expected error=true, got false")
	}
}

func TestForwardLogRequestHTTPSuccess(t *testing.T) {
	logServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}))
	defer logServer.Close()

	app := Config{LoggerServiceURL: logServer.URL}
	rr := httptest.NewRecorder()

	app.forwardLogRequestHTTP(rr, LogPayload{Name: "test", Data: "payload"})

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	response := decodeJSONResponse(t, rr)
	if response.Error {
		t.Fatalf("expected error=false, got true")
	}
	if response.Message != "Logged" {
		t.Fatalf("expected logged message, got %q", response.Message)
	}
}

func TestForwardLogRequestHTTPBadGatewayOnDownstreamFailure(t *testing.T) {
	logServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer logServer.Close()

	app := Config{LoggerServiceURL: logServer.URL}
	rr := httptest.NewRecorder()

	app.forwardLogRequestHTTP(rr, LogPayload{Name: "test", Data: "payload"})

	if rr.Code != http.StatusBadGateway {
		t.Fatalf("expected status %d, got %d", http.StatusBadGateway, rr.Code)
	}

	response := decodeJSONResponse(t, rr)
	if !response.Error {
		t.Fatalf("expected error=true, got false")
	}
}

func TestLogViaRPC(t *testing.T) {
	address, cleanup := startTestRPCServer(t)
	defer cleanup()

	app := Config{LoggerRPCAddr: address}
	rr := httptest.NewRecorder()

	app.logViaRPC(rr, LogPayload{Name: "rpc-test", Data: "payload"})

	if rr.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d", http.StatusAccepted, rr.Code)
	}

	response := decodeJSONResponse(t, rr)
	if response.Error {
		t.Fatalf("expected error=false, got true")
	}
	if response.Message == "" {
		t.Fatalf("expected a success message from RPC call")
	}
}

func TestLogViaGRPCRejectsInvalidJSON(t *testing.T) {
	app := Config{}

	req := httptest.NewRequest(http.MethodPost, "/log-grpc", bytes.NewBufferString("{"))
	rr := httptest.NewRecorder()

	app.logViaGRPC(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}

	response := decodeJSONResponse(t, rr)
	if !response.Error {
		t.Fatalf("expected error=true, got false")
	}
}

func TestLogViaGRPC(t *testing.T) {
	address, cleanup := startTestGRPCServer(t)
	defer cleanup()

	app := Config{LoggerGRPCAddr: address}

	body, err := json.Marshal(RequestPayload{
		Action: "log",
		Log: LogPayload{
			Name: "grpc-test",
			Data: "payload",
		},
	})
	if err != nil {
		t.Fatalf("failed to marshal request body: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/log-grpc", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	app.logViaGRPC(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d", http.StatusAccepted, rr.Code)
	}

	response := decodeJSONResponse(t, rr)
	if response.Error {
		t.Fatalf("expected error=false, got true")
	}
	if response.Message != "Logged via GRPC" {
		t.Fatalf("expected grpc success message, got %q", response.Message)
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

type rpcTestServer struct{}

func (rpcTestServer) LogInfo(payload RPCPayload, result *string) error {
	*result = "Processed payload via RPC: " + payload.Name
	return nil
}

func startTestRPCServer(t *testing.T) (string, func()) {
	t.Helper()

	testServer := rpc.NewServer()
	if err := testServer.RegisterName("RPCServer", rpcTestServer{}); err != nil {
		t.Fatalf("failed to register RPC test server: %v", err)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start RPC listener: %v", err)
	}

	done := make(chan struct{})
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				select {
				case <-done:
					return
				default:
					return
				}
			}

			go testServer.ServeConn(conn)
		}
	}()

	cleanup := func() {
		close(done)
		_ = listener.Close()
	}

	return listener.Addr().String(), cleanup
}

type grpcTestLogServer struct {
	logs.UnimplementedLogServiceServer
}

func (grpcTestLogServer) Write(ctx context.Context, req *logs.LogRequest) (*logs.LogResponse, error) {
	return &logs.LogResponse{Result: "success"}, nil
}

func startTestGRPCServer(t *testing.T) (string, func()) {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start gRPC listener: %v", err)
	}

	grpcServer := grpc.NewServer()
	logs.RegisterLogServiceServer(grpcServer, grpcTestLogServer{})

	go func() {
		_ = grpcServer.Serve(listener)
	}()

	cleanup := func() {
		grpcServer.GracefulStop()
		_ = listener.Close()
	}

	return listener.Addr().String(), cleanup
}
