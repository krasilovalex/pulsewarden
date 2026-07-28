package worker

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	usecasecheck "github.com/wayzzoo/pulsewarden/internal/usecase/check"
)

type dueRunnerFunc func(
	context.Context,
	int,
) (usecasecheck.RunDueResult, error)

func (f dueRunnerFunc) Execute(
	ctx context.Context,
	limit int,
) (usecasecheck.RunDueResult, error) {
	return f(ctx, limit)
}

func TestWorkerRunsImmediatelyAndOnTicks(t *testing.T) {
	calls := make(chan int, 2)

	runner := dueRunnerFunc(func(
		_ context.Context,
		limit int,
	) (usecasecheck.RunDueResult, error) {
		calls <- limit

		return usecasecheck.RunDueResult{}, nil
	})

	application, err := New(Config{
		Logger:       testLogger(),
		Runner:       runner,
		PollInterval: time.Second,
		BatchSize:    25,
	})
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	defer cancel()

	ticks := make(chan time.Time, 1)
	stopped := make(chan struct{})

	go func() {
		defer close(stopped)
		application.run(ctx, ticks)
	}()

	assertRunnerCall(t, calls, 25)

	ticks <- time.Now()

	assertRunnerCall(t, calls, 25)

	cancel()

	select {
	case <-stopped:
	case <-time.After(time.Second):
		t.Fatal("worker did not stop after context cancellation")
	}
}

func TestWorkerContinuesAfterCycleError(t *testing.T) {
	cycleErr := errors.New("database unavailable")

	var (
		mutex sync.Mutex
		calls int
	)

	called := make(chan struct{}, 2)

	runner := dueRunnerFunc(func(
		_ context.Context,
		_ int,
	) (usecasecheck.RunDueResult, error) {
		mutex.Lock()
		calls++
		currentCall := calls
		mutex.Unlock()

		called <- struct{}{}

		if currentCall == 1 {
			return usecasecheck.RunDueResult{}, cycleErr
		}

		return usecasecheck.RunDueResult{
			Claimed:   1,
			Succeeded: 1,
		}, nil
	})

	application, err := New(Config{
		Logger:       testLogger(),
		Runner:       runner,
		PollInterval: time.Second,
		BatchSize:    10,
	})
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	defer cancel()

	ticks := make(chan time.Time, 1)
	stopped := make(chan struct{})

	go func() {
		defer close(stopped)
		application.run(ctx, ticks)
	}()

	assertCycleCompleted(t, called)

	ticks <- time.Now()

	assertCycleCompleted(t, called)

	mutex.Lock()
	gotCalls := calls
	mutex.Unlock()

	if gotCalls != 2 {
		t.Fatalf(
			"runner calls = %d, want 2",
			gotCalls,
		)
	}

	cancel()

	select {
	case <-stopped:
	case <-time.After(time.Second):
		t.Fatal("worker did not stop")
	}
}

func TestNewValidatesConfig(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
	}{
		{
			name: "missing logger",
			cfg: Config{
				Runner:       successfulRunner(),
				PollInterval: time.Second,
				BatchSize:    10,
			},
		},
		{
			name: "missing runner",
			cfg: Config{
				Logger:       testLogger(),
				PollInterval: time.Second,
				BatchSize:    10,
			},
		},
		{
			name: "invalid poll interval",
			cfg: Config{
				Logger:       testLogger(),
				Runner:       successfulRunner(),
				PollInterval: 0,
				BatchSize:    10,
			},
		},
		{
			name: "invalid batch size",
			cfg: Config{
				Logger:       testLogger(),
				Runner:       successfulRunner(),
				PollInterval: time.Second,
				BatchSize:    0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.cfg)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

func successfulRunner() DueRunner {
	return dueRunnerFunc(func(
		context.Context,
		int,
	) (usecasecheck.RunDueResult, error) {
		return usecasecheck.RunDueResult{}, nil
	})
}

func testLogger() *slog.Logger {
	return slog.New(
		slog.NewJSONHandler(
			io.Discard,
			nil,
		),
	)
}

func assertRunnerCall(
	t *testing.T,
	calls <-chan int,
	wantLimit int,
) {
	t.Helper()

	select {
	case limit := <-calls:
		if limit != wantLimit {
			t.Fatalf(
				"runner limit = %d, want %d",
				limit,
				wantLimit,
			)
		}

	case <-time.After(time.Second):
		t.Fatal("runner was not called")
	}
}

func assertCycleCompleted(
	t *testing.T,
	called <-chan struct{},
) {
	t.Helper()

	select {
	case <-called:
	case <-time.After(time.Second):
		t.Fatal("worker cycle did not complete")
	}
}
