package application

import (
	"context"
	"fmt"

	"github.com/flyluman/scratch/internal/domain"
	"github.com/flyluman/scratch/internal/ports"
)

type UpdateItemUseCase struct {
	repo  ports.ItemRepository
	clock ports.Clock
	bus   ports.EventBus
}

func NewUpdateItemUseCase(repo ports.ItemRepository, clock ports.Clock, bus ports.EventBus) *UpdateItemUseCase {
	return &UpdateItemUseCase{repo: repo, clock: clock, bus: bus}
}

func (uc *UpdateItemUseCase) Execute(ctx context.Context, actorID, id, name, description string) (domain.Item, error) {
	existing, err := uc.repo.Get(ctx, id)
	if err != nil {
		return domain.Item{}, fmt.Errorf("%w: %w", ErrNotFound, err)
	}

	old := existing
	existing.Name = name
	existing.Description = description
	existing.UpdatedAt = uc.clock.Now()

	if err := existing.Validate(); err != nil {
		return domain.Item{}, fmt.Errorf("%w: %w", ErrValidation, err)
	}

	if err := uc.repo.Update(ctx, existing); err != nil {
		return domain.Item{}, fmt.Errorf("update item: %w", err)
	}

	_ = uc.bus.Publish(ctx, ports.Event{
		Type:       "UPDATE",
		ActorID:    actorID,
		Resource:   "item",
		ResourceID: id,
		OldValue:   old,
		NewValue:   existing,
	})

	return existing, nil
}
