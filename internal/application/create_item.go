package application

import (
	"context"
	"fmt"

	"github.com/flyluman/scratch/internal/domain"
	"github.com/flyluman/scratch/internal/ports"
)

type CreateItemUseCase struct {
	repo  ports.ItemRepository
	idGen ports.IDGenerator
	clock ports.Clock
	bus   ports.EventBus
}

func NewCreateItemUseCase(repo ports.ItemRepository, idGen ports.IDGenerator, clock ports.Clock, bus ports.EventBus) *CreateItemUseCase {
	return &CreateItemUseCase{repo: repo, idGen: idGen, clock: clock, bus: bus}
}

func (uc *CreateItemUseCase) Execute(ctx context.Context, actorID, name, description string) (domain.Item, error) {
	item := domain.Item{
		ID:          uc.idGen.New(),
		Name:        name,
		Description: description,
		CreatedAt:   uc.clock.Now(),
		UpdatedAt:   uc.clock.Now(),
	}

	if err := item.Validate(); err != nil {
		return domain.Item{}, fmt.Errorf("%w: %w", ErrValidation, err)
	}

	if err := uc.repo.Save(ctx, item); err != nil {
		return domain.Item{}, fmt.Errorf("save item: %w", err)
	}

	_ = uc.bus.Publish(ctx, ports.Event{
		Type:       "CREATE",
		ActorID:    actorID,
		Resource:   "item",
		ResourceID: item.ID,
		NewValue:   item,
	})

	return item, nil
}
