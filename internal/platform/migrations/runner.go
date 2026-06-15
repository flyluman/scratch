package migrations

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Runner struct {
	pool          *pgxpool.Pool
	migrationsDir string
}

func NewRunner(pool *pgxpool.Pool, migrationsDir string) *Runner {
	return &Runner{pool: pool, migrationsDir: migrationsDir}
}

func (r *Runner) Run(ctx context.Context) error {
	if err := r.ensureSchemaMigrations(ctx); err != nil {
		return fmt.Errorf("ensure schema_migrations: %w", err)
	}

	entries, err := os.ReadDir(r.migrationsDir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".up.sql") {
			files = append(files, e.Name())
		}
	}

	sort.Strings(files)

	for _, file := range files {
		name := strings.TrimSuffix(file, ".up.sql")
		applied, err := r.hasMigration(ctx, name)
		if err != nil {
			return fmt.Errorf("check migration %s: %w", name, err)
		}
		if applied {
			continue
		}

		content, err := os.ReadFile(filepath.Join(r.migrationsDir, file))
		if err != nil {
			return fmt.Errorf("read migration %s: %w", file, err)
		}

		if err := r.applyMigration(ctx, name, string(content)); err != nil {
			return fmt.Errorf("apply migration %s: %w", file, err)
		}
	}

	return nil
}

func (r *Runner) ensureSchemaMigrations(ctx context.Context) error {
	const stmt = `CREATE TABLE IF NOT EXISTS schema_migrations (
	version TEXT PRIMARY KEY,
	applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
);`
	_, err := r.pool.Exec(ctx, stmt)
	return err
}

func (r *Runner) hasMigration(ctx context.Context, version string) (bool, error) {
	var exists bool
	row := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)`, version)
	if err := row.Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

func (r *Runner) applyMigration(ctx context.Context, version, sqlStmt string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if _, err := tx.Exec(ctx, sqlStmt); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, version); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
