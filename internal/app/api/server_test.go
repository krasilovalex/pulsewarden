package api

import (
	"context"
	"errors"
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
		name       string
		path       string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "liveness",
			path:       "/healthz",
			wantStatus: http.StatusOK,
			wantBody:   "{\"status\":\"ok\"}\n",
		},
		{
			name:       "readiness",
			path:       "/readyz",
			wantStatus: http.StatusOK,
			wantBody:   "{\"status\":\"ready\"}\n",
		},
	}

	server := NewServer(ServerConfig{
		Logger:           newTestLogger(),
		ReadinessTimeout: time.Second,
		Postgres: readinessCheckerFunc(func(context.Context) error {
			return nil
		}),
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := performRequest(
				t,
				server.Handler,
				http.MethodGet,
				tt.path,
			)

			assertJSONResponse(
				t,
				response,
				tt.wantStatus,
				tt.wantBody,
			)
		})
	}
}

func TestReadinessEndpoint(t *testing.T) {
	t.Run("returns ready when checker succeeds", func(t *testing.T) {
		server := NewServer(ServerConfig{
			Logger:           newTestLogger(),
			ReadinessTimeout: time.Second,
			Postgres: readinessCheckerFunc(func(context.Context) error {
				return nil
			}),
		})

		response := performRequest(
			t,
			server.Handler,
			http.MethodGet,
			"/readyz",
		)

		assertJSONResponse(
			t,
			response,
			http.StatusOK,
			"{\"status\":\"ready\"}\n",
		)
	})

	t.Run("returns unavailable when checker fails", func(t *testing.T) {
		server := NewServer(ServerConfig{
			Logger:           newTestLogger(),
			ReadinessTimeout: time.Second,
			Postgres: readinessCheckerFunc(func(context.Context) error {
				return errors.New("database unavailable")
			}),
		})

		response := performRequest(
			t,
			server.Handler,
			http.MethodGet,
			"/readyz",
		)

		assertJSONResponse(
			t,
			response,
			http.StatusServiceUnavailable,
			"{\"status\":\"unavailable\"}\n",
		)
	})

	t.Run("returns unavailable when checker is missing", func(t *testing.T) {
		server := NewServer(ServerConfig{
			Logger:           newTestLogger(),
			ReadinessTimeout: time.Second,
		})

		response := performRequest(
			t,
			server.Handler,
			http.MethodGet,
			"/readyz",
		)

		assertJSONResponse(
			t,
			response,
			http.StatusServiceUnavailable,
			"{\"status\":\"unavailable\"}\n",
		)
	})

	t.Run("returns unavailable on timeout", func(t *testing.T) {
		timeout := 20 * time.Millisecond

		server := NewServer(ServerConfig{
			Logger:           newTestLogger(),
			ReadinessTimeout: timeout,
			Postgres: readinessCheckerFunc(func(ctx context.Context) error {
				<-ctx.Done()
				return ctx.Err()
			}),
		})

		startedAt := time.Now()

		response := performRequest(
			t,
			server.Handler,
			http.MethodGet,
			"/readyz",
		)

		elapsed := time.Since(startedAt)

		assertJSONResponse(
			t,
			response,
			http.StatusServiceUnavailable,
			"{\"status\":\"unavailable\"}\n",
		)

		if elapsed < timeout {
			t.Fatalf(
				"request completed in %s, expected at least %s",
				elapsed,
				timeout,
			)
		}

		if elapsed > time.Second {
			t.Fatalf(
				"request completed in %s, expected timeout close to %s",
				elapsed,
				timeout,
			)
		}
	})

	t.Run("propagates request cancellation to checker", func(t *testing.T) {
		checkerStarted := make(chan struct{})
		checkerCanceled := make(chan error, 1)

		server := NewServer(ServerConfig{
			Logger:           newTestLogger(),
			ReadinessTimeout: time.Second,
			Postgres: readinessCheckerFunc(func(ctx context.Context) error {
				close(checkerStarted)

				<-ctx.Done()
				checkerCanceled <- ctx.Err()

				return ctx.Err()
			}),
		})

		requestContext, cancelRequest := context.WithCancel(
			context.Background(),
		)

		request := httptest.NewRequest(
			http.MethodGet,
			"/readyz",
			nil,
		).WithContext(requestContext)

		response := httptest.NewRecorder()
		requestCompleted := make(chan struct{})

		go func() {
			defer close(requestCompleted)
			server.Handler.ServeHTTP(response, request)
		}()

		select {
		case <-checkerStarted:
		case <-time.After(time.Second):
			t.Fatal("checker did not start")
		}

		cancelRequest()

		select {
		case err := <-checkerCanceled:
			if !errors.Is(err, context.Canceled) {
				t.Fatalf(
					"checker context error = %v, want %v",
					err,
					context.Canceled,
				)
			}
		case <-time.After(time.Second):
			t.Fatal("checker context was not canceled")
		}

		select {
		case <-requestCompleted:
		case <-time.After(time.Second):
			t.Fatal("request did not complete after cancellation")
		}

		result := response.Result()
		defer result.Body.Close()

		if result.StatusCode != http.StatusServiceUnavailable {
			t.Fatalf(
				"status code = %d, want %d",
				result.StatusCode,
				http.StatusServiceUnavailable,
			)
		}
	})
}

func newTestLogger() *slog.Logger {
	return slog.New(
		slog.NewJSONHandler(io.Discard, nil),
	)
}

func performRequest(
	t *testing.T,
	handler http.Handler,
	method string,
	path string,
) *http.Response {
	t.Helper()

	request := httptest.NewRequest(method, path, nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	result := response.Result()

	t.Cleanup(func() {
		_ = result.Body.Close()
	})

	return result
}

func assertJSONResponse(
	t *testing.T,
	response *http.Response,
	wantStatus int,
	wantBody string,
) {
	t.Helper()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}

	if response.StatusCode != wantStatus {
		t.Fatalf(
			"status code = %d, want %d",
			response.StatusCode,
			wantStatus,
		)
	}

	if contentType := response.Header.Get("Content-Type"); contentType != "application/json" {
		t.Fatalf(
			"Content-Type = %q, want application/json",
			contentType,
		)
	}

	if string(body) != wantBody {
		t.Fatalf(
			"body = %q, want %q",
			body,
			wantBody,
		)
	}
}
