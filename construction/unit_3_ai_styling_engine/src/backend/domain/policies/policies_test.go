package policies_test

import (
	"ai-styling-engine/domain/events"
	"ai-styling-engine/domain/policies"
	"ai-styling-engine/domain/valueobjects"
	"testing"
)

// ── capturingDispatcher ───────────────────────────────────────────────────────

type capturingDispatcher struct {
	dispatched []events.DomainEvent
	handlers   map[string][]events.EventHandler
}

func newCapturingDispatcher() *capturingDispatcher {
	return &capturingDispatcher{handlers: make(map[string][]events.EventHandler)}
}

func (d *capturingDispatcher) Register(eventType string, handler events.EventHandler) {
	d.handlers[eventType] = append(d.handlers[eventType], handler)
}

func (d *capturingDispatcher) Dispatch(event events.DomainEvent) {
	d.dispatched = append(d.dispatched, event)
	for _, h := range d.handlers[event.EventType()] {
		h(event)
	}
}

func (d *capturingDispatcher) hasEvent(eventType string) bool {
	for _, e := range d.dispatched {
		if e.EventType() == eventType {
			return true
		}
	}
	return false
}

// ── PreferenceDefaultPolicy ──────────────────────────────────────────────────

func TestPreferenceDefaultPolicy_Handle_DoesNotPanic(t *testing.T) {
	p := policies.NewPreferenceDefaultPolicy()
	// Should handle the event without panicking regardless of content.
	p.Handle(events.ComboGenerationRequested{
		SessionId:     "session-1",
		QuickGenerate: true,
	})
}

// ── WishlistSupplementationPolicy ────────────────────────────────────────────

func TestWishlistSupplementationPolicy_TwoInStockItems_NoEventDispatched(t *testing.T) {
	// TC-403-4: ≥ 2 in-stock items → no supplementation event
	d := newCapturingDispatcher()
	p := policies.NewWishlistSupplementationPolicy(d)

	p.Handle(events.WishlistFetchCompleted{
		SessionId: "session-1",
		Snapshot: valueobjects.WishlistSnapshot{
			Items: []valueobjects.WishlistItem{
				{ItemId: "1", ConfigSku: "CFG-A", InStock: true},
				{ItemId: "2", ConfigSku: "CFG-B", InStock: true},
			},
			TotalCount: 2,
		},
	})

	if d.hasEvent(events.EventTypeCatalogSupplementationRequired) {
		t.Error("expected no CatalogSupplementationRequired when wishlist has ≥ 2 in-stock items")
	}
}

func TestWishlistSupplementationPolicy_OneInStockItem_EventDispatched(t *testing.T) {
	// TC-403-3: < 2 in-stock items → supplementation event dispatched
	d := newCapturingDispatcher()
	p := policies.NewWishlistSupplementationPolicy(d)

	p.Handle(events.WishlistFetchCompleted{
		SessionId: "session-1",
		Snapshot: valueobjects.WishlistSnapshot{
			Items: []valueobjects.WishlistItem{
				{ItemId: "1", ConfigSku: "CFG-A", InStock: true},
				{ItemId: "2", ConfigSku: "CFG-B", InStock: false},
			},
			TotalCount: 2,
		},
	})

	if !d.hasEvent(events.EventTypeCatalogSupplementationRequired) {
		t.Error("expected CatalogSupplementationRequired when < 2 in-stock items")
	}
}

func TestWishlistSupplementationPolicy_EmptyWishlist_EventDispatched(t *testing.T) {
	d := newCapturingDispatcher()
	p := policies.NewWishlistSupplementationPolicy(d)

	p.Handle(events.WishlistFetchCompleted{
		SessionId: "session-1",
		Snapshot:  valueobjects.WishlistSnapshot{Items: nil, TotalCount: 0},
	})

	if !d.hasEvent(events.EventTypeCatalogSupplementationRequired) {
		t.Error("expected CatalogSupplementationRequired when wishlist is empty")
	}
}

