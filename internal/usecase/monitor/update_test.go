package monitor

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	domainmonitor "github.com/wayzzoo/pulsewarden/internal/domain/monitor"
)

type updaterRepositoryStub struct {
	getResult domainmonitor.Monitor
	getErr    error

	updateResult domainmonitor.Monitor
	updateErr    error

	gotGetID      uuid.UUID
	gotUpdateItem domainmonitor.Monitor
	updateCalled  bool
}

func (stub *updaterRepositoryStub) GetByID(
	_ context.Context,
	id uuid.UUID,
) (domainmonitor.Monitor, error) {
	stub.gotGetID = id

	return stub.getResult, stub.getErr
}

func (stub *updaterRepositoryStub) Update(
	_ context.Context,
	input domainmonitor.Monitor,
) (domainmonitor.Monitor, error) {
	stub.updateCalled = true
	stub.gotUpdateItem = input

	if stub.updateErr != nil {
		return domainmonitor.Monitor{}, stub.updateErr
	}

	if stub.updateResult.ID != uuid.Nil {
		return stub.updateResult, nil
	}

	return input, nil
}

func TestUpdateExecute(t *testing.T) {
	id := uuid.New()

	now := time.Date(
		2026,
		time.July,
		26,
		3,
		30,
		0,
		0,
		time.UTC,
	)

	current := domainmonitor.Monitor{
		ID:                 id,
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

	repository := &updaterRepositoryStub{
		getResult: current,
	}

	useCase := NewUpdate(repository)
	useCase.now = func() time.Time {
		return now
	}

	name := "  Updated API  "
	enabled := false

	result, err := useCase.Execute(
		context.Background(),
		id,
		domainmonitor.UpdateMonitor{
			Name:    &name,
			Enabled: &enabled,
		},
	)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if repository.gotGetID != id {
		t.Fatalf(
			"GetByID() ID = %s, want %s",
			repository.gotGetID,
			id,
		)
	}

	if !repository.updateCalled {
		t.Fatal("Update() was not called")
	}

	if repository.gotUpdateItem.Name != "Updated API" {
		t.Fatalf(
			"updated name = %q, want %q",
			repository.gotUpdateItem.Name,
			"Updated API",
		)
	}

	if repository.gotUpdateItem.Enabled {
		t.Fatal("updated Enabled = true, want false")
	}

	if repository.gotUpdateItem.URL != current.URL {
		t.Fatalf(
			"updated URL = %q, want unchanged %q",
			repository.gotUpdateItem.URL,
			current.URL,
		)
	}

	if !repository.gotUpdateItem.UpdatedAt.Equal(now) {
		t.Fatalf(
			"updated_at = %s, want %s",
			repository.gotUpdateItem.UpdatedAt,
			now,
		)
	}

	if result.Name != "Updated API" {
		t.Fatalf(
			"result name = %q, want %q",
			result.Name,
			"Updated API",
		)
	}
}

func TestUpdateExecuteGetError(t *testing.T) {
	expectedErr := errors.New("get failed")

	repository := &updaterRepositoryStub{
		getErr: expectedErr,
	}

	useCase := NewUpdate(repository)

	_, err := useCase.Execute(
		context.Background(),
		uuid.New(),
		domainmonitor.UpdateMonitor{},
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"Execute() error = %v, want %v",
			err,
			expectedErr,
		)
	}

	if repository.updateCalled {
		t.Fatal("Update() was called after GetByID() error")
	}
}

func TestUpdateExecuteValidationError(t *testing.T) {
	repository := &updaterRepositoryStub{
		getResult: domainmonitor.Monitor{
			ID:                 uuid.New(),
			Name:               "Example API",
			URL:                "https://example.com/health",
			Method:             "GET",
			Interval:           30 * time.Second,
			Timeout:            time.Second,
			ExpectedStatusFrom: 200,
			ExpectedStatusTo:   299,
			Enabled:            true,
		},
	}

	useCase := NewUpdate(repository)

	_, err := useCase.Execute(
		context.Background(),
		repository.getResult.ID,
		domainmonitor.UpdateMonitor{},
	)

	if !errors.Is(err, domainmonitor.ErrInvalidMonitor) {
		t.Fatalf(
			"Execute() error = %v, want ErrInvalidMonitor",
			err,
		)
	}

	if repository.updateCalled {
		t.Fatal("Update() was called after validation error")
	}
}

func TestUpdateExecuteRepositoryUpdateError(t *testing.T) {
	expectedErr := errors.New("update failed")

	current := domainmonitor.Monitor{
		ID:                 uuid.New(),
		Name:               "Example API",
		URL:                "https://example.com/health",
		Method:             "GET",
		Interval:           30 * time.Second,
		Timeout:            time.Second,
		ExpectedStatusFrom: 200,
		ExpectedStatusTo:   299,
		Enabled:            true,
	}

	repository := &updaterRepositoryStub{
		getResult: current,
		updateErr: expectedErr,
	}

	useCase := NewUpdate(repository)

	name := "Updated API"

	_, err := useCase.Execute(
		context.Background(),
		current.ID,
		domainmonitor.UpdateMonitor{
			Name: &name,
		},
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"Execute() error = %v, want %v",
			err,
			expectedErr,
		)
	}

	if !repository.updateCalled {
		t.Fatal("Update() was not called")
	}
}
