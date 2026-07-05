package middleware

import (
	"errors"
	"log/slog"
	"net/http"
	"runtime/debug"
)

type responseTracker struct {
	http.ResponseWriter
	wroteHeader bool
}

func (w *responseTracker) WriteHeader(statusCode int) {
	if w.wroteHeader {
		return
	}

	w.wroteHeader = true
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *responseTracker) Write(data []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}

	return w.ResponseWriter.Write(data)
}

func Recovery(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tracker := &responseTracker{
			ResponseWriter: w,
		}

		defer func() {
			recovered := recover()
			if recovered == nil {
				return
			}

			if err, ok := recovered.(error); ok && errors.Is(err, http.ErrAbortHandler) {
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
