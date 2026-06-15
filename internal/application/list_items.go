package application

import (
	"context"
	"errors"

	"github.com/flyluman/scratch/internal/domain"
	"github.com/flyluman/scratch/internal/ports"
)

type ListItemUseCase struct {
	repo ports.ItemRepository
}

func NewListItemUseCase(repo ports.ItemRepository) *ListItemUseCase {
	return &ListItemUseCase{repo: repo}
}

func (uc *ListItemUseCase) Execute(ctx context.Context, cursor string, limit int) ([]domain.Item, string, error) {
	if limit <= 0 || limit > 100 {
		return nil, "", errors.New("limit must be 1–100")
	}
	return uc.repo.List(ctx, cursor, limit)
}
