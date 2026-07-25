package monitor

import (
	"testing"
	"time"
)

func TestUpdateMonitorIsEmpty(t *testing.T) {
	update := UpdateMonitor{}

	if !update.IsEmpty() {
		t.Fatal("IsEmpty() = false, want true")
	}
}

func TestUpdateMonitorIsNotEmpty(t *testing.T) {
	enabled := false

	update := UpdateMonitor{
		Enabled: &enabled,
	}

	if update.IsEmpty() {
		t.Fatal("IsEmpty() = true, want false")
	}
}

func TestUpdateMonitorApply(t *testing.T) {
	name := "Updated API"
	interval := 45 * time.Second
	enabled := false

	target := Monitor{
		Name:               "Old API",
		URL:                "https://example.com/health",
		Method:             "GET",
		Interval:           30 * time.Second,
		Timeout:            1500 * time.Millisecond,
		ExpectedStatusFrom: 200,
		ExpectedStatusTo:   299,
		Enabled:            true,
	}

	update := UpdateMonitor{
		Name:     &name,
		Interval: &interval,
		Enabled:  &enabled,
	}

	update.Apply(&target)

	if target.Name != name {
		t.Fatalf(
			"Name = %q, want %q",
			target.Name,
			name,
		)
	}

	if target.Interval != interval {
		t.Fatalf(
			"Interval = %s, want %s",
			target.Interval,
			interval,
		)
	}

	if target.Enabled {
		t.Fatalf("Enabled = true, want false")
	}

	if target.URL != "https://example.com/health" {
		t.Fatalf(
			"URL = %q, want unchanged value",
			target.URL,
		)
	}

	if target.Timeout != 1500*time.Millisecond {
		t.Fatalf(
			"Timeout = %s, want unchanged value",
			target.Timeout,
		)
	}
}
