package api

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/wayzzoo/pulsewarden/internal/app/api/middleware"
	"github.com/wayzzoo/pulsewarden/internal/app/api/response"
	domainmonitor "github.com/wayzzoo/pulsewarden/internal/domain/monitor"
	repositorypostgres "github.com/wayzzoo/pulsewarden/internal/repository/postgres"
)

type MonitorUpdater interface {
	Execute(
		ctx context.Context,
		id uuid.UUID,
		update domainmonitor.UpdateMonitor,
	) (domainmonitor.Monitor, error)
}

func updateMonitorHandler(
	logger *slog.Logger,
	updater MonitorUpdater,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID, _ := middleware.RequestIDFromContext(r.Context())

		id, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			_ = response.WriteError(
				w,
				http.StatusBadRequest,
				"invalid monitor id",
				"monitor id must be a valid UUID",
				requestID,
			)
			return
		}

		var request updateMonitorRequest

		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()

		if err := decoder.Decode(&request); err != nil {
			_ = response.WriteError(
				w,
				http.StatusBadRequest,
				"invalid_request_body",
				"request body must be valid JSON",
				requestID,
			)
			return
		}

		result, err := updater.Execute(
			r.Context(),
			id,
			request.toDomain(),
		)
		if err != nil {
			switch {
			case errors.Is(err, repositorypostgres.ErrMonitorNotFound):
				_ = response.WriteError(
					w,
					http.StatusNotFound,
					"monitor_not_found",
					"monitor not found",
					requestID,
				)

			case errors.Is(err, domainmonitor.ErrInvalidMonitor):
				_ = response.WriteError(

					w,
					http.StatusBadRequest,
					"invalid_monitor",
					err.Error(),
					requestID,
				)
			default:
				logger.Error(
					"update monitor failed",
					slog.Any("error", err),
					slog.String("monitor_id", id.String()),
					slog.String("request_id", requestID),
				)

				_ = response.WriteError(
					w,
					http.StatusInternalServerError,
					"internal_error",
					"internal server error",
					requestID,
				)
			}

			return
		}

		_ = response.WriteJSON(
			w,
			http.StatusOK,
			monitorToResponse(result),
		)
	}
}
