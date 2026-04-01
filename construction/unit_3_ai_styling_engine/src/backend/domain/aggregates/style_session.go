package aggregates

import (
	"ai-styling-engine/domain/entities"
	"ai-styling-engine/domain/events"
	"ai-styling-engine/domain/valueobjects"
	"fmt"
)

// StyleSession is the core transient aggregate for the combo generation pipeline.
// It drives state transitions and raises domain events at each stage.
type StyleSession struct {
	id            valueobjects.StyleSessionId
	shopperSession valueobjects.ShopperSession
	preferences   *valueobjects.StylePreferences
	excludedIds   valueobjects.ExcludedComboIds
	wishlist      *valueobjects.WishlistSnapshot
	catalogItems  []valueobjects.ComboItem
	combos        []entities.Combo
	fallback      *entities.FallbackResult
	quickGenerate bool
	exhausted     bool
	needsCatalog  bool
	dispatcher    events.EventDispatcher
}

func NewStyleSession(
	id valueobjects.StyleSessionId,
	session valueobjects.ShopperSession,
	preferences *valueobjects.StylePreferences,
	excludedIds valueobjects.ExcludedComboIds,
	dispatcher events.EventDispatcher,
) *StyleSession {
	s := &StyleSession{
		id:             id,
		shopperSession: session,
		preferences:    preferences,
		excludedIds:    excludedIds,
		dispatcher:     dispatcher,
	}
	if preferences == nil || preferences.IsEmpty() {
		s.quickGenerate = true
	}
	dispatcher.Dispatch(events.ComboGenerationRequested{
		SessionId:     id,
		Preferences:   preferences,
		ExcludedIds:   excludedIds,
		QuickGenerate: s.quickGenerate,
	})
	return s
}

func (s *StyleSession) Id() valueobjects.StyleSessionId            { return s.id }
func (s *StyleSession) ShopperSession() valueobjects.ShopperSession { return s.shopperSession }
func (s *StyleSession) Preferences() *valueobjects.StylePreferences { return s.preferences }
func (s *StyleSession) ExcludedIds() valueobjects.ExcludedComboIds  { return s.excludedIds }
func (s *StyleSession) Wishlist() *valueobjects.WishlistSnapshot    { return s.wishlist }
func (s *StyleSession) CatalogItems() []valueobjects.ComboItem      { return s.catalogItems }
func (s *StyleSession) Combos() []entities.Combo                    { return s.combos }
func (s *StyleSession) Fallback() *entities.FallbackResult          { return s.fallback }
func (s *StyleSession) QuickGenerate() bool                         { return s.quickGenerate }
func (s *StyleSession) IsExhausted() bool                           { return s.exhausted }
func (s *StyleSession) NeedsCatalog() bool                          { return s.needsCatalog }

// LoadWishlist attaches the fetched wishlist snapshot and raises WishlistFetchCompleted.
// The WishlistSupplementationPolicy responds to this event and may raise CatalogSupplementationRequired.
func (s *StyleSession) LoadWishlist(snapshot valueobjects.WishlistSnapshot) {
	s.wishlist = &snapshot
	s.dispatcher.Dispatch(events.WishlistFetchCompleted{
		SessionId: s.id,
		Snapshot:  snapshot,
	})
}

// MarkNeedsCatalog is called by WishlistSupplementationPolicy when supplementation is required.
func (s *StyleSession) MarkNeedsCatalog() {
	s.needsCatalog = true
}

// LoadCatalogItems stores supplementary catalog and Complete-the-Look items, then raises CatalogItemsFetched.
func (s *StyleSession) LoadCatalogItems(items []valueobjects.ComboItem) {
	s.catalogItems = items
	s.dispatcher.Dispatch(events.CatalogItemsFetched{
		SessionId:    s.id,
		CatalogItems: items,
	})
}

// CompleteCombos stores the final combo set after exclusion filtering and raises CombosGenerated.
// Returns an error if the wishlist has not been loaded yet (invariant: wishlist must be loaded first).
func (s *StyleSession) CompleteCombos(combos []entities.Combo) error {
	if s.wishlist == nil {
		return fmt.Errorf("wishlist must be loaded before completing combos")
	}
	// Filter out excluded combo IDs.
	var filtered []entities.Combo
	for _, c := range combos {
		if !s.excludedIds.Contains(c.Id) {
			filtered = append(filtered, c)
		}
	}
	if len(filtered) == 0 {
		s.exhausted = true
	}
	s.combos = filtered
	s.dispatcher.Dispatch(events.CombosGenerated{
		SessionId:  s.id,
		ComboCount: len(filtered),
	})
	return nil
}

// TriggerFallback stores the fallback result and raises FallbackTriggered.
func (s *StyleSession) TriggerFallback(fallback entities.FallbackResult) {
	s.fallback = &fallback
	s.dispatcher.Dispatch(events.FallbackTriggered{
		SessionId:    s.id,
		Alternatives: fallback.Alternatives,
	})
}
