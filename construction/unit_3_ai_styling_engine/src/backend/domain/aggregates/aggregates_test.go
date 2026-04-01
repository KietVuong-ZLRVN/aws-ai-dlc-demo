package aggregates_test

import (
	"ai-styling-engine/domain/aggregates"
	"ai-styling-engine/domain/entities"
	"ai-styling-engine/domain/events"
	"ai-styling-engine/domain/services"
	"ai-styling-engine/domain/valueobjects"
	"fmt"
	"testing"
)

// ── Test helpers ──────────────────────────────────────────────────────────────

// capturingDispatcher records every event dispatched to it.
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

func newSession(d events.EventDispatcher, prefs *valueobjects.StylePreferences) *aggregates.StyleSession {
	return aggregates.NewStyleSession(
		"session-test-id",
		valueobjects.ShopperSession{SessionToken: "tok"},
		prefs,
		valueobjects.NewExcludedComboIds(nil),
		d,
	)
}

func twoInStockItems() valueobjects.WishlistSnapshot {
	return valueobjects.WishlistSnapshot{
		Items: []valueobjects.WishlistItem{
			{ItemId: "1", ConfigSku: "CFG-A", InStock: true},
			{ItemId: "2", ConfigSku: "CFG-B", InStock: true},
		},
		TotalCount: 2,
	}
}

// ── StyleSession ──────────────────────────────────────────────────────────────

func TestStyleSession_Construction_RaisesComboGenerationRequested(t *testing.T) {
	d := newCapturingDispatcher()
	newSession(d, nil)
	if !d.hasEvent(events.EventTypeComboGenerationRequested) {
		t.Error("expected ComboGenerationRequested to be raised on construction")
	}
}

func TestStyleSession_NilPreferences_QuickGenerateTrue(t *testing.T) {
	// TC-301-4
	d := newCapturingDispatcher()
	s := newSession(d, nil)
	if !s.QuickGenerate() {
		t.Error("expected QuickGenerate=true when preferences are nil")
	}
}

func TestStyleSession_EmptyPreferences_QuickGenerateTrue(t *testing.T) {
	d := newCapturingDispatcher()
	empty := &valueobjects.StylePreferences{}
	s := newSession(d, empty)
	if !s.QuickGenerate() {
		t.Error("expected QuickGenerate=true when preferences are empty")
	}
}

func TestStyleSession_NonEmptyPreferences_QuickGenerateFalse(t *testing.T) {
	d := newCapturingDispatcher()
	prefs := &valueobjects.StylePreferences{
		Occasions: []valueobjects.Occasion{valueobjects.OccasionCasual},
	}
	s := newSession(d, prefs)
	if s.QuickGenerate() {
		t.Error("expected QuickGenerate=false when preferences are set")
	}
}

func TestStyleSession_LoadWishlist_RaisesWishlistFetchCompleted(t *testing.T) {
	d := newCapturingDispatcher()
	s := newSession(d, nil)
	d.dispatched = nil // reset after construction event

	s.LoadWishlist(twoInStockItems())
	if !d.hasEvent(events.EventTypeWishlistFetchCompleted) {
		t.Error("expected WishlistFetchCompleted to be raised after LoadWishlist")
	}
}

func TestStyleSession_LoadWishlist_StoresSnapshot(t *testing.T) {
	d := newCapturingDispatcher()
	s := newSession(d, nil)
	snap := twoInStockItems()
	s.LoadWishlist(snap)
	if s.Wishlist() == nil {
		t.Fatal("expected Wishlist() to be non-nil after LoadWishlist")
	}
	if len(s.Wishlist().Items) != 2 {
		t.Errorf("expected 2 wishlist items, got %d", len(s.Wishlist().Items))
	}
}

func TestStyleSession_LoadCatalogItems_RaisesCatalogItemsFetched(t *testing.T) {
	d := newCapturingDispatcher()
	s := newSession(d, nil)
	d.dispatched = nil

	s.LoadCatalogItems([]valueobjects.ComboItem{{ConfigSku: "CFG-X"}})
	if !d.hasEvent(events.EventTypeCatalogItemsFetched) {
		t.Error("expected CatalogItemsFetched to be raised after LoadCatalogItems")
	}
}

