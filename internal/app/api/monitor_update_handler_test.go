package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	domainmonitor "github.com/wayzzoo/pulsewarden/internal/domain/monitor"
	repositorypostgres "github.com/wayzzoo/pulsewarden/internal/repository/postgres"
)

type monitorUpdaterFunc func(
	context.Context,
	uuid.UUID,
	domainmonitor.UpdateMonitor,
) (domainmonitor.Monitor, error)

func (f monitorUpdaterFunc) Execute(
	ctx context.Context,
	id uuid.UUID,
	update domainmonitor.UpdateMonitor,
) (domainmonitor.Monitor, error) {
	return f(ctx, id, update)
}

func updateMonitorTestLogger() *slog.Logger {
	return slog.New(
		slog.NewJSONHandler(io.Discard, nil),
	)
}

func TestUpdateMonitorHandler(t *testing.T) {
	id := uuid.New()

	now := time.Date(
		2026,
		time.July,
		26,
		16,
		0,
		0,
		0,
		time.UTC,
	)

	server := NewServer(ServerConfig{
		Logger: updateMonitorTestLogger(),
		MonitorUpdater: monitorUpdaterFunc(func(
			_ context.Context,
			gotID uuid.UUID,
			update domainmonitor.UpdateMonitor,
		) (domainmonitor.Monitor, error) {
			if gotID != id {
				t.Fatalf(
					"monitor ID = %s, want %s",
					gotID,
					id,
				)
			}

			if update.Name == nil {
				t.Fatal("update Name = nil, want non-nil")
			}

			if *update.Name != "Updated monitor" {
				t.Fatalf(
					"update Name = %q, want %q",
					*update.Name,
					"Updated monitor",
				)
			}

			if update.Enabled == nil {
				t.Fatal("update Enabled = nil, want non-nil")
			}

			if *update.Enabled {
				t.Fatal("update Enabled = true, want false")
			}

			return domainmonitor.Monitor{
				ID:                 id,
				Name:               *update.Name,
				URL:                "https://example.com/health",
				Method:             "GET",
				Interval:           30 * time.Second,
				Timeout:            1500 * time.Millisecond,
				ExpectedStatusFrom: 200,
				ExpectedStatusTo:   299,
				Enabled:            *update.Enabled,
				NextCheckAt:        now.Add(time.Minute),
				CreatedAt:          now.Add(-time.Hour),
				UpdatedAt:          now,
			}, nil
		}),
	})

	request := httptest.NewRequest(
		http.MethodPatch,
		"/api/v1/monitors/"+id.String(),
		strings.NewReader(`{
			"name": "Updated monitor",
			"enabled": false
		}`),
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

	if payload.Name != "Updated monitor" {
		t.Fatalf(
			"Name = %q, want %q",
			payload.Name,
			"Updated monitor",
		)
	}

	if payload.Enabled {
		t.Fatal("Enabled = true, want false")
	}
}

func TestUpdateMonitorHandlerInvalidID(t *testing.T) {
	server := NewServer(ServerConfig{
		Logger: updateMonitorTestLogger(),
		MonitorUpdater: monitorUpdaterFunc(func(
			_ context.Context,
			_ uuid.UUID,
			_ domainmonitor.UpdateMonitor,
		) (domainmonitor.Monitor, error) {
			t.Fatal("updater must not be called")

			return domainmonitor.Monitor{}, nil
		}),
	})

	request := httptest.NewRequest(
		http.MethodPatch,
		"/api/v1/monitors/not-a-uuid",
		strings.NewReader(`{"name":"Updated"}`),
	)

	response := httptest.NewRecorder()
	server.Handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf(
			"status code = %d, want %d",
			response.Code,
			http.StatusBadRequest,
		)
	}
}

func TestUpdateMonitorHandlerInvalidBody(t *testing.T) {
	id := uuid.New()

	server := NewServer(ServerConfig{
		Logger: updateMonitorTestLogger(),
		MonitorUpdater: monitorUpdaterFunc(func(
			_ context.Context,
			_ uuid.UUID,
			_ domainmonitor.UpdateMonitor,
		) (domainmonitor.Monitor, error) {
			t.Fatal("updater must not be called")

			return domainmonitor.Monitor{}, nil
		}),
	})

	request := httptest.NewRequest(
		http.MethodPatch,
		"/api/v1/monitors/"+id.String(),
		strings.NewReader(`{"banana":true}`),
	)

	response := httptest.NewRecorder()
	server.Handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf(
			"status code = %d, want %d",
			response.Code,
			http.StatusBadRequest,
		)
	}
}

func TestUpdateMonitorHandlerValidationError(t *testing.T) {
	id := uuid.New()

	server := NewServer(ServerConfig{
		Logger: updateMonitorTestLogger(),
		MonitorUpdater: monitorUpdaterFunc(func(
			_ context.Context,
			_ uuid.UUID,
			_ domainmonitor.UpdateMonitor,
		) (domainmonitor.Monitor, error) {
			return domainmonitor.Monitor{}, fmt.Errorf(
				"validate monitor update: %w",
				&domainmonitor.ValidationError{
					Field:   "update",
					Message: "must contain at least one field",
				},
			)
		}),
	})

	request := httptest.NewRequest(
		http.MethodPatch,
		"/api/v1/monitors/"+id.String(),
		strings.NewReader(`{}`),
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
			Code string `json:"code"`
		} `json:"error"`
	}

	if err := json.NewDecoder(result.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if payload.Error.Code != "invalid_monitor" {
		t.Fatalf(
			"error code = %q, want %q",
			payload.Error.Code,
			"invalid_monitor",
		)
	}
}

func TestUpdateMonitorHandlerNotFound(t *testing.T) {
	id := uuid.New()

	server := NewServer(ServerConfig{
		Logger: updateMonitorTestLogger(),
		MonitorUpdater: monitorUpdaterFunc(func(
			_ context.Context,
			_ uuid.UUID,
			_ domainmonitor.UpdateMonitor,
		) (domainmonitor.Monitor, error) {
			return domainmonitor.Monitor{},
				repositorypostgres.ErrMonitorNotFound
		}),
	})

	request := httptest.NewRequest(
		http.MethodPatch,
		"/api/v1/monitors/"+id.String(),
		strings.NewReader(`{"name":"Updated"}`),
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
