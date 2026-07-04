package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
)

const (
	RequestIDHeader = "X-Request-ID"

	maxRequestIDLength = 128
	randomIDBytes      = 16
)

type requestIDContextKey struct{}

type requestIDGenerator func() (string, error)

func RequestID(next http.Handler) http.Handler {
	return requestID(next, generateRequestID)
}

func RequestIDFromContext(ctx context.Context) (string, bool) {
	requestID, ok := ctx.Value(requestIDContextKey{}).(string)
	return requestID, ok
}

func requestID(
	next http.Handler,
	generate requestIDGenerator,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(RequestIDHeader)

		if !validRequestID(id) {
			var err error

			id, err = generate()
			if err != nil {
				http.Error(
					w,
					http.StatusText(http.StatusInternalServerError),
					http.StatusInternalServerError,
				)
				return
			}
		}

		w.Header().Set(RequestIDHeader, id)

		ctx := context.WithValue(
			r.Context(),
			requestIDContextKey{},
			id,
		)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func generateRequestID() (string, error) {
	randomBytes := make([]byte, randomIDBytes)

	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("generate request ID: %w", err)
	}

	return hex.EncodeToString(randomBytes), nil
}

func validRequestID(value string) bool {
	if value == "" || len(value) > maxRequestIDLength {
		return false
	}

	for _, char := range value {
		switch {
		case char >= 'a' && char <= 'z':
		case char >= 'A' && char <= 'Z':
		case char >= '0' && char <= '9':
		case char == '.', char == '_', char == '-':
		default:
			return false
		}
	}

	return true
}
