package valueobjects_test

import (
	"ai-styling-engine/domain/valueobjects"
	"testing"
)

// ── BudgetRange ──────────────────────────────────────────────────────────────

func TestNewBudgetRange_Valid(t *testing.T) {
	b, err := valueobjects.NewBudgetRange(50, 200)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if b.Min != 50 || b.Max != 200 {
		t.Errorf("unexpected range: %v", b)
	}
}

func TestNewBudgetRange_MaxNotGreaterThanMin(t *testing.T) {
	_, err := valueobjects.NewBudgetRange(200, 50)
	if err == nil {
		t.Fatal("expected error when max <= min, got nil")
	}
}

func TestNewBudgetRange_MaxEqualMin(t *testing.T) {
	_, err := valueobjects.NewBudgetRange(100, 100)
	if err == nil {
		t.Fatal("expected error when max == min, got nil")
	}
}

func TestNewBudgetRange_NegativeMin(t *testing.T) {
	_, err := valueobjects.NewBudgetRange(-10, 100)
	if err == nil {
		t.Fatal("expected error when min < 0, got nil")
	}
}

func TestBudgetRange_Contains(t *testing.T) {
	b, _ := valueobjects.NewBudgetRange(50, 200)
	tests := []struct {
		price    valueobjects.Money
		expected bool
	}{
		{50, true},
		{125, true},
		{200, true},
		{49, false},
		{201, false},
	}
	for _, tt := range tests {
		if got := b.Contains(tt.price); got != tt.expected {
			t.Errorf("Contains(%v) = %v, want %v", tt.price, got, tt.expected)
		}
	}
}

// ── ColorPalette ─────────────────────────────────────────────────────────────