func TestStyleSession_CompleteCombos_RaisesCombosGenerated(t *testing.T) {
	// TC-401-5
	d := newCapturingDispatcher()
	s := newSession(d, nil)
	s.LoadWishlist(twoInStockItems())
	d.dispatched = nil

	combos := []entities.Combo{
		entities.NewCombo("combo-1", []valueobjects.ComboItem{{}, {}}, 1),
	}
	if err := s.CompleteCombos(combos); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !d.hasEvent(events.EventTypeCombosGenerated) {
		t.Error("expected CombosGenerated to be raised after CompleteCombos")
	}
	if len(s.Combos()) != 1 {
		t.Errorf("expected 1 combo, got %d", len(s.Combos()))
	}
}

func TestStyleSession_CompleteCombos_FiltersExcludedIds(t *testing.T) {
	// TC-405-4
	d := newCapturingDispatcher()
	session := aggregates.NewStyleSession(
		"session-test-id",
		valueobjects.ShopperSession{SessionToken: "tok"},
		nil,
		valueobjects.NewExcludedComboIds([]string{"combo-1"}),
		d,
	)
	session.LoadWishlist(twoInStockItems())
	combos := []entities.Combo{
		entities.NewCombo("combo-1", nil, 1),
		entities.NewCombo("combo-2", nil, 2),
	}
	if err := session.CompleteCombos(combos); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, c := range session.Combos() {
		if c.Id == "combo-1" {
			t.Error("excluded combo-1 should not appear in Combos()")
		}
	}
	if len(session.Combos()) != 1 {
		t.Errorf("expected 1 remaining combo, got %d", len(session.Combos()))
	}
}

