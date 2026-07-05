package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteError(t *testing.T) {
	response := httptest.NewRecorder()

	err := WriteError(
		response,
		http.StatusBadRequest,
		"invalid_request",
		"request is invalid",
		"test-request-id",
	)
	if err != nil {
		t.Fatalf("WriteError() returned error: %v", err)
	}

	result := response.Result()
	defer result.Body.Close()

	if result.StatusCode != http.StatusBadRequest {
		t.Fatalf(
			"status code = %d, want %d",
			result.StatusCode,
			http.StatusBadRequest,
		)
	}

	if got := result.Header.Get("Content-Type"); got != "application/json" {
		t.Fatalf(
			"Content-Type = %q, want application/json",
			got,
		)
	}

	var payload ErrorEnvelope

	if err := json.NewDecoder(result.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	want := ErrorEnvelope{
		Error: ErrorBody{
			Code:      "invalid_request",
			Message:   "request is invalid",
			RequestID: "test-request-id",
		},
	}

	if payload != want {
		t.Fatalf("payload = %+v, want %+v", payload, want)
	}
}
