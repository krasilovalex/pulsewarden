package middleware

import (
	"errors"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/krasilovalex/pulsewarden/internal/app/api/response"
)

func Recovery(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tracker := trackResponse(w)

		defer func() {
			recovered := recover()
			if recovered == nil {
				return
			}

			if err, ok := recovered.(error); ok &&
				errors.Is(err, http.ErrAbortHandler) {
				panic(http.ErrAbortHandler)
			}

			requestID, _ := RequestIDFromContext(r.Context())

			log.ErrorContext(
				r.Context(),
				"panic recovered",
				slog.Any("panic", recovered),
				slog.String("request_id", requestID),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("stack", string(debug.Stack())),
			)

			if tracker.wroteHeader {
				return
			}

			if err := response.WriteError(
				tracker,
				http.StatusInternalServerError,
				"internal_error",
				"internal server error",
				requestID,
			); err != nil {
				log.ErrorContext(
					r.Context(),
					"write panic response",
					slog.Any("error", err),
					slog.String("request_id", requestID),
				)
			}
		}()

		next.ServeHTTP(tracker, r)
	})
}
