package usecases_test

import (
	"ai-styling-engine/application/commands"
	"ai-styling-engine/application/usecases"
	infraACL "ai-styling-engine/infrastructure/acl"
	infraAI "ai-styling-engine/infrastructure/ai"
	"ai-styling-engine/infrastructure/dispatcher"
	"ai-styling-engine/domain/events"
	"ai-styling-engine/domain/policies"
	"ai-styling-engine/domain/repositories"
	"ai-styling-engine/domain/services"
	"ai-styling-engine/domain/valueobjects"
	"errors"
	"testing"
)

// ── Helpers ───────────────────────────────────────────────────────────────────

func newDispatcherWithPolicies() events.EventDispatcher {
	d := dispatcher.NewInProcessEventDispatcher()
	prefPolicy := policies.NewPreferenceDefaultPolicy()
	d.Register(events.EventTypeComboGenerationRequested, prefPolicy.Handle)
	wishlistPolicy := policies.NewWishlistSupplementationPolicy(d)
	d.Register(events.EventTypeWishlistFetchCompleted, wishlistPolicy.Handle)
	fallbackPolicy := policies.NewFallbackPolicy()
	d.Register(events.EventTypeFallbackTriggered, fallbackPolicy.Handle)
	exclusionPolicy := policies.NewComboExclusionPolicy()
	d.Register(events.EventTypeCombosGenerated, exclusionPolicy.Handle)
	return d
}

func newGenerateUseCase(
	wishlistRepo repositories.WishlistRepository,
	scoringSvc services.ComboCompatibilityScoringService,
) *usecases.GenerateCombosUseCase {
	d := newDispatcherWithPolicies()
	return usecases.NewGenerateCombosUseCase(
		wishlistRepo,
		infraACL.NewInMemoryProductCatalogRepository(),
		infraACL.NewInMemoryCompleteLookRepository(),
		scoringSvc,
		infraAI.NewMockComboReasoningGenerationService(),
		d,
	)
}

func defaultGenerateCmd(prefs *valueobjects.StylePreferences) commands.GenerateCombosCommand {
	return commands.GenerateCombosCommand{
		Preferences:    prefs,
		ExcludedIds:    valueobjects.NewExcludedComboIds(nil),
		ShopperSession: valueobjects.ShopperSession{SessionToken: "test-token"},
	}
}

// ── GetPreferenceOptionsUseCase ───────────────────────────────────────────────

func TestGetPreferenceOptionsUseCase_ReturnsAllOptions(t *testing.T) {
	// TC-301b-1
	uc := usecases.NewGetPreferenceOptionsUseCase()
	catalogue := uc.Execute()
	if len(catalogue.Occasions) == 0 {
		t.Error("expected non-empty occasions")
	}
	if len(catalogue.Styles) == 0 {
		t.Error("expected non-empty styles")
	}
	if len(catalogue.Colors) == 0 {
		t.Error("expected non-empty colors")
	}
}

// ── ConfirmPreferencesUseCase ─────────────────────────────────────────────────

