package audit

import (
	"context"

	"github.com/flyluman/scratch/internal/ports"
)

func NewEventHandler(logger ports.AuditLogger) ports.EventHandler {
	return func(ctx context.Context, event ports.Event) error {
		return logger.Log(ctx, event.Type, event.Resource, event.ResourceID, event.ActorID, event.OldValue, event.NewValue)
	}
}
