package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestIDUsesValidIncomingHeader(t *testing.T) {
	const incomingID = "client-request-123"

	var contextID string

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var ok bool

		contextID, ok = RequestIDFromContext(r.Context())
		if !ok {
			t.Fatal("request ID is missing from context")
		}

		w.WriteHeader(http.StatusNoContent)
	})

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set(RequestIDHeader, incomingID)

	response := httptest.NewRecorder()

	RequestID(next).ServeHTTP(response, request)

	if contextID != incomingID {
		t.Fatalf("context request ID = %q, want %q", contextID, incomingID)
	}

	if got := response.Header().Get(RequestIDHeader); got != incomingID {
		t.Fatalf("response request ID = %q, want %q", got, incomingID)
	}

	if response.Code != http.StatusNoContent {
		t.Fatalf(
			"status code = %d, want %d",
			response.Code,
			http.StatusNoContent,
		)
	}
}

func TestRequestIDGeneratesIDWhenHeaderIsMissing(t *testing.T) {
	const generatedID = "generated-request-id"

	generator := func() (string, error) {
		return generatedID, nil
	}

	var contextID string

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contextID, _ = RequestIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()

	requestID(next, generator).ServeHTTP(response, request)

	if contextID != generatedID {
		t.Fatalf("context request ID = %q, want %q", contextID, generatedID)
	}

	if got := response.Header().Get(RequestIDHeader); got != generatedID {
		t.Fatalf("response request ID = %q, want %q", got, generatedID)
	}
}

func TestRequestIDReplacesInvalidIncomingHeader(t *testing.T) {
	const generatedID = "safe-generated-id"

	generator := func() (string, error) {
		return generatedID, nil
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set(RequestIDHeader, "unsafe\nrequest-id")

	response := httptest.NewRecorder()

	requestID(next, generator).ServeHTTP(response, request)

	if got := response.Header().Get(RequestIDHeader); got != generatedID {
		t.Fatalf("response request ID = %q, want %q", got, generatedID)
	}
}

func TestRequestIDReturnsInternalServerErrorWhenGenerationFails(t *testing.T) {
	generator := func() (string, error) {
		return "", errors.New("random source unavailable")
	}

	nextCalled := false

	next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		nextCalled = true
	})

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()

	requestID(next, generator).ServeHTTP(response, request)

	if nextCalled {
		t.Fatal("next handler was called")
	}

	if response.Code != http.StatusInternalServerError {
		t.Fatalf(
			"status code = %d, want %d",
			response.Code,
			http.StatusInternalServerError,
		)
	}
}

func TestValidRequestID(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{
			name:  "letters digits and separators",
			value: "Request-123_test.value",
			want:  true,
		},
		{
			name:  "empty",
			value: "",
			want:  false,
		},
		{
			name:  "contains whitespace",
			value: "request 123",
			want:  false,
		},
		{
			name:  "contains slash",
			value: "request/123",
			want:  false,
		},
		{
			name:  "too long",
			value: string(make([]byte, maxRequestIDLength+1)),
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validRequestID(tt.value)

			if got != tt.want {
				t.Fatalf(
					"validRequestID(%q) = %v, want %v",
					tt.value,
					got,
					tt.want,
				)
			}
		})
	}
}
