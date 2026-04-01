package ai_test

import (
	infraAI "ai-styling-engine/infrastructure/ai"
	"ai-styling-engine/domain/services"
	"ai-styling-engine/domain/valueobjects"
	"strings"
	"testing"
)

// ── MockComboCompatibilityScoringService ──────────────────────────────────────

func TestMockScoringService_TwoInStockItems_ReturnsCombos(t *testing.T) {
	svc := infraAI.NewMockComboCompatibilityScoringService()
	input := services.ScoringInput{
		WishlistItems: []valueobjects.WishlistItem{
			{ConfigSku: "CFG-A", SimpleSku: "SKU-A", Name: "Blazer", InStock: true},
			{ConfigSku: "CFG-B", SimpleSku: "SKU-B", Name: "Trousers", InStock: true},
		},
		ExcludedComboIds: valueobjects.NewExcludedComboIds(nil),
	}
	result, err := svc.Score(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsFallback() {
		t.Fatal("expected combos result, got fallback")
	}
	if len(result.Candidates) == 0 {
		t.Error("expected at least one combo candidate")
	}
}

func TestMockScoringService_EachCandidateHasAtLeastTwoItems(t *testing.T) {
	// TC-401-1
	svc := infraAI.NewMockComboCompatibilityScoringService()
	input := services.ScoringInput{
		WishlistItems: []valueobjects.WishlistItem{
			{ConfigSku: "CFG-A", SimpleSku: "SKU-A", InStock: true},
			{ConfigSku: "CFG-B", SimpleSku: "SKU-B", InStock: true},
		},
		ExcludedComboIds: valueobjects.NewExcludedComboIds(nil),
	}
	result, _ := svc.Score(input)
	for _, c := range result.Candidates {
		if len(c.Items) < 2 {
			t.Errorf("combo %q has fewer than 2 items: %d", c.Id, len(c.Items))
		}
	}
}

func TestMockScoringService_WishlistItemsHaveWishlistSource(t *testing.T) {
	// TC-401-2
	svc := infraAI.NewMockComboCompatibilityScoringService()
	input := services.ScoringInput{
		WishlistItems: []valueobjects.WishlistItem{
			{ConfigSku: "CFG-A", SimpleSku: "SKU-A", Name: "Blazer", InStock: true},
			{ConfigSku: "CFG-B", SimpleSku: "SKU-B", Name: "Trousers", InStock: true},
		},
		ExcludedComboIds: valueobjects.NewExcludedComboIds(nil),
	}
	result, _ := svc.Score(input)
	for _, c := range result.Candidates {
		for _, item := range c.Items {
			if item.Source != valueobjects.ItemSourceWishlist && item.Source != valueobjects.ItemSourceCatalog {
				t.Errorf("item has unexpected source: %q", item.Source)
			}
		}
	}
}

func TestMockScoringService_OnlyOneInStockItem_ReturnsFallback(t *testing.T) {
	svc := infraAI.NewMockComboCompatibilityScoringService()
	input := services.ScoringInput{
		WishlistItems: []valueobjects.WishlistItem{
			{ConfigSku: "CFG-A", InStock: true},
		},
		ExcludedComboIds: valueobjects.NewExcludedComboIds(nil),
	}
	result, err := svc.Score(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsFallback() {
		t.Fatal("expected fallback when only one in-stock item with no supplementary items")
	}
	if result.Fallback.Message == "" {
		t.Error("fallback message should not be empty")
	}
}

func TestMockScoringService_AllExcluded_ReturnsFallback(t *testing.T) {
	svc := infraAI.NewMockComboCompatibilityScoringService()
	// Score once to find the IDs.
	input := services.ScoringInput{
		WishlistItems: []valueobjects.WishlistItem{
			{ConfigSku: "CFG-A", SimpleSku: "SKU-A", InStock: true},
			{ConfigSku: "CFG-B", SimpleSku: "SKU-B", InStock: true},
		},
		ExcludedComboIds: valueobjects.NewExcludedComboIds(nil),
	}
	first, _ := svc.Score(input)
	ids := make([]string, len(first.Candidates))
	for i, c := range first.Candidates {
		ids[i] = c.Id
	}

	// Score again with all IDs excluded.
	input.ExcludedComboIds = valueobjects.NewExcludedComboIds(ids)
	result, err := svc.Score(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsFallback() {
		t.Fatal("expected fallback when all combos are excluded")
	}
}

func TestMockScoringService_CandidateCountAtMostFive(t *testing.T) {
	// TC-401-3
	svc := infraAI.NewMockComboCompatibilityScoringService()
	input := services.ScoringInput{
		WishlistItems: []valueobjects.WishlistItem{
			{ConfigSku: "CFG-A", SimpleSku: "SKU-A", InStock: true},
			{ConfigSku: "CFG-B", SimpleSku: "SKU-B", InStock: true},
		},
		ExcludedComboIds: valueobjects.NewExcludedComboIds(nil),
	}
	result, _ := svc.Score(input)
	if len(result.Candidates) > 5 {
		t.Errorf("expected at most 5 combos, got %d", len(result.Candidates))
	}
}

// ── MockComboReasoningGenerationService ──────────────────────────────────────

func TestMockReasoningService_ReturnsNonEmptyReasoning(t *testing.T) {
	// TC-402-1
	svc := infraAI.NewMockComboReasoningGenerationService()
	candidate := services.ComboCandidate{
		Id: "combo-1",
		Items: []valueobjects.ComboItem{
			{Name: "Linen Blazer"},
			{Name: "Wide-Leg Trousers"},
		},
	}
	r, err := svc.GenerateReasoning(candidate, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Text == "" {
		t.Error("expected non-empty reasoning text")
	}
}

func TestMockReasoningService_ReasoningContainsItemName(t *testing.T) {
	// TC-402-3
	svc := infraAI.NewMockComboReasoningGenerationService()
	candidate := services.ComboCandidate{
		Items: []valueobjects.ComboItem{
			{Name: "Linen Blazer"},
			{Name: "Wide-Leg Trousers"},
		},
	}
	r, _ := svc.GenerateReasoning(candidate, nil)
	if !strings.Contains(r.Text, "Linen Blazer") {
		t.Errorf("expected reasoning to mention 'Linen Blazer', got: %q", r.Text)
	}
}

func TestMockReasoningService_WithPreferences_ReasoningReferencesOccasion(t *testing.T) {
	// TC-402-4
	svc := infraAI.NewMockComboReasoningGenerationService()
	candidate := services.ComboCandidate{
		Items: []valueobjects.ComboItem{
			{Name: "Silk Top"},
			{Name: "Linen Skirt"},
		},
	}
	prefs := &valueobjects.StylePreferences{
		Occasions: []valueobjects.Occasion{valueobjects.OccasionBeach},
	}
	r, _ := svc.GenerateReasoning(candidate, prefs)
	if !strings.Contains(r.Text, "beach") {
		t.Errorf("expected reasoning to reference occasion 'beach', got: %q", r.Text)
	}
}

func TestMockReasoningService_EmptyItems_DoesNotError(t *testing.T) {
	svc := infraAI.NewMockComboReasoningGenerationService()
	candidate := services.ComboCandidate{Id: "combo-1", Items: nil}
	r, err := svc.GenerateReasoning(candidate, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Text == "" {
		t.Error("expected fallback reasoning text for empty items")
	}
}

// ── MockPreferenceInterpretationService ──────────────────────────────────────

func TestMockInterpretationService_EmptyPrefs_ReturnsGenericSummary(t *testing.T) {
	// TC-302-5 / TC-302-2
	svc := infraAI.NewMockPreferenceInterpretationService()
	summary, err := svc.Interpret(valueobjects.StylePreferences{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.Text == "" {
		t.Error("expected non-empty summary for empty preferences")
	}
}

func TestMockInterpretationService_WithOccasion_SummaryNonEmpty(t *testing.T) {
	// TC-302-1
	svc := infraAI.NewMockPreferenceInterpretationService()
	prefs := valueobjects.StylePreferences{
		Occasions: []valueobjects.Occasion{valueobjects.OccasionCasual},
	}
	summary, err := svc.Interpret(prefs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.Text == "" {
		t.Error("expected non-empty summary")
	}
}

func TestMockInterpretationService_FreeTextOnly_SummaryContainsFreeText(t *testing.T) {
	// TC-302-3
	svc := infraAI.NewMockPreferenceInterpretationService()
	prefs := valueobjects.StylePreferences{FreeText: "something light for a summer trip"}
	summary, _ := svc.Interpret(prefs)
	if !strings.Contains(summary.Text, "summer trip") {
		t.Errorf("expected summary to reference free text, got: %q", summary.Text)
	}
}

func TestMockInterpretationService_EchosPreferences(t *testing.T) {
	svc := infraAI.NewMockPreferenceInterpretationService()
	prefs := valueobjects.StylePreferences{
		Occasions: []valueobjects.Occasion{valueobjects.OccasionCasual},
		Styles:    []valueobjects.StyleDirection{valueobjects.StyleMinimalist},
	}
	summary, _ := svc.Interpret(prefs)
	if len(summary.Preferences.Occasions) != 1 {
		t.Error("expected preferences to be echoed back in summary")
	}
}
