package application

import (
	"context"
	"fmt"

	"github.com/flyluman/scratch/internal/domain"
	"github.com/flyluman/scratch/internal/ports"
)

type GetItemUseCase struct {
	repo ports.ItemRepository
}

func NewGetItemUseCase(repo ports.ItemRepository) *GetItemUseCase {
	return &GetItemUseCase{repo: repo}
}

func (uc *GetItemUseCase) Execute(ctx context.Context, id string) (domain.Item, error) {
	item, err := uc.repo.Get(ctx, id)
	if err != nil {
		return domain.Item{}, fmt.Errorf("%w: %w", ErrNotFound, err)
	}
	return item, nil
}
