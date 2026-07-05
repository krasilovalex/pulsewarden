package middleware

import (
	"errors"
	"log/slog"
	"net/http"
	"runtime/debug"
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

			http.Error(
				tracker,
				http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError,
			)
		}()

		next.ServeHTTP(tracker, r)
	})
}
