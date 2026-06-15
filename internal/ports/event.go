package ports

import "context"

type Event struct {
	Type       string
	ActorID    string
	Resource   string
	ResourceID string
	OldValue   any
	NewValue   any
}

type EventHandler func(ctx context.Context, event Event) error

type EventBus interface {
	Publish(ctx context.Context, event Event) error
	Subscribe(handler EventHandler)
}
