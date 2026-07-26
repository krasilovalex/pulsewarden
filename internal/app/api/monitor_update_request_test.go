package api

import (
	"testing"
	"time"
)

func TestUpdateMonitorRequestToDomain(t *testing.T) {
	name := "Updated API"
	url := "https://example.com/health"
	method := "get"
	intervalSeconds := int64(45)
	timeoutMilliseconds := int64(1500)
	expectedStatusFrom := 200
	expectedStatusTo := 204
	enabled := false

	request := updateMonitorRequest{
		Name:                &name,
		URL:                 &url,
		Method:              &method,
		IntervalSeconds:     &intervalSeconds,
		TimeoutMilliseconds: &timeoutMilliseconds,
		ExpectedStatusFrom:  &expectedStatusFrom,
		ExpectedStatusTo:    &expectedStatusTo,
		Enabled:             &enabled,
	}

	result := request.toDomain()

	if result.Name == nil || *result.Name != name {
		t.Fatalf(
			"Name = %v, want %q",
			result.Name,
			name,
		)
	}

	if result.URL == nil || *result.URL != url {
		t.Fatalf(
			"URL = %v, want %q",
			result.URL,
			url,
		)
	}

	if result.Method == nil || *result.Method != method {
		t.Fatalf(
			"Method = %v, want %q",
			result.Method,
			method,
		)
	}

	if result.Interval == nil {
		t.Fatalf("Interval = nil, want non-nil")
	}

	if *result.Interval != 45*time.Second {
		t.Fatalf(
			"Interval = %s, want %s",
			*result.Interval,
			45*time.Second,
		)
	}

	if result.Timeout == nil {
		t.Fatalf("Timeout = nil, want non-nil")
	}

	if *result.Timeout != 1500*time.Millisecond {
		t.Fatalf(
			"Timeout = %s, want %s",
			*result.Timeout,
			1500*time.Millisecond,
		)
	}

	if result.ExpectedStatusFrom == nil ||
		*result.ExpectedStatusFrom != expectedStatusFrom {
		t.Fatalf(
			"ExpectedStatusFrom = %v, want %d",
			result.ExpectedStatusFrom,
			expectedStatusFrom,
		)
	}

	if result.ExpectedStatusTo == nil ||
		*result.ExpectedStatusTo != expectedStatusTo {
		t.Fatalf(
			"ExpectedStatusTo = %v, want %d",
			result.ExpectedStatusTo,
			expectedStatusTo,
		)
	}

	if result.Enabled == nil {
		t.Fatal("Enabled = nil, want non-nil")
	}

	if *result.Enabled {
		t.Fatal("Enabled = true, want false")
	}
}

func TestUpdateMonitorRequestToDomainEmpty(t *testing.T) {
	result := (updateMonitorRequest{}).toDomain()

	if !result.IsEmpty() {
		t.Fatal("UpdateMonitor.IsEmpty() = false, want true")
	}
}
