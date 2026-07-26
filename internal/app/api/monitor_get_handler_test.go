package api

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	domainmonitor "github.com/wayzzoo/pulsewarden/internal/domain/monitor"
	repositorypostgres "github.com/wayzzoo/pulsewarden/internal/repository/postgres"
)

type monitorGetterFunc func(
	context.Context,
	uuid.UUID,
) (domainmonitor.Monitor, error)

func (f monitorGetterFunc) Execute(
	ctx context.Context,
	id uuid.UUID,
) (domainmonitor.Monitor, error) {
	return f(ctx, id)
}

func testLogger() *slog.Logger {
	return slog.New(
		slog.NewJSONHandler(io.Discard, nil),
	)
}

func TestGetMonitorHandler(t *testing.T) {
	id := uuid.New()

	now := time.Date(
		2026,
		time.July,
		25,
		12,
		0,
		0,
		0,
		time.UTC,
	)

	server := NewServer(ServerConfig{
		Logger: testLogger(),
		MonitorGetter: monitorGetterFunc(func(
			_ context.Context,
			gotID uuid.UUID,
		) (domainmonitor.Monitor, error) {
			if gotID != id {
				t.Fatalf(
					"monitor ID = %s, want %s",
					gotID,
					id,
				)
			}

			return domainmonitor.Monitor{
				ID:                 id,
				Name:               "Pulsewarden API",
				URL:                "https://example.com/health",
				Method:             "GET",
				Interval:           30 * time.Second,
				Timeout:            1500 * time.Millisecond,
				ExpectedStatusFrom: 200,
				ExpectedStatusTo:   299,
				Enabled:            true,
				NextCheckAt:        now,
				CreatedAt:          now,
				UpdatedAt:          now,
			}, nil
		}),
	})

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/monitors/"+id.String(),
		nil,
	)

	response := httptest.NewRecorder()

	server.Handler.ServeHTTP(response, request)

	result := response.Result()
	defer result.Body.Close()

	if result.StatusCode != http.StatusOK {
		t.Fatalf(
			"status code = %d, want %d",
			result.StatusCode,
			http.StatusOK,
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

	if payload.Name != "Pulsewarden API" {
		t.Fatalf(
			"name = %q, want %q",
			payload.Name,
			"Pulsewarden API",
		)
	}
}

func TestGetMonitorHandlerInvalidID(t *testing.T) {
	server := NewServer(ServerConfig{
		Logger: testLogger(),
		MonitorGetter: monitorGetterFunc(func(
			_ context.Context,
			_ uuid.UUID,
		) (domainmonitor.Monitor, error) {
			t.Fatal("getter must not be called")

			return domainmonitor.Monitor{}, nil
		}),
	})

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/monitors/not-a-uuid",
		nil,
	)

	response := httptest.NewRecorder()

	server.Handler.ServeHTTP(response, request)

	result := response.Result()
	defer result.Body.Close()

	if result.StatusCode != http.StatusBadRequest {
		t.Fatalf(
			"status code = %d, want %d",
			result.StatusCode,
			http.StatusBadRequest,
		)
	}

	var payload struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.NewDecoder(result.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if payload.Error.Code != "invalid_monitor_id" {
		t.Fatalf(
			"error code = %q, want %q",
			payload.Error.Code,
			"invalid_monitor_id",
		)
	}
}

func TestGetMonitorHandlerNotFound(t *testing.T) {
	id := uuid.New()

	server := NewServer(ServerConfig{
		Logger: testLogger(),
		MonitorGetter: monitorGetterFunc(func(
			_ context.Context,
			_ uuid.UUID,
		) (domainmonitor.Monitor, error) {
			return domainmonitor.Monitor{},
				repositorypostgres.ErrMonitorNotFound
		}),
	})

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/monitors/"+id.String(),
		nil,
	)

	response := httptest.NewRecorder()

	server.Handler.ServeHTTP(response, request)

	result := response.Result()
	defer result.Body.Close()

	if result.StatusCode != http.StatusNotFound {
		t.Fatalf(
			"status code = %d, want %d",
			result.StatusCode,
			http.StatusNotFound,
		)
	}

	var payload struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.NewDecoder(result.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if payload.Error.Code != "monitor_not_found" {
		t.Fatalf(
			"error code = %q, want %q",
			payload.Error.Code,
			"monitor_not_found",
		)
	}

	if payload.Error.Message != "monitor not found" {
		t.Fatalf(
			"error message = %q, want %q",
			payload.Error.Message,
			"monitor not found",
		)
	}
}
