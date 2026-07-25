package monitor

import (
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNormalizeAndValidate(t *testing.T) {
	now := time.Date(
		2026,
		time.July,
		8,
		12,
		0,
		0,
		0,
		time.FixedZone("test", 3*60*60),
	)

	validInput := func() NewMonitor {
		return NewMonitor{
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

	tests := []struct {
		name      string
		mutate    func(*NewMonitor)
		wantField string
	}{
		{
			name: "blank name",
			mutate: func(input *NewMonitor) {
				input.Name = "   "
			},
			wantField: "name",
		},
		{
			name: "name too long",
			mutate: func(input *NewMonitor) {
				input.Name = strings.Repeat("a", MaxNameLength+1)
			},
			wantField: "name",
		},
		{
			name: "invalid URL",
			mutate: func(input *NewMonitor) {
				input.URL = "not a URL"
			},
			wantField: "url",
		},
		{
			name: "unsupported URL scheme",
			mutate: func(input *NewMonitor) {
				input.URL = "ftp://example.com/file"
			},
			wantField: "url",
		},
		{
			name: "URL without host",
			mutate: func(input *NewMonitor) {
				input.URL = "https:///health"
			},
			wantField: "url",
		},
		{
			name: "URL with credentials",
			mutate: func(input *NewMonitor) {
				input.URL = "https://user:secret@example.com/health"
			},
			wantField: "url",
		},
		{
			name: "URL with fragment",
			mutate: func(input *NewMonitor) {
				input.URL = "https://example.com/health#status"
			},
			wantField: "url",
		},
		{
			name: "unsupported method",
			mutate: func(input *NewMonitor) {
				input.Method = http.MethodPost
			},
			wantField: "method",
		},
		{
			name: "interval below minimum",
			mutate: func(input *NewMonitor) {
				input.Interval = MinInterval - time.Second
			},
			wantField: "interval",
		},
		{
			name: "interval above maximum",
			mutate: func(input *NewMonitor) {
				input.Interval = MaxInterval + time.Second
			},
			wantField: "interval",
		},
		{
			name: "timeout below minimum",
			mutate: func(input *NewMonitor) {
				input.Timeout = MinTimeout - time.Millisecond
			},
			wantField: "timeout",
		},
		{
			name: "timeout above maximum",
			mutate: func(input *NewMonitor) {
				input.Timeout = MaxTimeout + time.Second
			},
			wantField: "timeout",
		},
		{
			name: "timeout equal to interval",
			mutate: func(input *NewMonitor) {
				input.Interval = 10 * time.Second
				input.Timeout = 10 * time.Second
			},
			wantField: "timeout",
		},
		{
			name: "invalid expected status from",
			mutate: func(input *NewMonitor) {
				input.ExpectedStatusFrom = 99
			},
			wantField: "expected_status_from",
		},
		{
			name: "invalid expected status to",
			mutate: func(input *NewMonitor) {
				input.ExpectedStatusTo = 600
			},
			wantField: "expected_status_to",
		},
		{
			name: "reversed expected status range",
			mutate: func(input *NewMonitor) {
				input.ExpectedStatusFrom = 300
				input.ExpectedStatusTo = 200
			},
			wantField: "expected_status_to",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := validInput()
			tt.mutate(&input)

			_, err := NormalizeAndValidate(input, now)
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if !errors.Is(err, ErrInvalidMonitor) {
				t.Fatalf(
					"error = %v, want ErrInvalidMonitor",
					err,
				)
			}

			var validationErr *ValidationError
			if !errors.As(err, &validationErr) {
				t.Fatalf(
					"error type = %T, want *ValidationError",
					err,
				)
			}

			if validationErr.Field != tt.wantField {
				t.Fatalf(
					"field = %q, want %q",
					validationErr.Field,
					tt.wantField,
				)
			}
		})
	}
}

func TestNormalizeAndValidateAppliesDefaults(t *testing.T) {
	now := time.Date(
		2026,
		time.July,
		8,
		9,
		0,
		0,
		0,
		time.FixedZone("test", 3*60*60),
	)

	result, err := NormalizeAndValidate(
		NewMonitor{
			Name:     " Example API ",
			URL:      " https://example.com/health ",
			Interval: 30 * time.Second,
			Timeout:  2 * time.Second,
			Enabled:  true,
		},
		now,
	)
	if err != nil {
		t.Fatalf("NormalizeAndValidate() error: %v", err)
	}

	if result.Name != "Example API" {
		t.Fatalf(
			"name = %q, want %q",
			result.Name,
			"Example API",
		)
	}

	if result.URL != "https://example.com/health" {
		t.Fatalf(
			"URL = %q, want %q",
			result.URL,
			"https://example.com/health",
		)
	}

	if result.Method != http.MethodGet {
		t.Fatalf(
			"method = %q, want GET",
			result.Method,
		)
	}

	if result.ExpectedStatusFrom != http.StatusOK {
		t.Fatalf(
			"expected status from = %d, want %d",
			result.ExpectedStatusFrom,
			http.StatusOK,
		)
	}

	if result.ExpectedStatusTo != 299 {
		t.Fatalf(
			"expected status to = %d, want 299",
			result.ExpectedStatusTo,
		)
	}

	if !result.NextCheckAt.Equal(now.UTC()) {
		t.Fatalf(
			"next check at = %s, want %s",
			result.NextCheckAt,
			now.UTC(),
		)
	}

	if result.NextCheckAt.Location() != time.UTC {
		t.Fatalf(
			"next check location = %s, want UTC",
			result.NextCheckAt.Location(),
		)
	}
}

func TestNormalizeAndValidateUpdate(t *testing.T) {
	now := time.Date(
		2026,
		time.July,
		25,
		13,
		0,
		0,
		0,
		time.UTC,
	)

	current := Monitor{
		ID:                 uuid.New(),
		Name:               "Old API",
		URL:                "https://example.com/health",
		Method:             "GET",
		Interval:           30 * time.Second,
		Timeout:            1500 * time.Millisecond,
		ExpectedStatusFrom: 200,
		ExpectedStatusTo:   299,
		Enabled:            true,
		NextCheckAt:        now.Add(time.Minute),
		CreatedAt:          now.Add(-time.Hour),
		UpdatedAt:          now.Add(-time.Hour),
	}

	t.Run("applies and normalizes update", func(t *testing.T) {
		name := "  Updated API  "
		method := " get "
		enabled := false

		result, err := NormalizeAndValidateUpdate(
			current,
			UpdateMonitor{
				Name:    &name,
				Method:  &method,
				Enabled: &enabled,
			},
			now,
		)
		if err != nil {
			t.Fatalf("NormalizeAndValidateUpdate() error = %v", err)
		}

		if result.Name != "Updated API" {
			t.Fatalf(
				"Name = %q, want %q",
				result.Name,
				"Updated API",
			)
		}

		if result.Method != "GET" {
			t.Fatalf(
				"Method = %q, want %q",
				result.Method,
				"GET",
			)
		}

		if result.Enabled {
			t.Fatal("Enabled = true, want false")
		}

		if result.ID != current.ID {
			t.Fatalf(
				"ID = %s, want %s",
				result.ID,
				current.ID,
			)
		}

		if !result.UpdatedAt.Equal(now) {
			t.Fatalf(
				"UpdatedAt = %s, want %s",
				result.UpdatedAt,
				now,
			)
		}
	})

	t.Run("rejects empty update", func(t *testing.T) {
		_, err := NormalizeAndValidateUpdate(
			current,
			UpdateMonitor{},
			now,
		)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !errors.Is(err, ErrInvalidMonitor) {
			t.Fatalf(
				"error = %v, want ErrInvalidMonitor",
				err,
			)
		}
	})

	t.Run("rejects invalid resulting monitor", func(t *testing.T) {
		timeout := 30 * time.Second

		_, err := NormalizeAndValidateUpdate(
			current,
			UpdateMonitor{
				Timeout: &timeout,
			},
			now,
		)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !errors.Is(err, ErrInvalidMonitor) {
			t.Fatalf(
				"error = %v, want ErrInvalidMonitor",
				err,
			)
		}
	})
}
