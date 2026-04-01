package eventbus

import (
	"context"
	"sync/atomic"
	"testing"
	"wishlist/domain/event"

	"pgregory.net/rapid"
)

// testEvent is a minimal event for testing.
type testEvent struct {
	name string
}

func (e testEvent) EventName() string { return e.name }

// 4.10.1 PBT: all N subscribers receive the published event exactly once
func TestInMemoryEventBus_AllSubscribersReceive(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		n := rapid.IntRange(1, 10).Draw(t, "nSubscribers")
		bus := NewInMemoryEventBus()

		counters := make([]atomic.Int32, n)
		for i := 0; i < n; i++ {
			idx := i
			bus.Subscribe("test.event", func(_ context.Context, _ event.Event) {
				counters[idx].Add(1)
			})
		}

		bus.Publish(context.Background(), testEvent{name: "test.event"})

		for i := 0; i < n; i++ {
			if counters[i].Load() != 1 {
				t.Fatalf("subscriber[%d] called %d times, want 1", i, counters[i].Load())
			}
		}
	})
}

// 4.10.2 PBT: subscribers for event A do not receive events of type B
func TestInMemoryEventBus_SubscriberIsolation(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		bus := NewInMemoryEventBus()

		var aCalled, bCalled atomic.Int32
		bus.Subscribe("event.A", func(_ context.Context, _ event.Event) { aCalled.Add(1) })
		bus.Subscribe("event.B", func(_ context.Context, _ event.Event) { bCalled.Add(1) })

		bus.Publish(context.Background(), testEvent{name: "event.A"})

		if aCalled.Load() != 1 {
			t.Fatalf("event.A handler called %d times, want 1", aCalled.Load())
		}
		if bCalled.Load() != 0 {
			t.Fatalf("event.B handler called %d times after event.A published, want 0", bCalled.Load())
		}
	})
}

// 4.10.3 PBT: handlers are called in registration order
func TestInMemoryEventBus_HandlerCallOrder(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		n := rapid.IntRange(2, 8).Draw(t, "n")
		bus := NewInMemoryEventBus()

		order := make([]int, 0, n)
		for i := 0; i < n; i++ {
			idx := i
			bus.Subscribe("ordered.event", func(_ context.Context, _ event.Event) {
				order = append(order, idx)
			})
		}

		bus.Publish(context.Background(), testEvent{name: "ordered.event"})

		if len(order) != n {
			t.Fatalf("order length: got %d, want %d", len(order), n)
		}
		for i, v := range order {
			if v != i {
				t.Fatalf("order[%d] = %d, want %d", i, v, i)
			}
		}
	})
}

// 4.10.4: no subscribers is a no-op (no panic)
func TestInMemoryEventBus_NoSubscribersNoPanic(t *testing.T) {
	bus := NewInMemoryEventBus()
	// Must not panic
	bus.Publish(context.Background(), testEvent{name: "ghost.event"})
}

// panic recovery: a panicking handler does not prevent other handlers from running
func TestInMemoryEventBus_PanicRecovery(t *testing.T) {
	bus := NewInMemoryEventBus()

	var secondCalled bool
	bus.Subscribe("panic.event", func(_ context.Context, _ event.Event) {
		panic("intentional panic in test")
	})
	bus.Subscribe("panic.event", func(_ context.Context, _ event.Event) {
		secondCalled = true
	})

	// Must not panic out of Publish
	bus.Publish(context.Background(), testEvent{name: "panic.event"})

	if !secondCalled {
		t.Fatal("second handler was not called after first handler panicked")
	}
}

// 4.10.5: handler receives the identical event that was published
func TestInMemoryEventBus_HandlerReceivesCorrectEvent(t *testing.T) {
	bus := NewInMemoryEventBus()
	published := testEvent{name: "exact.event"}

	var received event.Event
	bus.Subscribe("exact.event", func(_ context.Context, e event.Event) {
		received = e
	})

	bus.Publish(context.Background(), published)

	if received == nil {
		t.Fatal("handler was not called")
	}
	got, ok := received.(testEvent)
	if !ok {
		t.Fatalf("received event type: got %T, want testEvent", received)
	}
	if got.name != published.name {
		t.Fatalf("event name: got %q, want %q", got.name, published.name)
	}
}
