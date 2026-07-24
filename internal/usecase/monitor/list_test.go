package monitor

import (
	"context"
	"errors"
	"testing"

	domainmonitor "github.com/krasilovalex/pulsewarden/internal/domain/monitor"
)

type listRepositoryStub struct {
	result []domainmonitor.Monitor
	err    error
}

func (stub listRepositoryStub) List(
	context.Context,
) ([]domainmonitor.Monitor, error) {
	return stub.result, stub.err
}

func TestListExecute(t *testing.T) {
	expected := []domainmonitor.Monitor{
		{Name: "First"},
		{Name: "Second"},
	}

	uc := NewList(listRepositoryStub{
		result: expected,
	})

	result, err := uc.Execute(context.Background())
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if len(result) != len(expected) {
		t.Fatalf(
			"result length = %d, want %d",
			len(result),
			len(expected),
		)
	}
}

func TestListExecuteRepositoryError(t *testing.T) {
	expectedErr := errors.New("repository failed")

	uc := NewList(listRepositoryStub{
		err: expectedErr,
	})

	_, err := uc.Execute(context.Background())
	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"Execute() error = %v, want %v",
			err,
			expectedErr,
		)
	}
}
