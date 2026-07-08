package api

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type readinessCheckerFunc func(context.Context) error

func (f readinessCheckerFunc) Ping(ctx context.Context) error {
	return f(ctx)
}

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

	testLogger := slog.New(
		slog.NewJSONHandler(io.Discard, nil),
	)

	server := NewServer(ServerConfig{
		Logger:           testLogger,
		ReadinessTimeout: time.Second,
		Postgres: readinessCheckerFunc(func(context.Context) error {
			return nil
		}),
	})

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
