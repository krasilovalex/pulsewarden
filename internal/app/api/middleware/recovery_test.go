package middleware

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/krasilovalex/pulsewarden/internal/app/api/response"
)

func TestRecoveryReturnsInternalServerError(t *testing.T) {
	var logs bytes.Buffer

	log := slog.New(slog.NewJSONHandler(&logs, nil))

	next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	})

	request := httptest.NewRequest(http.MethodGet, "/panic", nil)
	request.Header.Set(RequestIDHeader, "test-request-id")

	recorder := httptest.NewRecorder()

	handler := RequestID(Recovery(log, next))
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"status code = %d, want %d",
			recorder.Code,
			http.StatusInternalServerError,
		)
	}

	var payload response.ErrorEnvelope

	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response body: %v", err)
	}

	wantPayload := response.ErrorEnvelope{
		Error: response.ErrorBody{
			Code:      "internal_error",
			Message:   "internal server error",
			RequestID: "test-request-id",
		},
	}

	if payload != wantPayload {
		t.Fatalf(
			"response payload = %+v, want %+v",
			payload,
			wantPayload,
		)
	}

	if got := recorder.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf(
			"Content-Type = %q, want application/json",
			got,
		)
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
