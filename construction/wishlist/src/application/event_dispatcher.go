package application

import (
	"context"
)

// Event is a local interface matching domain/event.Event to avoid import cycles.
type Event interface {
	EventName() string
}

// EventBus is a local interface that application services depend on.
// Infrastructure implements this via InMemoryEventBus.
type EventBus interface {
	Publish(ctx context.Context, e Event)
}
