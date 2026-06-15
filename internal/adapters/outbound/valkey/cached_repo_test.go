package valkey_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/flyluman/scratch/internal/adapters/outbound/memory"
	"github.com/flyluman/scratch/internal/adapters/outbound/valkey"
	"github.com/flyluman/scratch/internal/domain"
	"github.com/flyluman/scratch/internal/platform/config"
	"github.com/flyluman/scratch/internal/testutil"
	"github.com/google/uuid"
)

func TestCachedItemRepositoryIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skip integration test in short mode")
	}

	ctx := context.Background()
	containers, err := testutil.StartContainers(ctx)
	if err != nil {
		t.Fatalf("start containers: %v", err)
	}
	defer containers.Cleanup()

	cfg := config.Config{
		ValkeyAddr: containers.ValkeyAddr,
		ValkeyDB:   0,
	}

	client, err := valkey.NewClient(cfg)
	if err != nil {
		t.Fatalf("new valkey client: %v", err)
	}

	cache := valkey.NewItemCache(client)
	primary := memory.NewItemRepository()
	repo := valkey.NewCachedItemRepository(primary, cache)

	t.Run("Save caches item", func(t *testing.T) {
		item := domain.Item{
			ID:          uuid.NewString(),
			Name:        "cached-item",
			Description: "cached description",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		}

		if err := repo.Save(ctx, item); err != nil {
			t.Fatalf("save: %v", err)
		}

		raw, err := cache.Get(ctx, "item:"+item.ID)
		if err != nil {
			t.Fatalf("cache get: %v", err)
		}

		var got domain.Item
		if err := json.Unmarshal([]byte(raw), &got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got.Name != item.Name {
			t.Fatalf("expected %s, got %s", item.Name, got.Name)
		}
	})

	t.Run("Get reads from cache on second call", func(t *testing.T) {
		item := domain.Item{
			ID:          uuid.NewString(),
			Name:        "cache-hit",
			Description: "test",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		}

		if err := repo.Save(ctx, item); err != nil {
			t.Fatalf("save: %v", err)
		}

		got, err := repo.Get(ctx, item.ID)
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		if got.Name != item.Name {
			t.Fatalf("expected %s, got %s", item.Name, got.Name)
		}
	})
}
