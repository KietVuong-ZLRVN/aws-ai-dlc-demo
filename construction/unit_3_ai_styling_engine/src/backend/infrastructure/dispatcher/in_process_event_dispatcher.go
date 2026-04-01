package dispatcher

import (
	"ai-styling-engine/domain/events"
	"fmt"
)

// InProcessEventDispatcher is a synchronous, in-memory event dispatcher.
// Handlers are invoked in registration order within the same call stack.
// If a handler panics, the panic is recovered and re-panicked as a wrapped error.
type InProcessEventDispatcher struct {
	handlers map[string][]events.EventHandler
}

func NewInProcessEventDispatcher() *InProcessEventDispatcher {
	return &InProcessEventDispatcher{
		handlers: make(map[string][]events.EventHandler),
	}
}

// Register adds a handler for the given event type.
// Multiple handlers may be registered for the same event type.
func (d *InProcessEventDispatcher) Register(eventType string, handler events.EventHandler) {
	d.handlers[eventType] = append(d.handlers[eventType], handler)
}

// Dispatch synchronously invokes all handlers registered for the event's type.
// If a handler panics, the panic is recovered and re-panicked as a wrapped error
// so callers can distinguish handler failures from other panics.
func (d *InProcessEventDispatcher) Dispatch(event events.DomainEvent) {
	for _, handler := range d.handlers[event.EventType()] {
		func() {
			defer func() {
				if r := recover(); r != nil {
					panic(fmt.Errorf("event handler panic for %s: %v", event.EventType(), r))
				}
			}()
			handler(event)
		}()
	}
}
