package valkey

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/flyluman/scratch/internal/domain"
	"github.com/flyluman/scratch/internal/ports"
	"github.com/sony/gobreaker"
)

type CachedItemRepository struct {
	primary ports.ItemRepository
	cache   ports.Cache
	cb      *gobreaker.CircuitBreaker
	ttl     time.Duration
}

func NewCachedItemRepository(primary ports.ItemRepository, cache ports.Cache) *CachedItemRepository {
	return &CachedItemRepository{
		primary: primary,
		cache:   cache,
		cb: gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:        "valkey-cache",
			MaxRequests: 3,
			Interval:    10 * time.Second,
			Timeout:     30 * time.Second,
		}),
		ttl: 5 * time.Minute,
	}
}

func (r *CachedItemRepository) cacheKey(id string) string {
	return "item:" + id
}

func (r *CachedItemRepository) Save(ctx context.Context, item domain.Item) error {
	if err := r.primary.Save(ctx, item); err != nil {
		return err
	}
	go r.setCache(context.Background(), item)
	return nil
}

func (r *CachedItemRepository) Get(ctx context.Context, id string) (domain.Item, error) {
	val, err := r.cb.Execute(func() (any, error) {
		return r.cache.Get(ctx, r.cacheKey(id))
	})
	if err == nil {
		return decodeItem(val.(string))
	}

	item, err := r.primary.Get(ctx, id)
	if err != nil {
		return domain.Item{}, err
	}

	go r.setCache(context.Background(), item)
	return item, nil
}

func (r *CachedItemRepository) List(ctx context.Context, cursor string, limit int) ([]domain.Item, string, error) {
	return r.primary.List(ctx, cursor, limit)
}

func (r *CachedItemRepository) Update(ctx context.Context, item domain.Item) error {
	if err := r.primary.Update(ctx, item); err != nil {
		return err
	}
	go r.setCache(context.Background(), item)
	return nil
}

func (r *CachedItemRepository) Delete(ctx context.Context, id string) error {
	if err := r.primary.Delete(ctx, id); err != nil {
		return err
	}
	go func() { _ = r.cache.Del(context.Background(), r.cacheKey(id)) }()
	return nil
}

func (r *CachedItemRepository) SoftDelete(ctx context.Context, id string) error {
	if err := r.primary.SoftDelete(ctx, id); err != nil {
		return err
	}
	go func() { _ = r.cache.Del(context.Background(), r.cacheKey(id)) }()
	return nil
}

func (r *CachedItemRepository) setCache(ctx context.Context, item domain.Item) {
	data, err := json.Marshal(item)
	if err != nil {
		return
	}
	_ = r.cache.Set(ctx, r.cacheKey(item.ID), string(data), r.ttl)
}

func decodeItem(data string) (domain.Item, error) {
	var item domain.Item
	if err := json.Unmarshal([]byte(data), &item); err != nil {
		return domain.Item{}, fmt.Errorf("cache decode: %w", err)
	}
	return item, nil
}