func TestNewColorPalette_Valid(t *testing.T) {
	_, err := valueobjects.NewColorPalette(
		[]valueobjects.Color{valueobjects.ColorBeige, valueobjects.ColorWhite},
		[]valueobjects.Color{valueobjects.ColorBlack},
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestNewColorPalette_SameColorInBothLists(t *testing.T) {
	// TC-SEC-4
	_, err := valueobjects.NewColorPalette(
		[]valueobjects.Color{valueobjects.ColorBlack},
		[]valueobjects.Color{valueobjects.ColorBlack},
	)
	if err == nil {
		t.Fatal("expected error when a color appears in both lists, got nil")
	}
}

func TestNewColorPalette_EmptyLists(t *testing.T) {
	_, err := valueobjects.NewColorPalette(nil, nil)
	if err != nil {
		t.Fatalf("expected no error for empty lists, got %v", err)
	}
}

// ── StylePreferences ─────────────────────────────────────────────────────────

func TestStylePreferences_IsEmpty_AllZero(t *testing.T) {
	// TC-301-3
	prefs := valueobjects.StylePreferences{}
	if !prefs.IsEmpty() {
		t.Error("expected IsEmpty() = true for zero-value StylePreferences")
	}
}

func TestStylePreferences_IsEmpty_WithOccasion(t *testing.T) {
	// TC-301b-2
	prefs := valueobjects.StylePreferences{
		Occasions: []valueobjects.Occasion{valueobjects.OccasionCasual},
	}
	if prefs.IsEmpty() {
		t.Error("expected IsEmpty() = false when occasions are set")
	}
}

func TestStylePreferences_IsEmpty_WithStyle(t *testing.T) {
	prefs := valueobjects.StylePreferences{
		Styles: []valueobjects.StyleDirection{valueobjects.StyleMinimalist},
	}
	if prefs.IsEmpty() {
		t.Error("expected IsEmpty() = false when styles are set")
	}
}

func TestStylePreferences_IsEmpty_WithFreeTextOnly(t *testing.T) {
	prefs := valueobjects.StylePreferences{FreeText: "something casual"}
	if prefs.IsEmpty() {
		t.Error("expected IsEmpty() = false when freeText is set")
	}
}

func TestStylePreferences_IsEmpty_WithBudget(t *testing.T) {
	b, _ := valueobjects.NewBudgetRange(0, 200)
	prefs := valueobjects.StylePreferences{Budget: &b}
	if prefs.IsEmpty() {
		t.Error("expected IsEmpty() = false when budget is set")
	}
}

// ── ExcludedComboIds ─────────────────────────────────────────────────────────

func TestExcludedComboIds_Contains(t *testing.T) {
	// TC-405-2
	e := valueobjects.NewExcludedComboIds([]string{"combo-a", "combo-b"})
	if !e.Contains("combo-a") {
		t.Error("expected Contains(combo-a) = true")
	}
	if e.Contains("combo-c") {
		t.Error("expected Contains(combo-c) = false")
	}
}

func TestExcludedComboIds_Add_Immutable(t *testing.T) {
	// TC-405-3
	original := valueobjects.NewExcludedComboIds([]string{"combo-a"})
	updated := original.Add("combo-b")

	if original.Contains("combo-b") {
		t.Error("original should not contain combo-b after Add()")
	}
	if !updated.Contains("combo-b") {
		t.Error("updated should contain combo-b")
	}
	if !updated.Contains("combo-a") {
		t.Error("updated should still contain combo-a")
	}
}

func TestExcludedComboIds_EmptyStringsIgnored(t *testing.T) {
	e := valueobjects.NewExcludedComboIds([]string{"", "combo-a", ""})
	if e.Contains("") {
		t.Error("empty string should not be stored as a valid excluded ID")
	}
	if !e.Contains("combo-a") {
		t.Error("combo-a should be present")
	}
}

// ── ComboReasoning ───────────────────────────────────────────────────────────

func TestNewComboReasoning_Valid(t *testing.T) {
	r, err := valueobjects.NewComboReasoning("A great pairing of styles.")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if r.Text == "" {
		t.Error("expected non-empty Text")
	}
}

func TestNewComboReasoning_EmptyText(t *testing.T) {
	// TC-402-2
	_, err := valueobjects.NewComboReasoning("")
	if err == nil {
		t.Fatal("expected error for empty reasoning text, got nil")
	}
}

// ── WishlistSnapshot ─────────────────────────────────────────────────────────

func TestWishlistSnapshot_InStockItems(t *testing.T) {
	snap := valueobjects.WishlistSnapshot{
		Items: []valueobjects.WishlistItem{
			{ItemId: "1", InStock: true},
			{ItemId: "2", InStock: false},
			{ItemId: "3", InStock: true},
		},
		TotalCount: 3,
	}
	inStock := snap.InStockItems()
	if len(inStock) != 2 {
		t.Errorf("expected 2 in-stock items, got %d", len(inStock))
	}
	for _, item := range inStock {
		if !item.InStock {
			t.Error("InStockItems() returned an out-of-stock item")
		}
	}
}

func TestWishlistSnapshot_InStockItems_AllOutOfStock(t *testing.T) {
	snap := valueobjects.WishlistSnapshot{
		Items: []valueobjects.WishlistItem{
			{ItemId: "1", InStock: false},
		},
		TotalCount: 1,
	}
	inStock := snap.InStockItems()
	if len(inStock) != 0 {
		t.Errorf("expected 0 in-stock items, got %d", len(inStock))
	}
}

// ── CatalogSearchFilters ──────────────────────────────────────────────────────

func TestCatalogSearchFiltersFromPreferences_Nil(t *testing.T) {
	f := valueobjects.CatalogSearchFiltersFromPreferences(nil)
	if f.Limit != 20 {
		t.Errorf("expected default Limit=20 for nil preferences, got %d", f.Limit)
	}
	if f.Occasion != nil {
		t.Error("expected nil Occasion for nil preferences")
	}
}

func TestCatalogSearchFiltersFromPreferences_WithOccasion(t *testing.T) {
	occ := valueobjects.OccasionBeach
	prefs := &valueobjects.StylePreferences{
		Occasions: []valueobjects.Occasion{occ},
	}
	f := valueobjects.CatalogSearchFiltersFromPreferences(prefs)
	if f.Occasion == nil || *f.Occasion != occ {
		t.Errorf("expected Occasion=%v, got %v", occ, f.Occasion)
	}
}

// ── Phase 2 gap tests ────────────────────────────────────────────────────────

// TC-DOM-1: NewBudgetRange(0,0) must fail — max not > min
func TestNewBudgetRange_ZeroZero_ReturnsError(t *testing.T) {
	_, err := valueobjects.NewBudgetRange(0, 0)
	if err == nil {
		t.Fatal("expected error for NewBudgetRange(0,0), got nil")
	}
}

// TC-DOM-2: NewBudgetRange(0,1) must succeed — min=0 is valid
func TestNewBudgetRange_ZeroMin_Succeeds(t *testing.T) {
	b, err := valueobjects.NewBudgetRange(0, 1)
	if err != nil {
		t.Fatalf("expected no error for NewBudgetRange(0,1), got %v", err)
	}
	if b.Min != 0 || b.Max != 1 {
		t.Errorf("unexpected range: min=%v max=%v", b.Min, b.Max)
	}
}

// TC-DOM-3: Multiple colors in both lists with one overlap → error
func TestNewColorPalette_MultipleColorsOneOverlap_ReturnsError(t *testing.T) {
	_, err := valueobjects.NewColorPalette(
		[]valueobjects.Color{valueobjects.ColorBeige, valueobjects.ColorNavy},
		[]valueobjects.Color{valueobjects.ColorWhite, valueobjects.ColorNavy},
	)
	if err == nil {
		t.Fatal("expected error when one color appears in both lists, got nil")
	}
}

// TC-DOM-4: Same color twice in preferred only → no error
func TestNewColorPalette_DuplicateInPreferredOnly_NoError(t *testing.T) {
	_, err := valueobjects.NewColorPalette(
		[]valueobjects.Color{valueobjects.ColorBlack, valueobjects.ColorBlack},
		[]valueobjects.Color{valueobjects.ColorWhite},
	)
	if err != nil {
		t.Fatalf("expected no error for duplicate in preferred list only, got %v", err)
	}
}

// TC-DOM-5: IsEmpty returns false when only Colors.Preferred is set
func TestStylePreferences_IsEmpty_WithPreferredColorOnly_False(t *testing.T) {
	palette, _ := valueobjects.NewColorPalette(
		[]valueobjects.Color{valueobjects.ColorBeige}, nil,
	)
	prefs := valueobjects.StylePreferences{Colors: palette}
	if prefs.IsEmpty() {
		t.Error("expected IsEmpty()=false when Colors.Preferred is set")
	}
}

// TC-DOM-6: IsEmpty returns false when only Colors.Excluded is set
func TestStylePreferences_IsEmpty_WithExcludedColorOnly_False(t *testing.T) {
	palette, _ := valueobjects.NewColorPalette(
		nil, []valueobjects.Color{valueobjects.ColorBlack},
	)
	prefs := valueobjects.StylePreferences{Colors: palette}
	if prefs.IsEmpty() {
		t.Error("expected IsEmpty()=false when Colors.Excluded is set")
	}
}

// TC-DOM-7: NewComboReasoning with whitespace-only text returns error
func TestNewComboReasoning_WhitespaceOnly_ReturnsError(t *testing.T) {
	_, err := valueobjects.NewComboReasoning("   ")
	if err == nil {
		t.Fatal("expected error for whitespace-only reasoning text, got nil")
	}
}

// TC-DOM-8: InStockItems on empty snapshot returns empty slice without panic
func TestWishlistSnapshot_InStockItems_EmptySnapshot_ReturnsEmptySlice(t *testing.T) {
	snap := valueobjects.WishlistSnapshot{}
	result := snap.InStockItems()
	if result == nil {
		// nil is acceptable — just must not panic
	}
	if len(result) != 0 {
		t.Errorf("expected 0 in-stock items for empty snapshot, got %d", len(result))
	}
}

// TC-DOM-9: CatalogSearchFiltersFromPreferences with budget → PriceRange populated
func TestCatalogSearchFiltersFromPreferences_WithBudget_PriceRangeSet(t *testing.T) {
	b, _ := valueobjects.NewBudgetRange(50, 200)
	prefs := &valueobjects.StylePreferences{Budget: &b}
	f := valueobjects.CatalogSearchFiltersFromPreferences(prefs)
	if f.PriceRange == nil {
		t.Fatal("expected PriceRange to be set when budget is provided")
	}
	if f.PriceRange.Min != 50 || f.PriceRange.Max != 200 {
		t.Errorf("unexpected PriceRange: %+v", f.PriceRange)
	}
}

// TC-DOM-10: CatalogSearchFiltersFromPreferences with preferred colors → Colors populated
func TestCatalogSearchFiltersFromPreferences_WithPreferredColors_ColorsSet(t *testing.T) {
	palette, _ := valueobjects.NewColorPalette(
		[]valueobjects.Color{valueobjects.ColorBeige, valueobjects.ColorWhite}, nil,
	)
	prefs := &valueobjects.StylePreferences{Colors: palette}
	f := valueobjects.CatalogSearchFiltersFromPreferences(prefs)
	if len(f.Colors) != 2 {
		t.Errorf("expected 2 colors in filters, got %d", len(f.Colors))
	}
}
