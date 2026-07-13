package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreateMonitorRequestToDomain(t *testing.T) {
	enabled := false

	request := createMonitorRequest{
		Name:                "Example API",
		URL:                 "https://example.com/health",
		Method:              http.MethodGet,
		IntervalSeconds:     30,
		TimeoutMilliseconds: 1500,
		ExpectedStatusFrom:  200,
		ExpectedStatusTo:    299,
		Enabled:             &enabled,
	}

	result := request.toDomain()

	if result.Name != request.Name {
		t.Fatalf("name = %q, want %q", result.Name, request.Name)
	}

	if result.Interval.Seconds() != 30 {
		t.Fatalf("interval = %s, want 30s", result.Interval)
	}

	if result.Timeout.Milliseconds() != 1500 {
		t.Fatalf("timeout = %s, want 1500ms", result.Timeout)
	}

	if result.Enabled {
		t.Fatal("enabled = true, want false")
	}
}

func TestCreateMonitorRequestDefaultsEnabledToTrue(t *testing.T) {
	result := (createMonitorRequest{}).toDomain()

	if !result.Enabled {
		t.Fatal("enabled = false, want true")
	}
}

func TestDecodeJSONRejectsUnknownFields(t *testing.T) {
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/monitors",
		strings.NewReader(`{"name":"test","unknown":true}`),
	)

	response := httptest.NewRecorder()

	var payload createMonitorRequest

	err := decodeJSON(response, request, &payload)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDecodeJSONRejectsMultipleObjects(t *testing.T) {
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/monitors",
		strings.NewReader(`{"name":"first"} {"name":"second"}`),
	)

	response := httptest.NewRecorder()

	var payload createMonitorRequest

	err := decodeJSON(response, request, &payload)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
