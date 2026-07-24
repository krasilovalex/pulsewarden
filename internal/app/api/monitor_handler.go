package api

import (
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

func writeMonitorCreated(
	w http.ResponseWriter,
	created domainmonitor.Monitor,
) {
	_ = response.WriteJSON(
		w,
		http.StatusCreated,
		monitorToResponse(created),
	)
}

func monitorToResponse(
	item domainmonitor.Monitor,
) createMonitorResponse {
	return createMonitorResponse{
		ID:                  item.ID.String(),
		Name:                item.Name,
		URL:                 item.URL,
		Method:              item.Method,
		IntervalSeconds:     int64(item.Interval.Seconds()),
		TimeoutMilliseconds: item.Timeout.Milliseconds(),
		ExpectedStatusFrom:  item.ExpectedStatusFrom,
		ExpectedStatusTo:    item.ExpectedStatusTo,
		Enabled:             item.Enabled,
		NextCheckAt:         item.NextCheckAt.UTC().Format(time.RFC3339Nano),
		CreatedAt:           item.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt:           item.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
}
