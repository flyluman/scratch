package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/flyluman/scratch/internal/domain"
	"github.com/jackc/pgx/v5"
)

type ItemRepository struct {
	pool *Pool
}

func NewItemRepository(pool *Pool) *ItemRepository {
	return &ItemRepository{pool: pool}
}

func (r *ItemRepository) Save(ctx context.Context, item domain.Item) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO items (id, name, description, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		item.ID, item.Name, item.Description, item.CreatedAt, item.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert item: %w", err)
	}
	return nil
}

func (r *ItemRepository) Get(ctx context.Context, id string) (domain.Item, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, name, description, created_at, updated_at, deleted_at
		 FROM items WHERE id = $1 AND deleted_at IS NULL`, id)

	var item domain.Item
	err := row.Scan(&item.ID, &item.Name, &item.Description, &item.CreatedAt, &item.UpdatedAt, &item.DeletedAt)
	if err != nil {
		return domain.Item{}, fmt.Errorf("%w: %w", domain.ErrNotFound, err)
	}
	return item, nil
}

func (r *ItemRepository) List(ctx context.Context, cursor string, limit int) ([]domain.Item, string, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	var rows pgx.Rows
	var err error

	if cursor == "" {
		rows, err = r.pool.Query(ctx,
			`SELECT id, name, description, created_at, updated_at, deleted_at
			 FROM items WHERE deleted_at IS NULL
			 ORDER BY created_at DESC, id DESC
			 LIMIT $1`, limit+1)
	} else {
		// cursor = last_id:last_created_at (encoded)
		parts := strings.SplitN(cursor, ":", 2)
		if len(parts) != 2 {
			return nil, "", fmt.Errorf("invalid cursor")
		}
		lastID := parts[0]
		lastCreatedAt, err := time.Parse(time.RFC3339Nano, parts[1])
		if err != nil {
			return nil, "", fmt.Errorf("invalid cursor time: %w", err)
		}

		rows, err = r.pool.Query(ctx,
			`SELECT id, name, description, created_at, updated_at, deleted_at
			 FROM items WHERE deleted_at IS NULL
			   AND (created_at < $2 OR (created_at = $2 AND id < $1))
			 ORDER BY created_at DESC, id DESC
			 LIMIT $3`, lastID, lastCreatedAt, limit+1)
	}

	if err != nil {
		return nil, "", fmt.Errorf("list items: %w", err)
	}
	defer rows.Close()

	var items []domain.Item
	for rows.Next() {
		var item domain.Item
		if err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.CreatedAt, &item.UpdatedAt, &item.DeletedAt); err != nil {
			return nil, "", fmt.Errorf("scan: %w", err)
		}
		items = append(items, item)
	}

	var nextCursor string
	if len(items) > limit {
		items = items[:limit]
		last := items[len(items)-1]
		nextCursor = last.ID + ":" + last.CreatedAt.Format(time.RFC3339Nano)
	}

	return items, nextCursor, nil
}

func (r *ItemRepository) Update(ctx context.Context, item domain.Item) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE items SET name = $1, description = $2, updated_at = $3
		 WHERE id = $4 AND deleted_at IS NULL`,
		item.Name, item.Description, item.UpdatedAt, item.ID,
	)
	if err != nil {
		return fmt.Errorf("update item: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("%w: item %s", domain.ErrNotFound, item.ID)
	}
	return nil
}

func (r *ItemRepository) Delete(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM items WHERE id = $1 AND deleted_at IS NULL`, id)
	if err != nil {
		return fmt.Errorf("delete item: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("%w: item %s", domain.ErrNotFound, id)
	}
	return nil
}

func (r *ItemRepository) SoftDelete(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE items SET deleted_at = NOW(), updated_at = NOW()
		 WHERE id = $1 AND deleted_at IS NULL`, id)
	if err != nil {
		return fmt.Errorf("soft delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("%w: item %s", domain.ErrNotFound, id)
	}
	return nil
}
