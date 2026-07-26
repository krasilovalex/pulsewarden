package monitor

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	domainmonitor "github.com/wayzzoo/pulsewarden/internal/domain/monitor"
)

type getterRepositoryStub struct {
	result domainmonitor.Monitor
	err    error
	gotID  uuid.UUID
}

func (stub *getterRepositoryStub) GetByID(
	_ context.Context,
	id uuid.UUID,
) (domainmonitor.Monitor, error) {
	stub.gotID = id

	return stub.result, stub.err
}

func TestGetExecute(t *testing.T) {
	id := uuid.New()

	expected := domainmonitor.Monitor{
		ID:   id,
		Name: "Pulsewarden API",
	}

	repository := &getterRepositoryStub{
		result: expected,
	}

	uc := NewGet(repository)

	result, err := uc.Execute(
		context.Background(),
		id,
	)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if repository.gotID != id {
		t.Fatalf(
			"repository ID = %s, wand %s",
			repository.gotID,
			id,
		)
	}

	if result.ID != expected.ID {
		t.Fatalf(
			"result iD = %s, want %s",
			result.ID,
			expected.ID,
		)
	}

	if result.Name != expected.Name {
		t.Fatalf(
			"result name = %q, want %q",
			result.Name,
			expected.Name,
		)
	}
}

func TestGetExecuteRepositoryError(t *testing.T) {
	expectedErr := errors.New("repository failed")

	repository := &getterRepositoryStub{
		err: expectedErr,
	}

	uc := NewGet(repository)

	_, err := uc.Execute(
		context.Background(),
		uuid.New(),
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"Execute() error = %v, want %v",
			err,
			expectedErr,
		)
	}
}
