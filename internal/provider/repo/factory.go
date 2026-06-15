package repo

import (
	"github.com/flyluman/scratch/internal/adapters/outbound/memory"
	"github.com/flyluman/scratch/internal/adapters/outbound/postgres"
	"github.com/flyluman/scratch/internal/ports"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Factory struct {
	pool *pgxpool.Pool
}

func NewFactory(pool *pgxpool.Pool) *Factory {
	return &Factory{pool: pool}
}

func (f *Factory) NewItemRepository() ports.ItemRepository {
	if f.pool != nil {
		p := &postgres.Pool{Pool: f.pool}
		return postgres.NewItemRepository(p)
	}
	return memory.NewItemRepository()
}
