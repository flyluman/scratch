package application

import (
	"context"
	"fmt"

	"github.com/flyluman/scratch/internal/ports"
)

type DeleteItemUseCase struct {
	repo  ports.ItemRepository
	flags ports.FeatureFlags
	bus   ports.EventBus
}

func NewDeleteItemUseCase(repo ports.ItemRepository, flags ports.FeatureFlags, bus ports.EventBus) *DeleteItemUseCase {
	return &DeleteItemUseCase{repo: repo, flags: flags, bus: bus}
}

func (uc *DeleteItemUseCase) Execute(ctx context.Context, actorID, id string) error {
	if uc.flags.SoftDelete() {
		if err := uc.repo.SoftDelete(ctx, id); err != nil {
			return fmt.Errorf("%w: %w", ErrNotFound, err)
		}
	} else {
		if err := uc.repo.Delete(ctx, id); err != nil {
			return fmt.Errorf("%w: %w", ErrNotFound, err)
		}
	}

	_ = uc.bus.Publish(ctx, ports.Event{
		Type:       "DELETE",
		ActorID:    actorID,
		Resource:   "item",
		ResourceID: id,
	})

	return nil
}
