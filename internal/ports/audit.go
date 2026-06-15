package ports

import "context"

type AuditLogger interface {
	Log(ctx context.Context, action, resourceType, resourceID, actorID string, oldVal, newVal any) error
}
