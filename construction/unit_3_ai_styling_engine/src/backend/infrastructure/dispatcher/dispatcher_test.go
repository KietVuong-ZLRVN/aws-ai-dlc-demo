package dispatcher_test

import (
	"ai-styling-engine/domain/events"
	"ai-styling-engine/infrastructure/dispatcher"
	"testing"
)

// minimal event for testing
type testEvent struct{ value string }

func (e testEvent) EventType() string { return "test.event" }

// TC-INFRA-2: Two handlers for same event type — both called in registration order
func TestInProcessEventDispatcher_TwoHandlers_BothCalledInOrder(t *testing.T) {
	d := dispatcher.NewInProcessEventDispatcher()
	var order []int
	d.Register("test.event", func(e events.DomainEvent) { order = append(order, 1) })
	d.Register("test.event", func(e events.DomainEvent) { order = append(order, 2) })

	d.Dispatch(testEvent{})

	if len(order) != 2 {
		t.Fatalf("expected 2 handlers called, got %d", len(order))
	}
	if order[0] != 1 || order[1] != 2 {
		t.Errorf("expected order [1,2], got %v", order)
	}
}

// TC-INFRA-3: Dispatching event with no registered handlers — does not panic
func TestInProcessEventDispatcher_NoHandlers_DoesNotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Dispatch with no handlers panicked: %v", r)
		}
	}()
	d := dispatcher.NewInProcessEventDispatcher()
	d.Dispatch(testEvent{value: "no handlers"})
}

// TC-INFRA-4: A panicking handler — panic is recovered and re-panicked as a wrapped error
func TestInProcessEventDispatcher_PanickingHandler_PropagatesPanic(t *testing.T) {
	d := dispatcher.NewInProcessEventDispatcher()
	d.Register("test.event", func(e events.DomainEvent) {
		panic("handler failure")
	})

	defer func() {
		r := recover()
		if r == nil {
			t.Error("expected panic to propagate from handler, but it was swallowed")
			return
		}
		// The dispatcher wraps the panic as an error.
		if _, ok := r.(error); !ok {
			t.Errorf("expected panic value to be an error, got %T: %v", r, r)
		}
	}()
	d.Dispatch(testEvent{})
}
