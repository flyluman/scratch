package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/flyluman/scratch/internal/adapters/outbound/postgres"
	"github.com/flyluman/scratch/internal/domain"
	"github.com/flyluman/scratch/internal/platform/migrations"
	"github.com/flyluman/scratch/internal/testutil"
	"github.com/google/uuid"
)

func TestPostgresItemRepositoryIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skip integration test in short mode")
	}

	ctx := context.Background()
	containers, err := testutil.StartContainers(ctx)
	if err != nil {
		t.Fatalf("start containers: %v", err)
	}
	defer containers.Cleanup()

	pool, err := postgres.NewPool(ctx, containers.PostgresDSN)
	if err != nil {
		t.Fatalf("new pool: %v", err)
	}
	defer pool.Close()

	runner := migrations.NewRunner(pool.Pool, "migrations")
	if err := runner.Run(ctx); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	repo := postgres.NewItemRepository(pool)

	t.Run("Save and Get", func(t *testing.T) {
		item := domain.Item{
			ID:          uuid.NewString(),
			Name:        "test-item",
			Description: "test description",
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
			t.Fatalf("expected name %s, got %s", item.Name, got.Name)
		}
	})

	t.Run("Get not found", func(t *testing.T) {
		_, err := repo.Get(ctx, "nonexistent")
		if err == nil {
			t.Fatal("expected error for nonexistent item")
		}
	})
}
