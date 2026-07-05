package middleware

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRecoveryReturnsInternalServerError(t *testing.T) {
	var logs bytes.Buffer

	log := slog.New(slog.NewJSONHandler(&logs, nil))

	next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	})

	request := httptest.NewRequest(http.MethodGet, "/panic", nil)
	request.Header.Set(RequestIDHeader, "test-request-id")

	response := httptest.NewRecorder()

	handler := RequestID(Recovery(log, next))
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf(
			"status code = %d, want %d",
			response.Code,
			http.StatusInternalServerError,
		)
	}

	if got := response.Body.String(); got != "Internal Server Error\n" {
		t.Fatalf("body = %q, want %q", got, "Internal Server Error\n")
	}

	var entry map[string]any
	if err := json.Unmarshal(logs.Bytes(), &entry); err != nil {
		t.Fatalf("decode log entry: %v", err)
	}

	if got := entry["msg"]; got != "panic recovered" {
		t.Fatalf("msg = %v, want panic recovered", got)
	}

	if got := entry["request_id"]; got != "test-request-id" {
		t.Fatalf("request_id = %v, want test-request-id", got)
	}

	if got := entry["path"]; got != "/panic" {
		t.Fatalf("path = %v, want /panic", got)
	}

	if stack, ok := entry["stack"].(string); !ok || stack == "" {
		t.Fatal("stack trace is missing")
	}
}

func TestRecoveryDoesNotOverwriteStartedResponse(t *testing.T) {
	var logs bytes.Buffer

	log := slog.New(slog.NewJSONHandler(&logs, nil))

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusAccepted)

		if _, err := w.Write([]byte("partial response")); err != nil {
			t.Fatalf("write response: %v", err)
		}

		panic("boom")
	})

	request := httptest.NewRequest(http.MethodGet, "/panic", nil)
	response := httptest.NewRecorder()

	Recovery(log, next).ServeHTTP(response, request)

	if response.Code != http.StatusAccepted {
		t.Fatalf(
			"status code = %d, want %d",
			response.Code,
			http.StatusAccepted,
		)
	}

	if got := response.Body.String(); got != "partial response" {
		t.Fatalf("body = %q, want partial response", got)
	}
}

func TestRecoveryPassesSuccessfulRequest(t *testing.T) {
	var logs bytes.Buffer

	log := slog.New(slog.NewJSONHandler(&logs, nil))

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()

	Recovery(log, next).ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf(
			"status code = %d, want %d",
			response.Code,
			http.StatusNoContent,
		)
	}

	if logs.Len() != 0 {
		t.Fatalf("unexpected logs: %s", logs.String())
	}
}

func TestRecoveryRepanicsForAbortHandler(t *testing.T) {
	var logs bytes.Buffer

	log := slog.New(slog.NewJSONHandler(&logs, nil))

	next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic(http.ErrAbortHandler)
	})

	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatal("expected panic")
		}

		err, ok := recovered.(error)
		if !ok || !errors.Is(err, http.ErrAbortHandler) {
			t.Fatalf("recovered panic = %v, want ErrAbortHandler", recovered)
		}
	}()

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()

	Recovery(log, next).ServeHTTP(response, request)
}
