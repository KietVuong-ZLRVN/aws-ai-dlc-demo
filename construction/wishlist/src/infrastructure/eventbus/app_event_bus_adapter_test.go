package eventbus

import (
	"context"
	"testing"
	"wishlist/application"
	"wishlist/domain/event"
)

// appTestEvent implements both application.Event and domain/event.Event.
type appTestEvent struct{ name string }

func (e appTestEvent) EventName() string { return e.name }

// appOnlyEvent implements only application.Event (not domain/event.Event).
type appOnlyEvent struct{ name string }

func (e appOnlyEvent) EventName() string { return e.name }

// Verify interface satisfaction at compile time.
var _ application.Event = appTestEvent{}
var _ event.Event = appTestEvent{}
var _ application.Event = appOnlyEvent{}

// 4.10.6: dual-interface event is passed through directly (no wrapping)
func TestAppEventBusAdapter_DualInterfaceEvent(t *testing.T) {
	bus := NewInMemoryEventBus()
	adapter := NewAppEventBusAdapter(bus)

	var received event.Event
	bus.Subscribe("dual.event", func(_ context.Context, e event.Event) {
		received = e
	})

	published := appTestEvent{name: "dual.event"}
	adapter.Publish(context.Background(), published)

	if received == nil {
		t.Fatal("handler was not called")
	}
	got, ok := received.(appTestEvent)
	if !ok {
		t.Fatalf("received type: got %T, want appTestEvent", received)
	}
	if got.name != published.name {
		t.Fatalf("event name: got %q, want %q", got.name, published.name)
	}
}

// 4.10.7: application-only event is wrapped and routed correctly
func TestAppEventBusAdapter_AppOnlyEvent(t *testing.T) {
	bus := NewInMemoryEventBus()
	adapter := NewAppEventBusAdapter(bus)

	var received event.Event
	bus.Subscribe("app.only.event", func(_ context.Context, e event.Event) {
		received = e
	})

	adapter.Publish(context.Background(), appOnlyEvent{name: "app.only.event"})

	if received == nil {
		t.Fatal("handler was not called")
	}
	if received.EventName() != "app.only.event" {
		t.Fatalf("event name: got %q, want %q", received.EventName(), "app.only.event")
	}
}
