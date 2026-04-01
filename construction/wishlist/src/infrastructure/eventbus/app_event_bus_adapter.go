package eventbus

import (
	"context"
	"wishlist/application"
	"wishlist/domain/event"
)

// appEvent wraps an application.Event to satisfy domain/event.Event.
type appEvent struct {
	inner application.Event
}

func (a appEvent) EventName() string {
	return a.inner.EventName()
}

// AppEventBusAdapter adapts InMemoryEventBus to satisfy application.EventBus.
// application.EventBus.Publish takes application.Event; domain eventbus takes domain/event.Event.
type AppEventBusAdapter struct {
	bus *InMemoryEventBus
}

func NewAppEventBusAdapter(bus *InMemoryEventBus) *AppEventBusAdapter {
	return &AppEventBusAdapter{bus: bus}
}

// Publish satisfies application.EventBus by wrapping the application.Event.
func (a *AppEventBusAdapter) Publish(ctx context.Context, e application.Event) {
	// domain events (WishlistItemAdded, etc.) implement both interfaces.
	// Try direct type assertion first; fall back to wrapping.
	if de, ok := e.(event.Event); ok {
		a.bus.Publish(ctx, de)
	} else {
		a.bus.Publish(ctx, appEvent{inner: e})
	}
}
