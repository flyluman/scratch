package memory

import (
	"context"
	"sync"
	"time"

	"github.com/flyluman/scratch/internal/domain"
)

type ItemRepository struct {
	mu    sync.RWMutex
	items map[string]domain.Item
}

func NewItemRepository() *ItemRepository {
	return &ItemRepository{items: make(map[string]domain.Item)}
}

func (r *ItemRepository) Save(ctx context.Context, item domain.Item) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[item.ID] = item
	return nil
}

func (r *ItemRepository) Get(ctx context.Context, id string) (domain.Item, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	item, ok := r.items[id]
	if !ok || item.DeletedAt != nil {
		return domain.Item{}, domain.ErrNotFound
	}
	return item, nil
}

func (r *ItemRepository) List(ctx context.Context, cursor string, limit int) ([]domain.Item, string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := make([]domain.Item, 0, len(r.items))
	for _, item := range r.items {
		items = append(items, item)
	}
	if len(items) == 0 {
		return nil, "", nil
	}
	return items, "", nil
}

func (r *ItemRepository) Update(ctx context.Context, item domain.Item) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[item.ID]; !ok {
		return domain.ErrNotFound
	}
	r.items[item.ID] = item
	return nil
}

func (r *ItemRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[id]; !ok {
		return domain.ErrNotFound
	}
	delete(r.items, id)
	return nil
}

func (r *ItemRepository) SoftDelete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	item, ok := r.items[id]
	if !ok {
		return domain.ErrNotFound
	}
	now := time.Now().UTC()
	item.DeletedAt = &now
	r.items[id] = item
	return nil
}
