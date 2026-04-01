package eventbus

import (
	"context"
	"log/slog"
	"sync"
	"wishlist/domain/event"
)

// EventHandler is a function that handles a domain event.
type EventHandler func(ctx context.Context, e event.Event)

// EventBus defines the interface for publishing and subscribing to domain events.
type EventBus interface {
	Publish(ctx context.Context, e event.Event)
	Subscribe(eventName string, handler EventHandler)
}

// InMemoryEventBus is a synchronous, in-memory implementation of EventBus.
type InMemoryEventBus struct {
	mu       sync.RWMutex
	handlers map[string][]EventHandler
}

func NewInMemoryEventBus() *InMemoryEventBus {
	return &InMemoryEventBus{
		handlers: make(map[string][]EventHandler),
	}
}

func (b *InMemoryEventBus) Subscribe(eventName string, handler EventHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[eventName] = append(b.handlers[eventName], handler)
}

// Publish calls all registered handlers synchronously.
// Handler panics are recovered and logged.
func (b *InMemoryEventBus) Publish(ctx context.Context, e event.Event) {
	b.mu.RLock()
	handlers := make([]EventHandler, len(b.handlers[e.EventName()]))
	copy(handlers, b.handlers[e.EventName()])
	b.mu.RUnlock()

	for _, h := range handlers {
		func() {
			defer func() {
				if r := recover(); r != nil {
					slog.Error("event handler panicked", "event", e.EventName(), "panic", r)
				}
			}()
			h(ctx, e)
		}()
	}
}