func TestStyleSession_CompleteCombos_AllExcluded_IsExhausted(t *testing.T) {
	d := newCapturingDispatcher()
	session := aggregates.NewStyleSession(
		"session-test-id",
		valueobjects.ShopperSession{SessionToken: "tok"},
		nil,
		valueobjects.NewExcludedComboIds([]string{"combo-1", "combo-2"}),
		d,
	)
	session.LoadWishlist(twoInStockItems())
	if err := session.CompleteCombos([]entities.Combo{
		entities.NewCombo("combo-1", nil, 1),
		entities.NewCombo("combo-2", nil, 2),
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !session.IsExhausted() {
		t.Error("expected IsExhausted=true when all combos are excluded")
	}
}

func TestStyleSession_TriggerFallback_RaisesFallbackTriggered(t *testing.T) {
	// TC-404-3
	d := newCapturingDispatcher()
	s := newSession(d, nil)
	d.dispatched = nil

	fb := entities.NewFallbackResult("No combos.", []valueobjects.AlternativeItem{
		{ConfigSku: "CFG-ALT", Reason: "Good replacement"},
	})
	s.TriggerFallback(fb)
	if !d.hasEvent(events.EventTypeFallbackTriggered) {
		t.Error("expected FallbackTriggered to be raised after TriggerFallback")
	}
	if s.Fallback() == nil {
		t.Error("expected Fallback() to be non-nil after TriggerFallback")
	}
}

// ── PreferenceConfirmation ────────────────────────────────────────────────────

// stubInterpretationService is a minimal mock for testing PreferenceConfirmation.
type stubInterpretationService struct {
	returnSummary valueobjects.PreferenceSummary
	returnErr     error
}

func (s *stubInterpretationService) Interpret(prefs valueobjects.StylePreferences) (valueobjects.PreferenceSummary, error) {
	return s.returnSummary, s.returnErr
}

func TestPreferenceConfirmation_Interpret_RaisesEvent(t *testing.T) {
	// TC-302-4
	d := newCapturingDispatcher()
	prefs := valueobjects.StylePreferences{
		Occasions: []valueobjects.Occasion{valueobjects.OccasionCasual},
	}
	pc := aggregates.NewPreferenceConfirmation(prefs, d)
	svc := &stubInterpretationService{
		returnSummary: valueobjects.PreferenceSummary{Text: "Casual look."},
	}
	summary, err := pc.Interpret(svc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.Text != "Casual look." {
		t.Errorf("unexpected summary text: %q", summary.Text)
	}
	if !d.hasEvent(events.EventTypePreferencesInterpreted) {
		t.Error("expected PreferencesInterpreted event to be raised")
	}
}

func TestPreferenceConfirmation_Interpret_PropagatesError(t *testing.T) {
	d := newCapturingDispatcher()
	pc := aggregates.NewPreferenceConfirmation(valueobjects.StylePreferences{}, d)
	svc := &stubInterpretationService{returnErr: fmt.Errorf("service unavailable")}
	_, err := pc.Interpret(svc)
	if err == nil {
		t.Fatal("expected error to be propagated from service, got nil")
	}
}

// stubPreferenceInterpretationService satisfies the services.PreferenceInterpretationService interface
// for the aggregate test above.
var _ services.PreferenceInterpretationService = (*stubInterpretationService)(nil)

// ── Phase 4 gap tests ────────────────────────────────────────────────────────

// TC-DOM-13: CompleteCombos before LoadWishlist — returns error (invariant enforced)
func TestStyleSession_CompleteCombos_BeforeLoadWishlist_ReturnsError(t *testing.T) {
	d := newCapturingDispatcher()
	s := newSession(d, nil)
	// Call CompleteCombos without ever calling LoadWishlist.
	err := s.CompleteCombos([]entities.Combo{
		entities.NewCombo("combo-1", []valueobjects.ComboItem{{}, {}}, 1),
	})
	if err == nil {
		t.Error("expected error when CompleteCombos called before LoadWishlist, got nil")
	}
}

// TC-DOM-14: IsExhausted is reflected in the combos slice being empty
func TestStyleSession_IsExhausted_CombosSliceIsEmpty(t *testing.T) {
	d := newCapturingDispatcher()
	session := aggregates.NewStyleSession(
		"session-exhausted",
		valueobjects.ShopperSession{SessionToken: "tok"},
		nil,
		valueobjects.NewExcludedComboIds([]string{"combo-1", "combo-2"}),
		d,
	)
	session.LoadWishlist(twoInStockItems())
	if err := session.CompleteCombos([]entities.Combo{
		entities.NewCombo("combo-1", nil, 1),
		entities.NewCombo("combo-2", nil, 2),
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !session.IsExhausted() {
		t.Error("expected IsExhausted=true")
	}
	if len(session.Combos()) != 0 {
		t.Errorf("expected empty Combos() when exhausted, got %d", len(session.Combos()))
	}
}

// TC-DOM-15: PreferenceConfirmation.Interpret echoes preferences exactly
func TestPreferenceConfirmation_Interpret_EchoesPreferences(t *testing.T) {
	d := newCapturingDispatcher()
	b, _ := valueobjects.NewBudgetRange(50, 200)
	prefs := valueobjects.StylePreferences{
		Occasions: []valueobjects.Occasion{valueobjects.OccasionCasual, valueobjects.OccasionBeach},
		Styles:    []valueobjects.StyleDirection{valueobjects.StyleMinimalist},
		Budget:    &b,
		FreeText:  "light summer look",
	}
	pc := aggregates.NewPreferenceConfirmation(prefs, d)
	svc := &stubInterpretationService{
		returnSummary: valueobjects.PreferenceSummary{
			Text:        "Casual beach look.",
			Preferences: prefs,
		},
	}
	summary, err := pc.Interpret(svc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(summary.Preferences.Occasions) != 2 {
		t.Errorf("expected 2 occasions echoed, got %d", len(summary.Preferences.Occasions))
	}
	if summary.Preferences.FreeText != "light summer look" {
		t.Errorf("expected FreeText echoed, got %q", summary.Preferences.FreeText)
	}
	if summary.Preferences.Budget == nil || summary.Preferences.Budget.Min != 50 {
		t.Error("expected Budget echoed correctly")
	}
}