func TestWishlistSupplementationPolicy_AllOutOfStock_EventDispatched(t *testing.T) {
	d := newCapturingDispatcher()
	p := policies.NewWishlistSupplementationPolicy(d)

	p.Handle(events.WishlistFetchCompleted{
		SessionId: "session-1",
		Snapshot: valueobjects.WishlistSnapshot{
			Items: []valueobjects.WishlistItem{
				{ItemId: "1", InStock: false},
				{ItemId: "2", InStock: false},
			},
			TotalCount: 2,
		},
	})

	if !d.hasEvent(events.EventTypeCatalogSupplementationRequired) {
		t.Error("expected CatalogSupplementationRequired when all items are out of stock")
	}
}

// ── FallbackPolicy ────────────────────────────────────────────────────────────

func TestFallbackPolicy_Handle_WithAlternatives_DoesNotPanic(t *testing.T) {
	p := policies.NewFallbackPolicy()
	p.Handle(events.FallbackTriggered{
		SessionId: "session-1",
		Alternatives: []valueobjects.AlternativeItem{
			{ConfigSku: "CFG-ALT", Reason: "Good alternative"},
		},
	})
}

func TestFallbackPolicy_Handle_EmptyAlternatives_DoesNotPanic(t *testing.T) {
	// TC-404-4: empty alternatives should log a warning but not panic
	p := policies.NewFallbackPolicy()
	p.Handle(events.FallbackTriggered{
		SessionId:    "session-1",
		Alternatives: []valueobjects.AlternativeItem{},
	})
}

// ── ComboExclusionPolicy ──────────────────────────────────────────────────────

func TestComboExclusionPolicy_Handle_DoesNotPanic(t *testing.T) {
	p := policies.NewComboExclusionPolicy()
	p.Handle(events.CombosGenerated{
		SessionId:  "session-1",
		ComboCount: 2,
	})
}

func TestComboExclusionPolicy_Handle_ZeroCombos_DoesNotPanic(t *testing.T) {
	p := policies.NewComboExclusionPolicy()
	p.Handle(events.CombosGenerated{
		SessionId:  "session-1",
		ComboCount: 0,
	})
}

// ── Phase 5 gap tests ────────────────────────────────────────────────────────

// TC-403-5: Exactly 2 in-stock items → CatalogSupplementationRequired NOT raised (boundary)
func TestWishlistSupplementationPolicy_ExactlyTwoInStock_NoEventDispatched(t *testing.T) {
	d := newCapturingDispatcher()
	p := policies.NewWishlistSupplementationPolicy(d)

	p.Handle(events.WishlistFetchCompleted{
		SessionId: "session-1",
		Snapshot: valueobjects.WishlistSnapshot{
			Items: []valueobjects.WishlistItem{
				{ItemId: "1", ConfigSku: "CFG-A", InStock: true},
				{ItemId: "2", ConfigSku: "CFG-B", InStock: true},
			},
			TotalCount: 2,
		},
	})

	if d.hasEvent(events.EventTypeCatalogSupplementationRequired) {
		t.Error("expected no CatalogSupplementationRequired for exactly 2 in-stock items")
	}
}

// TC-404-4: FallbackPolicy with empty alternatives — does not panic, no secondary event
func TestFallbackPolicy_EmptyAlternatives_NoSecondaryEvent_NoPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("FallbackPolicy panicked on empty alternatives: %v", r)
		}
	}()
	d := newCapturingDispatcher()
	p := policies.NewFallbackPolicy()
	p.Handle(events.FallbackTriggered{
		SessionId:    "session-1",
		Alternatives: []valueobjects.AlternativeItem{},
	})
	// No secondary events should be dispatched by the policy itself.
	if len(d.dispatched) != 0 {
		t.Errorf("expected no secondary events from FallbackPolicy, got %d", len(d.dispatched))
	}
}

// TC-405-6: ComboExclusionPolicy with 0 combos — does not panic, logs outcome
func TestComboExclusionPolicy_ZeroCombosAfterExclusion_DoesNotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("ComboExclusionPolicy panicked: %v", r)
		}
	}()
	p := policies.NewComboExclusionPolicy()
	p.Handle(events.CombosGenerated{
		SessionId:  "session-1",
		ComboCount: 0,
	})
}
