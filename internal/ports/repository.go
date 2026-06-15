package ports

import (
	"context"

	"github.com/flyluman/scratch/internal/domain"
)

type BaseRepository[T any, ID comparable] interface {
	Save(ctx context.Context, entity T) error
	Get(ctx context.Context, id ID) (T, error)
	Update(ctx context.Context, entity T) error
	Delete(ctx context.Context, id ID) error
	SoftDelete(ctx context.Context, id ID) error
}

type ItemRepository interface {
	BaseRepository[domain.Item, string]
	List(ctx context.Context, cursor string, limit int) ([]domain.Item, string, error)
}
