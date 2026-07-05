package middleware

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAccessLogWritesRequestInformation(t *testing.T) {
	var output bytes.Buffer

	log := slog.New(slog.NewJSONHandler(&output, nil))

	times := []time.Time{
		time.Date(2026, 7, 4, 12, 0, 0, 0, time.UTC),
		time.Date(2026, 7, 4, 12, 0, 0, 0, time.UTC).
			Add(125 * time.Millisecond),
	}

	index := 0
	now := func() time.Time {
		value := times[index]
		index++
		return value
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)

		if _, err := w.Write([]byte("hello")); err != nil {
			t.Fatalf("write response: %v", err)
		}
	})

	request := httptest.NewRequest(
		http.MethodPost,
		"/monitors?token=secret",
		nil,
	)
	request.Header.Set(RequestIDHeader, "test-request-id")
	request.Header.Set("User-Agent", "pulsewarden-test")
	request.RemoteAddr = "192.0.2.1:1234"

	response := httptest.NewRecorder()

	handler := RequestID(accessLog(log, next, now))
	handler.ServeHTTP(response, request)

	var entry map[string]any
	if err := json.Unmarshal(output.Bytes(), &entry); err != nil {
		t.Fatalf("decode log entry: %v", err)
	}

	assertAccessLogString(t, entry, "msg", "HTTP request completed")
	assertAccessLogString(t, entry, "request_id", "test-request-id")
	assertAccessLogString(t, entry, "method", http.MethodPost)
	assertAccessLogString(t, entry, "path", "/monitors")
	assertAccessLogString(t, entry, "remote_addr", "192.0.2.1:1234")
	assertAccessLogString(t, entry, "user_agent", "pulsewarden-test")

	assertAccessLogNumber(t, entry, "status", http.StatusCreated)
	assertAccessLogNumber(t, entry, "response_bytes", 5)

	duration, ok := entry["duration_ms"].(float64)
	if !ok {
		t.Fatalf(
			"duration_ms has unexpected type %T",
			entry["duration_ms"],
		)
	}

	if duration != 125 {
		t.Fatalf("duration_ms = %v, want 125", duration)
	}
}

func TestAccessLogUsesOKWhenHandlerDoesNotWriteResponse(t *testing.T) {
	var output bytes.Buffer

	log := slog.New(slog.NewJSONHandler(&output, nil))

	now := func() time.Time {
		return time.Time{}
	}

	next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()

	accessLog(log, next, now).ServeHTTP(response, request)

	var entry map[string]any
	if err := json.Unmarshal(output.Bytes(), &entry); err != nil {
		t.Fatalf("decode log entry: %v", err)
	}

	assertAccessLogNumber(t, entry, "status", http.StatusOK)
	assertAccessLogNumber(t, entry, "response_bytes", 0)
}

func assertAccessLogString(
	t *testing.T,
	entry map[string]any,
	field string,
	want string,
) {
	t.Helper()

	got, ok := entry[field].(string)
	if !ok {
		t.Fatalf("%s has unexpected type %T", field, entry[field])
	}

	if got != want {
		t.Fatalf("%s = %q, want %q", field, got, want)
	}
}

func assertAccessLogNumber(
	t *testing.T,
	entry map[string]any,
	field string,
	want int,
) {
	t.Helper()

	got, ok := entry[field].(float64)
	if !ok {
		t.Fatalf("%s has unexpected type %T", field, entry[field])
	}

	if got != float64(want) {
		t.Fatalf("%s = %v, want %d", field, got, want)
	}
}
