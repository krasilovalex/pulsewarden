package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/krasilovalex/pulsewarden/internal/app/api/middleware"
	"github.com/krasilovalex/pulsewarden/internal/app/api/response"
	domainmonitor "github.com/krasilovalex/pulsewarden/internal/domain/monitor"
)

type createMonitorResponse struct {
	ID                  string `json:"id"`
	Name                string `json:"name"`
	URL                 string `json:"url"`
	Method              string `json:"method"`
	IntervalSeconds     int64  `json:"interval_seconds"`
	TimeoutMilliseconds int64  `json:"timeout_milliseconds"`
	ExpectedStatusFrom  int    `json:"expected_status_from"`
	ExpectedStatusTo    int    `json:"expected_status_to"`
	Enabled             bool   `json:"enabled"`
	NextCheckAt         string `json:"next_check_at"`
	CreatedAt           string `json:"created_at"`
	UpdatedAt           string `json:"updated_at"`
}

func createMonitorHandler(
	logger *slog.Logger,
	creator MonitorCreator,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID, _ := middleware.RequestIDFromContext(r.Context())
		var payload createMonitorRequest

		if err := decodeJSON(w, r, &payload); err != nil {
			_ = response.WriteError(w, http.StatusBadRequest, "invalid_request", "request_body is invalid", requestID)
			return
		}

		created, err := creator.Execute(
			r.Context(),
			payload.toDomain(),
		)
		if err != nil {
			if errors.Is(err, domainmonitor.ErrInvalidMonitor) {
				_ = response.WriteError(
					w,
					http.StatusBadRequest,
					"invalid_monitor",
					err.Error(),
					requestID,
				)
				return
			}

			logger.Error(
				"create monitor failed",
				slog.Any("error", err),
				slog.String("request_id", requestID),
			)

			_ = response.WriteError(
				w,
				http.StatusInternalServerError,
				"internal_error",
				"internal server error",
				requestID,
			)
			return
		}

		writeMonitorCreated(w, created)
	}
}

func writeMonitorCreated(w http.ResponseWriter, created domainmonitor.Monitor) {

	response := createMonitorResponse{
		ID:                  created.ID.String(),
		Name:                created.Name,
		URL:                 created.URL,
		Method:              created.Method,
		IntervalSeconds:     int64(created.Interval.Seconds()),
		TimeoutMilliseconds: created.Timeout.Milliseconds(),
		ExpectedStatusFrom:  created.ExpectedStatusFrom,
		ExpectedStatusTo:    created.ExpectedStatusTo,
		Enabled:             created.Enabled,
		NextCheckAt:         created.NextCheckAt.UTC().Format(time.RFC3339Nano),
		CreatedAt:           created.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt:           created.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	_ = json.NewEncoder(w).Encode(response)
}
