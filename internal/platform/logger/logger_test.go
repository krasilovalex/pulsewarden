package logger

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"testing"
)

func TestNewRejectsUnsupportedLevel(t *testing.T) {
	var output bytes.Buffer

	_, err := New(&output, "trace", "api")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestLoggerWritesStructuredFields(t *testing.T) {
	var output bytes.Buffer

	log, err := New(&output, "info", "api")
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	log.Info(
		"application started",
		slog.String("environment", "test"),
		slog.Int("port", 8080),
	)

	var entry map[string]any
	if err := json.Unmarshal(output.Bytes(), &entry); err != nil {
		t.Fatalf("decode log entry: %v", err)
	}

	assertStringField(t, entry, "level", "INFO")
	assertStringField(t, entry, "msg", "application started")
	assertStringField(t, entry, "service", "api")
	assertStringField(t, entry, "environment", "test")

	port, ok := entry["port"].(float64)
	if !ok {
		t.Fatalf("port has unexpected type %T", entry["port"])
	}

	if port != 8080 {
		t.Fatalf("port = %v, want 8080", port)
	}
}

func TestLoggerFiltersMessagesBelowConfiguredLevel(t *testing.T) {
	var output bytes.Buffer

	log, err := New(&output, "warn", "worker")
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	log.Info("ignored")
	if output.Len() != 0 {
		t.Fatalf("unexpected output for filtered message: %q", output.String())
	}

	log.Warn("visible")
	if output.Len() == 0 {
		t.Fatal("expected warning log entry")
	}
}

func assertStringField(
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
