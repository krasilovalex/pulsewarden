package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthEndpoints(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		wantBody string
	}{
		{
			name:     "liveness",
			path:     "/healthz",
			wantBody: "{\"status\":\"ok\"}\n",
		},
		{
			name:     "readiness",
			path:     "/readyz",
			wantBody: "{\"status\":\"ready\"}\n",
		},
	}

	server := NewServer(ServerConfig{})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, tt.path, nil)
			response := httptest.NewRecorder()

			server.Handler.ServeHTTP(response, request)

			result := response.Result()
			defer result.Body.Close()

			body, err := io.ReadAll(result.Body)
			if err != nil {
				t.Fatalf("read response body: %v", err)
			}

			if result.StatusCode != http.StatusOK {
				t.Fatalf(
					"status code = %d, want %d",
					result.StatusCode,
					http.StatusOK,
				)
			}

			if contentType := result.Header.Get("Content-Type"); contentType != "application/json" {
				t.Fatalf(
					"Content-Type = %q, want application/json",
					contentType,
				)
			}

			if string(body) != tt.wantBody {
				t.Fatalf("body = %q, want %q", body, tt.wantBody)
			}
		})
	}
}