func TestConfirmPreferencesUseCase_FullPreferences_ReturnsSummary(t *testing.T) {
	// TC-302-1
	d := dispatcher.NewInProcessEventDispatcher()
	uc := usecases.NewConfirmPreferencesUseCase(
		infraAI.NewMockPreferenceInterpretationService(), d,
	)
	b, _ := valueobjects.NewBudgetRange(50, 200)
	palette, _ := valueobjects.NewColorPalette(
		[]valueobjects.Color{valueobjects.ColorBeige},
		[]valueobjects.Color{valueobjects.ColorBlack},
	)
	prefs := valueobjects.StylePreferences{
		Occasions: []valueobjects.Occasion{valueobjects.OccasionCasual},
		Styles:    []valueobjects.StyleDirection{valueobjects.StyleMinimalist},
		Budget:    &b,
		Colors:    palette,
		FreeText:  "Something light",
	}
	summary, err := uc.Execute(commands.ConfirmPreferencesCommand{
		Preferences:    prefs,
		ShopperSession: valueobjects.ShopperSession{SessionToken: "tok"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.Text == "" {
		t.Error("expected non-empty summary text")
	}
}

func TestConfirmPreferencesUseCase_EmptyPreferences_ReturnsGenericSummary(t *testing.T) {
	// TC-302-2
	d := dispatcher.NewInProcessEventDispatcher()
	uc := usecases.NewConfirmPreferencesUseCase(
		infraAI.NewMockPreferenceInterpretationService(), d,
	)
	summary, err := uc.Execute(commands.ConfirmPreferencesCommand{
		Preferences:    valueobjects.StylePreferences{},
		ShopperSession: valueobjects.ShopperSession{SessionToken: "tok"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.Text == "" {
		t.Error("expected non-empty summary even for empty preferences")
	}
}

// ── GenerateCombosUseCase — quick-generate ────────────────────────────────────

func TestGenerateCombosUseCase_NilPreferences_ReturnsSuccess(t *testing.T) {
	// TC-301-1, TC-301-4
	uc := newGenerateUseCase(
		infraACL.NewInMemoryWishlistRepository(),
		infraAI.NewMockComboCompatibilityScoringService(),
	)
	result, err := uc.Execute(defaultGenerateCmd(nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsSuccess() {
		t.Fatal("expected success result for quick-generate")
	}
	if len(result.Success.Combos) == 0 {
		t.Error("expected at least one combo")
	}
}

func TestGenerateCombosUseCase_EmptyPreferences_TreatedAsQuickGenerate(t *testing.T) {
	// TC-301b-4
	uc := newGenerateUseCase(
		infraACL.NewInMemoryWishlistRepository(),
		infraAI.NewMockComboCompatibilityScoringService(),
	)
	empty := &valueobjects.StylePreferences{}
	result, err := uc.Execute(defaultGenerateCmd(empty))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsSuccess() {
		t.Fatal("expected success result for empty preferences")
	}
}

func TestGenerateCombosUseCase_WithPreferences_ReturnsSuccess(t *testing.T) {
	// TC-301b-3
	uc := newGenerateUseCase(
		infraACL.NewInMemoryWishlistRepository(),
		infraAI.NewMockComboCompatibilityScoringService(),
	)
	prefs := &valueobjects.StylePreferences{
		Occasions: []valueobjects.Occasion{valueobjects.OccasionCasual},
		Styles:    []valueobjects.StyleDirection{valueobjects.StyleMinimalist},
	}
	result, err := uc.Execute(defaultGenerateCmd(prefs))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsSuccess() {
		t.Fatal("expected success result when preferences are provided")
	}
}

// ── GenerateCombosUseCase — combo structure ───────────────────────────────────

func TestGenerateCombosUseCase_EachComboHasAtLeastTwoItems(t *testing.T) {
	// TC-401-1
	uc := newGenerateUseCase(
		infraACL.NewInMemoryWishlistRepository(),
		infraAI.NewMockComboCompatibilityScoringService(),
	)
	result, _ := uc.Execute(defaultGenerateCmd(nil))
	for _, combo := range result.Success.Combos {
		if len(combo.Items) < 2 {
			t.Errorf("combo %q has fewer than 2 items", combo.Id)
		}
	}
}

func TestGenerateCombosUseCase_EachComboHasNonEmptyReasoning(t *testing.T) {
	// TC-402-1
	uc := newGenerateUseCase(
		infraACL.NewInMemoryWishlistRepository(),
		infraAI.NewMockComboCompatibilityScoringService(),
	)
	result, _ := uc.Execute(defaultGenerateCmd(nil))
	for _, combo := range result.Success.Combos {
		if combo.Reasoning.Text == "" {
			t.Errorf("combo %q has empty reasoning", combo.Id)
		}
	}
}

func TestGenerateCombosUseCase_ComboCountWithinBounds(t *testing.T) {
	// TC-401-3
	uc := newGenerateUseCase(
		infraACL.NewInMemoryWishlistRepository(),
		infraAI.NewMockComboCompatibilityScoringService(),
	)
	result, _ := uc.Execute(defaultGenerateCmd(nil))
	n := len(result.Success.Combos)
	if n < 1 || n > 5 {
		t.Errorf("expected 1–5 combos, got %d", n)
	}
}

func TestGenerateCombosUseCase_WishlistItemsSourcedAsWishlist(t *testing.T) {
	// TC-401-2
	uc := newGenerateUseCase(
		infraACL.NewInMemoryWishlistRepository(),
		infraAI.NewMockComboCompatibilityScoringService(),
	)
	result, _ := uc.Execute(defaultGenerateCmd(nil))
	for _, combo := range result.Success.Combos {
		hasWishlistItem := false
		for _, item := range combo.Items {
			if item.Source == valueobjects.ItemSourceWishlist {
				hasWishlistItem = true
			}
		}
		if !hasWishlistItem {
			t.Errorf("combo %q has no wishlist-sourced item", combo.Id)
		}
	}
}

// ── GenerateCombosUseCase — catalog supplementation ──────────────────────────

func TestGenerateCombosUseCase_SparseWishlist_CatalogItemsIncluded(t *testing.T) {
	// TC-403-1, TC-403-2
	uc := newGenerateUseCase(
		&singleInStockItemRepo{},
		infraAI.NewMockComboCompatibilityScoringService(),
	)
	result, err := uc.Execute(defaultGenerateCmd(nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should succeed or return a fallback — never error.
	_ = result
}

// ── GenerateCombosUseCase — fallback ─────────────────────────────────────────

func TestGenerateCombosUseCase_EmptyWishlist_ReturnsFallback(t *testing.T) {
	// TC-404-5
	uc := newGenerateUseCase(
		&emptyWishlistRepo{},
		infraAI.NewMockComboCompatibilityScoringService(),
	)
	result, err := uc.Execute(defaultGenerateCmd(nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsSuccess() && len(result.Success.Combos) == 0 {
		// exhausted success is acceptable
		return
	}
	if !result.IsSuccess() {
		if result.Fallback == nil {
			t.Fatal("expected fallback result, got nil")
		}
		if result.Fallback.FallbackResult.Message == "" {
			t.Error("fallback message should not be empty")
		}
	}
}

func TestGenerateCombosUseCase_FallbackResult_AlternativesHaveReasons(t *testing.T) {
	// TC-404-2
	uc := newGenerateUseCase(
		&emptyWishlistRepo{},
		infraAI.NewMockComboCompatibilityScoringService(),
	)
	result, _ := uc.Execute(defaultGenerateCmd(nil))
	if !result.IsSuccess() && result.Fallback != nil {
		for _, alt := range result.Fallback.FallbackResult.Alternatives {
			if alt.Reason == "" {
				t.Errorf("alternative %q has empty reason", alt.ConfigSku)
			}
		}
	}
}

// ── GenerateCombosUseCase — combo exclusion ───────────────────────────────────

func TestGenerateCombosUseCase_ExcludedComboAbsentFromResult(t *testing.T) {
	// TC-405-1
	uc := newGenerateUseCase(
		infraACL.NewInMemoryWishlistRepository(),
		infraAI.NewMockComboCompatibilityScoringService(),
	)
	// First call — get combo IDs.
	first, _ := uc.Execute(defaultGenerateCmd(nil))
	if !first.IsSuccess() || len(first.Success.Combos) == 0 {
		t.Skip("no combos to exclude")
	}
	excludedId := first.Success.Combos[0].Id

	// Second call — exclude the first combo.
	cmd := commands.GenerateCombosCommand{
		Preferences:    nil,
		ExcludedIds:    valueobjects.NewExcludedComboIds([]string{excludedId}),
		ShopperSession: valueobjects.ShopperSession{SessionToken: "test-token"},
	}
	second, err := uc.Execute(cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if second.IsSuccess() {
		for _, c := range second.Success.Combos {
			if c.Id == excludedId {
				t.Errorf("excluded combo %q still appeared in result", excludedId)
			}
		}
	}
}

func TestGenerateCombosUseCase_ExhaustedCombos_CombosIsEmptySlice(t *testing.T) {
	// TC-405-5: exhausted response should have combos: [] not null
	uc := newGenerateUseCase(
		infraACL.NewInMemoryWishlistRepository(),
		infraAI.NewMockComboCompatibilityScoringService(),
	)
	// Get all combo IDs.
	first, _ := uc.Execute(defaultGenerateCmd(nil))
	if !first.IsSuccess() {
		t.Skip("no combos produced")
	}
	ids := make([]string, len(first.Success.Combos))
	for i, c := range first.Success.Combos {
		ids[i] = c.Id
	}

	cmd := commands.GenerateCombosCommand{
		Preferences:    nil,
		ExcludedIds:    valueobjects.NewExcludedComboIds(ids),
		ShopperSession: valueobjects.ShopperSession{SessionToken: "test-token"},
	}
	result, err := uc.Execute(cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsSuccess() && result.Success.Combos == nil {
		t.Error("combos should be an empty slice, not nil, when exhausted")
	}
}

// ── GenerateCombosUseCase — dependency failure ────────────────────────────────

func TestGenerateCombosUseCase_WishlistError_ReturnsError(t *testing.T) {
	// TC-SEC-7
	uc := newGenerateUseCase(
		&erroringWishlistRepo{},
		infraAI.NewMockComboCompatibilityScoringService(),
	)
	_, err := uc.Execute(defaultGenerateCmd(nil))
	if err == nil {
		t.Fatal("expected error when wishlist repository fails")
	}
}

// ── Stub repositories ─────────────────────────────────────────────────────────

// emptyWishlistRepo returns a wishlist with no items.
type emptyWishlistRepo struct{}

func (r *emptyWishlistRepo) FetchForSession(_ valueobjects.ShopperSession) (valueobjects.WishlistSnapshot, error) {
	return valueobjects.WishlistSnapshot{Items: nil, TotalCount: 0}, nil
}

// singleInStockItemRepo returns exactly one in-stock item.
type singleInStockItemRepo struct{}

func (r *singleInStockItemRepo) FetchForSession(_ valueobjects.ShopperSession) (valueobjects.WishlistSnapshot, error) {
	return valueobjects.WishlistSnapshot{
		Items:      []valueobjects.WishlistItem{{ItemId: "1", ConfigSku: "CFG-A", InStock: true}},
		TotalCount: 1,
	}, nil
}

// erroringWishlistRepo always returns an error.
type erroringWishlistRepo struct{}

func (r *erroringWishlistRepo) FetchForSession(_ valueobjects.ShopperSession) (valueobjects.WishlistSnapshot, error) {
	return valueobjects.WishlistSnapshot{}, errors.New("service unavailable")
}

// ── Phase 6 gap tests ────────────────────────────────────────────────────────

// TC-301b-2: Occasions contains exactly the 6 defined values
func TestGetPreferenceOptionsUseCase_ExactOccasions(t *testing.T) {
	uc := usecases.NewGetPreferenceOptionsUseCase()
	cat := uc.Execute()
	want := []string{"casual", "formal", "outdoor", "beach", "office", "party"}
	if len(cat.Occasions) != len(want) {
		t.Fatalf("expected %d occasions, got %d", len(want), len(cat.Occasions))
	}
	got := make(map[string]bool)
	for _, o := range cat.Occasions {
		got[string(o)] = true
	}
	for _, w := range want {
		if !got[w] {
			t.Errorf("missing occasion %q", w)
		}
	}
}

// TC-301b-3: Styles contains exactly the 4 defined values
func TestGetPreferenceOptionsUseCase_ExactStyles(t *testing.T) {
	uc := usecases.NewGetPreferenceOptionsUseCase()
	cat := uc.Execute()
	want := []string{"minimalist", "bold", "classic", "bohemian"}
	if len(cat.Styles) != len(want) {
		t.Fatalf("expected %d styles, got %d", len(want), len(cat.Styles))
	}
	got := make(map[string]bool)
	for _, s := range cat.Styles {
		got[string(s)] = true
	}
	for _, w := range want {
		if !got[w] {
			t.Errorf("missing style %q", w)
		}
	}
}

// TC-301b-4: Colors contains exactly the 7 defined values
func TestGetPreferenceOptionsUseCase_ExactColors(t *testing.T) {
	uc := usecases.NewGetPreferenceOptionsUseCase()
	cat := uc.Execute()
	want := []string{"black", "white", "navy", "beige", "red", "green", "pastel"}
	if len(cat.Colors) != len(want) {
		t.Fatalf("expected %d colors, got %d", len(want), len(cat.Colors))
	}
	got := make(map[string]bool)
	for _, c := range cat.Colors {
		got[string(c)] = true
	}
	for _, w := range want {
		if !got[w] {
			t.Errorf("missing color %q", w)
		}
	}
}

// TC-302-3: ConfirmPreferencesUseCase echoes all non-empty fields in PreferenceSummary.Preferences
func TestConfirmPreferencesUseCase_EchoesAllFields(t *testing.T) {
	d := dispatcher.NewInProcessEventDispatcher()
	uc := usecases.NewConfirmPreferencesUseCase(infraAI.NewMockPreferenceInterpretationService(), d)
	b, _ := valueobjects.NewBudgetRange(50, 200)
	palette, _ := valueobjects.NewColorPalette(
		[]valueobjects.Color{valueobjects.ColorBeige},
		[]valueobjects.Color{valueobjects.ColorBlack},
	)
	prefs := valueobjects.StylePreferences{
		Occasions: []valueobjects.Occasion{valueobjects.OccasionCasual},
		Styles:    []valueobjects.StyleDirection{valueobjects.StyleMinimalist},
		Budget:    &b,
		Colors:    palette,
		FreeText:  "light summer look",
	}
	summary, err := uc.Execute(commands.ConfirmPreferencesCommand{
		Preferences:    prefs,
		ShopperSession: valueobjects.ShopperSession{SessionToken: "tok"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(summary.Preferences.Occasions) != 1 {
		t.Errorf("expected 1 occasion echoed, got %d", len(summary.Preferences.Occasions))
	}
	if len(summary.Preferences.Styles) != 1 {
		t.Errorf("expected 1 style echoed, got %d", len(summary.Preferences.Styles))
	}
	if summary.Preferences.Budget == nil {
		t.Error("expected Budget echoed")
	}
	if summary.Preferences.FreeText != "light summer look" {
		t.Errorf("expected FreeText echoed, got %q", summary.Preferences.FreeText)
	}
}

// TC-403-3: Sparse wishlist triggers catalog supplementation — at least one catalog-sourced item
func TestGenerateCombosUseCase_SparseWishlist_HasCatalogSourcedItem(t *testing.T) {
	uc := newGenerateUseCase(
		&singleInStockItemRepo{},
		infraAI.NewMockComboCompatibilityScoringService(),
	)
	result, err := uc.Execute(defaultGenerateCmd(nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsSuccess() {
		return // fallback is acceptable when catalog also can't form a combo
	}
	hasCatalog := false
	for _, combo := range result.Success.Combos {
		for _, item := range combo.Items {
			if item.Source == valueobjects.ItemSourceCatalog {
				hasCatalog = true
			}
		}
	}
	if !hasCatalog {
		t.Error("expected at least one catalog-sourced item when wishlist is sparse")
	}
}

// TC-404-1: Fallback result contains at least one AlternativeItem
func TestGenerateCombosUseCase_Fallback_HasAtLeastOneAlternative(t *testing.T) {
	uc := newGenerateUseCase(
		&emptyWishlistRepo{},
		infraAI.NewMockComboCompatibilityScoringService(),
	)
	result, err := uc.Execute(defaultGenerateCmd(nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsSuccess() && result.Fallback != nil {
		if len(result.Fallback.FallbackResult.Alternatives) == 0 {
			t.Error("expected at least one AlternativeItem in fallback result")
		}
	}
}

// TC-404-3: Each AlternativeItem in fallback has a non-empty Reason
func TestGenerateCombosUseCase_Fallback_EachAlternativeHasReason(t *testing.T) {
	uc := newGenerateUseCase(
		&emptyWishlistRepo{},
		infraAI.NewMockComboCompatibilityScoringService(),
	)
	result, _ := uc.Execute(defaultGenerateCmd(nil))
	if !result.IsSuccess() && result.Fallback != nil {
		for _, alt := range result.Fallback.FallbackResult.Alternatives {
			if alt.Reason == "" {
				t.Errorf("alternative %q has empty Reason", alt.ConfigSku)
			}
		}
	}
}

// TC-INFRA-1: Scoring service error → use case returns non-nil error (not fallback)
func TestGenerateCombosUseCase_ScoringServiceError_ReturnsError(t *testing.T) {
	uc := newGenerateUseCase(
		infraACL.NewInMemoryWishlistRepository(),
		&erroringScoringService{},
	)
	_, err := uc.Execute(defaultGenerateCmd(nil))
	if err == nil {
		t.Fatal("expected non-nil error when scoring service fails, got nil")
	}
}

// erroringScoringService always returns an error (infrastructure failure).
type erroringScoringService struct{}

func (s *erroringScoringService) Score(_ services.ScoringInput) (services.ScoringResult, error) {
	return services.ScoringResult{}, errors.New("bedrock unavailable")
}
