package audit

import "context"

type NoOpAuditLogger struct{}

func NewNoOpAuditLogger() *NoOpAuditLogger {
	return &NoOpAuditLogger{}
}

func (n *NoOpAuditLogger) Log(ctx context.Context, action, resourceType, resourceID, actorID string, oldVal, newVal any) error {
	return nil
}
