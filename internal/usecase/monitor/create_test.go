package monitor

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"

	domainmonitor "github.com/wayzzoo/pulsewarden/internal/domain/monitor"
)

type creatorRepositoryFunc func(
	context.Context,
	domainmonitor.NewMonitor,
) (domainmonitor.Monitor, error)

func (f creatorRepositoryFunc) Create(
	ctx context.Context,
	input domainmonitor.NewMonitor,
) (domainmonitor.Monitor, error) {
	return f(ctx, input)
}

func TestCreateExecute(t *testing.T) {
	now := time.Date(
		2026,
		time.July,
		8,
		12,
		0,
		0,
		0,
		time.UTC,
	)

	expectedID := uuid.New()

	var repositoryInput domainmonitor.NewMonitor

	useCase := NewCreate(
		creatorRepositoryFunc(func(
			_ context.Context,
			input domainmonitor.NewMonitor,
		) (domainmonitor.Monitor, error) {
			repositoryInput = input

			return domainmonitor.Monitor{
				ID:                 expectedID,
				Name:               input.Name,
				URL:                input.URL,
				Method:             input.Method,
				Interval:           input.Interval,
				Timeout:            input.Timeout,
				ExpectedStatusFrom: input.ExpectedStatusFrom,
				ExpectedStatusTo:   input.ExpectedStatusTo,
				Enabled:            input.Enabled,
				NextCheckAt:        input.NextCheckAt,
				CreatedAt:          now,
				UpdatedAt:          now,
			}, nil
		}),
	)

	useCase.now = func() time.Time {
		return now
	}

	result, err := useCase.Execute(
		context.Background(),
		domainmonitor.NewMonitor{
			Name:     " Example API ",
			URL:      " https://example.com/health ",
			Interval: 30 * time.Second,
			Timeout:  2 * time.Second,
			Enabled:  true,
		},
	)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	if result.ID != expectedID {
		t.Fatalf(
			"ID = %s, want %s",
			result.ID,
			expectedID,
		)
	}

	if repositoryInput.Name != "Example API" {
		t.Fatalf(
			"repository name = %q, want %q",
			repositoryInput.Name,
			"Example API",
		)
	}

	if repositoryInput.Method != http.MethodGet {
		t.Fatalf(
			"repository method = %q, want GET",
			repositoryInput.Method,
		)
	}

	if !repositoryInput.NextCheckAt.Equal(now) {
		t.Fatalf(
			"repository next check at = %s, want %s",
			repositoryInput.NextCheckAt,
			now,
		)
	}
}

func TestCreateExecuteRejectsInvalidInput(t *testing.T) {
	repositoryCalled := false

	useCase := NewCreate(
		creatorRepositoryFunc(func(
			context.Context,
			domainmonitor.NewMonitor,
		) (domainmonitor.Monitor, error) {
			repositoryCalled = true
			return domainmonitor.Monitor{}, nil
		}),
	)

	_, err := useCase.Execute(
		context.Background(),
		domainmonitor.NewMonitor{
			Name:     "",
			URL:      "https://example.com",
			Interval: 30 * time.Second,
			Timeout:  time.Second,
		},
	)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, domainmonitor.ErrInvalidMonitor) {
		t.Fatalf(
			"error = %v, want ErrInvalidMonitor",
			err,
		)
	}

	if repositoryCalled {
		t.Fatal("repository was called for invalid input")
	}
}

func TestCreateExecuteWrapsRepositoryError(t *testing.T) {
	repositoryErr := errors.New("repository failure")

	useCase := NewCreate(
		creatorRepositoryFunc(func(
			context.Context,
			domainmonitor.NewMonitor,
		) (domainmonitor.Monitor, error) {
			return domainmonitor.Monitor{}, repositoryErr
		}),
	)

	_, err := useCase.Execute(
		context.Background(),
		domainmonitor.NewMonitor{
			Name:     "Example API",
			URL:      "https://example.com",
			Interval: 30 * time.Second,
			Timeout:  time.Second,
			Enabled:  true,
		},
	)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, repositoryErr) {
		t.Fatalf(
			"error = %v, want wrapped repository error",
			err,
		)
	}
}
