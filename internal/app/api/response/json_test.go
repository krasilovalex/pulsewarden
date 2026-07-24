package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteJSON(t *testing.T) {
	recorder := httptest.NewRecorder()

	payload := struct {
		Name string `json:"name"`
	}{
		Name: "Pulsewarden",
	}

	err := WriteJSON(
		recorder,
		http.StatusOK,
		payload,
	)
	if err != nil {
		t.Fatalf("WriteJSON() returned error: %v", err)
	}

	result := recorder.Result()
	defer result.Body.Close()

	if result.StatusCode != http.StatusOK {
		t.Fatalf(
			"status code = %d, want %d",
			result.StatusCode,
			http.StatusOK,
		)
	}

	if got := result.Header.Get("Content-Type"); got != "application/json" {
		t.Fatalf(
			"Content-Type = %q, want application/json",
			got,
		)
	}

	var decoded struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(result.Body).Decode(&decoded); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if decoded.Name != payload.Name {
		t.Fatalf(
			"name = %q, want %q",
			decoded.Name,
			payload.Name,
		)
	}
}
