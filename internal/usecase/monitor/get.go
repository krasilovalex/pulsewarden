package monitor

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	domainmonitor "github.com/wayzzoo/pulsewarden/internal/domain/monitor"
)

type GetterRepository interface {
	GetByID(
		ctx context.Context,
		id uuid.UUID,
	) (domainmonitor.Monitor, error)
}

type Get struct {
	repository GetterRepository
}

func NewGet(repository GetterRepository) *Get {
	return &Get{
		repository: repository,
	}
}

func (uc *Get) Execute(
	ctx context.Context,
	id uuid.UUID,
) (domainmonitor.Monitor, error) {
	result, err := uc.repository.GetByID(ctx, id)

	if err != nil {
		return domainmonitor.Monitor{}, fmt.Errorf(
			"get monitor: %w",
			err,
		)
	}
	return result, nil
}
