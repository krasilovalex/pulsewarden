package api

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	domainmonitor "github.com/wayzzoo/pulsewarden/internal/domain/monitor"
)

type monitorCreatorFunc func(
	context.Context,
	domainmonitor.NewMonitor,
) (domainmonitor.Monitor, error)

func (f monitorCreatorFunc) Execute(
	ctx context.Context,
	input domainmonitor.NewMonitor,
) (domainmonitor.Monitor, error) {
	return f(ctx, input)
}

func TestCreateMonitorHandler(t *testing.T) {
	now := time.Date(
		2026,
		time.July,
		9,
		12,
		0,
		0,
		0,
		time.UTC,
	)

	id := uuid.New()

	server := NewServer(ServerConfig{
		Logger: slog.New(
			slog.NewJSONHandler(io.Discard, nil),
		),
		MonitorCreator: monitorCreatorFunc(func(
			_ context.Context,
			input domainmonitor.NewMonitor,
		) (domainmonitor.Monitor, error) {
			return domainmonitor.Monitor{
				ID:                 id,
				Name:               input.Name,
				URL:                input.URL,
				Method:             input.Method,
				Interval:           input.Interval,
				Timeout:            input.Timeout,
				ExpectedStatusFrom: input.ExpectedStatusFrom,
				ExpectedStatusTo:   input.ExpectedStatusTo,
				Enabled:            input.Enabled,
				NextCheckAt:        now,
				CreatedAt:          now,
				UpdatedAt:          now,
			}, nil
		}),
	})

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/monitors",
		strings.NewReader(`{
			"name":"Example API",
			"url":"https://example.com/health",
			"method":"GET",
			"interval_seconds":30,
			"timeout_milliseconds":1500,
			"expected_status_from":200,
			"expected_status_to":299,
			"enabled":true
		}`),
	)

	response := httptest.NewRecorder()

	server.Handler.ServeHTTP(response, request)

	result := response.Result()
	defer result.Body.Close()

	if result.StatusCode != http.StatusCreated {
		t.Fatalf(
			"status code = %d, want %d",
			result.StatusCode,
			http.StatusCreated,
		)
	}

	if got := result.Header.Get("Content-Type"); got != "application/json" {
		t.Fatalf(
			"Content-Type = %q, want application/json",
			got,
		)
	}

	var payload createMonitorResponse

	if err := json.NewDecoder(result.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if payload.ID != id.String() {
		t.Fatalf(
			"ID = %q, want %q",
			payload.ID,
			id.String(),
		)
	}

	if payload.Name != "Example API" {
		t.Fatalf(
			"name = %q, want %q",
			payload.Name,
			"Example API",
		)
	}

	if payload.IntervalSeconds != 30 {
		t.Fatalf(
			"interval_seconds = %d, want 30",
			payload.IntervalSeconds,
		)
	}

	if payload.TimeoutMilliseconds != 1500 {
		t.Fatalf(
			"timeout_milliseconds = %d, want 1500",
			payload.TimeoutMilliseconds,
		)
	}
}
