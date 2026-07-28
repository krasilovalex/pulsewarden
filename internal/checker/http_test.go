package checker

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/wayzzoo/pulsewarden/internal/domain/checkresult"
	"github.com/wayzzoo/pulsewarden/internal/domain/monitor"
)

type roundTripperFunc func(
	request *http.Request,
) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(
	request *http.Request,
) (*http.Response, error) {
	return f(request)
}

func TestHTTPCheckerExpectedStatus(t *testing.T) {
	monitorID := uuid.New()

	startedAt := time.Date(
		2026,
		time.July,
		28,
		12,
		0,
		0,
		0,
		time.UTC,
	)

	finishedAt := startedAt.Add(125 * time.Millisecond)

	client := &http.Client{
		Transport: roundTripperFunc(func(
			request *http.Request,
		) (*http.Response, error) {
			if request.Method != http.MethodGet {
				t.Fatalf(
					"request method = %q, want GET",
					request.Method,
				)
			}

			return &http.Response{
				StatusCode: http.StatusNoContent,
				Header:     make(http.Header),
				Body: io.NopCloser(
					strings.NewReader(""),
				),
			}, nil
		}),
	}

	httpChecker := NewHTTPChecker(client)
	httpChecker.now = sequentialClock(
		startedAt,
		finishedAt,
	)

	result, err := httpChecker.Check(
		context.Background(),
		testMonitor(monitorID),
	)
	if err != nil {
		t.Fatalf("Check() error: %v", err)
	}

	if result.Status != checkresult.StatusUp {
		t.Fatalf(
			"status = %q, want %q",
			result.Status,
			checkresult.StatusUp,
		)
	}

	if result.StatusCode == nil {
		t.Fatal("status code is nil")
	}

	if *result.StatusCode != http.StatusNoContent {
		t.Fatalf(
			"status code = %d, want %d",
			*result.StatusCode,
			http.StatusNoContent,
		)
	}

	if result.Error != nil {
		t.Fatalf(
			"error message = %q, want nil",
			*result.Error,
		)
	}

	if result.Latency != 125*time.Millisecond {
		t.Fatalf(
			"latency = %s, want %s",
			result.Latency,
			125*time.Millisecond,
		)
	}

	if !result.CheckedAt.Equal(startedAt) {
		t.Fatalf(
			"checked at = %s, want %s",
			result.CheckedAt,
			startedAt,
		)
	}
}

func TestHTTPCheckerUnexpectedStatus(t *testing.T) {
	monitorID := uuid.New()

	startedAt := time.Date(
		2026,
		time.July,
		28,
		12,
		0,
		0,
		0,
		time.UTC,
	)

	client := &http.Client{
		Transport: roundTripperFunc(func(
			*http.Request,
		) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusServiceUnavailable,
				Header:     make(http.Header),
				Body: io.NopCloser(
					strings.NewReader("unavailable"),
				),
			}, nil
		}),
	}

	httpChecker := NewHTTPChecker(client)
	httpChecker.now = sequentialClock(
		startedAt,
		startedAt.Add(50*time.Millisecond),
	)

	result, err := httpChecker.Check(
		context.Background(),
		testMonitor(monitorID),
	)
	if err != nil {
		t.Fatalf("Check() error: %v", err)
	}

	if result.Status != checkresult.StatusDown {
		t.Fatalf(
			"status = %q, want %q",
			result.Status,
			checkresult.StatusDown,
		)
	}

	if result.StatusCode == nil {
		t.Fatal("status code is nil")
	}

	if *result.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf(
			"status code = %d, want %d",
			*result.StatusCode,
			http.StatusServiceUnavailable,
		)
	}

	if result.Error != nil {
		t.Fatalf(
			"error message = %q, want nil",
			*result.Error,
		)
	}
}

func TestHTTPCheckerRequestFailure(t *testing.T) {
	monitorID := uuid.New()

	startedAt := time.Date(
		2026,
		time.July,
		28,
		12,
		0,
		0,
		0,
		time.UTC,
	)

	client := &http.Client{
		Transport: roundTripperFunc(func(
			*http.Request,
		) (*http.Response, error) {
			return nil, errors.New("connection refused")
		}),
	}

	httpChecker := NewHTTPChecker(client)
	httpChecker.now = sequentialClock(
		startedAt,
		startedAt.Add(25*time.Millisecond),
	)

	result, err := httpChecker.Check(
		context.Background(),
		testMonitor(monitorID),
	)
	if err != nil {
		t.Fatalf("Check() error: %v", err)
	}

	if result.Status != checkresult.StatusDown {
		t.Fatalf(
			"status = %q, want %q",
			result.Status,
			checkresult.StatusDown,
		)
	}

	if result.StatusCode != nil {
		t.Fatalf(
			"status code = %d, want nil",
			*result.StatusCode,
		)
	}

	if result.Error == nil {
		t.Fatal("error message is nil")
	}

	if !strings.Contains(
		*result.Error,
		"connection refused",
	) {
		t.Fatalf(
			"error message = %q, want connection refused",
			*result.Error,
		)
	}

	if result.Latency != 25*time.Millisecond {
		t.Fatalf(
			"latency = %s, want %s",
			result.Latency,
			25*time.Millisecond,
		)
	}
}

func testMonitor(id uuid.UUID) monitor.Monitor {
	return monitor.Monitor{
		ID:                 id,
		Name:               "Example API",
		URL:                "https://example.com/health",
		Method:             http.MethodGet,
		Interval:           30 * time.Second,
		Timeout:            2 * time.Second,
		ExpectedStatusFrom: 200,
		ExpectedStatusTo:   299,
		Enabled:            true,
	}
}

func sequentialClock(
	values ...time.Time,
) func() time.Time {
	index := 0

	return func() time.Time {
		if index >= len(values) {
			panic("sequential clock exhausted")
		}

		value := values[index]
		index++

		return value
	}
}
