package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresAuditLogger struct {
	pool *pgxpool.Pool
}

func NewPostgresAuditLogger(pool *pgxpool.Pool) *PostgresAuditLogger {
	return &PostgresAuditLogger{pool: pool}
}

func (l *PostgresAuditLogger) Log(ctx context.Context, action, resourceType, resourceID, actorID string, oldVal, newVal any) error {
	var oldJSON, newJSON *json.RawMessage
	if oldVal != nil {
		b, _ := json.Marshal(oldVal)
		m := json.RawMessage(b)
		oldJSON = &m
	}
	if newVal != nil {
		b, _ := json.Marshal(newVal)
		m := json.RawMessage(b)
		newJSON = &m
	}

	_, err := l.pool.Exec(ctx,
		`INSERT INTO audit_events (id, action, resource_type, resource_id, actor_id, old_value, new_value, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		uuid.NewString(), action, resourceType, resourceID, actorID, oldJSON, newJSON, time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("insert audit event: %w", err)
	}
	return nil
}
