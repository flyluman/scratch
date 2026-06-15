package bus

import (
	"context"

	"github.com/flyluman/scratch/internal/ports"
)

type MemoryBus struct {
	handlers []ports.EventHandler
}

func NewMemoryBus() *MemoryBus {
	return &MemoryBus{}
}

func (b *MemoryBus) Publish(ctx context.Context, event ports.Event) error {
	for _, h := range b.handlers {
		h := h
		go func() {
			_ = h(ctx, event)
		}()
	}
	return nil
}

func (b *MemoryBus) Subscribe(handler ports.EventHandler) {
	b.handlers = append(b.handlers, handler)
}
