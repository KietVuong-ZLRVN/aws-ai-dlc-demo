package entities_test

import (
	"ai-styling-engine/domain/entities"
	"ai-styling-engine/domain/valueobjects"
	"testing"
)

// ── Combo ────────────────────────────────────────────────────────────────────

func TestNewCombo_FieldsSet(t *testing.T) {
	items := []valueobjects.ComboItem{
		{ConfigSku: "CFG-A", SimpleSku: "SKU-A", Source: valueobjects.ItemSourceWishlist},
		{ConfigSku: "CFG-B", SimpleSku: "SKU-B", Source: valueobjects.ItemSourceCatalog},
	}
	combo := entities.NewCombo("combo-1", items, 1)
	if combo.Id != "combo-1" {
		t.Errorf("expected Id=combo-1, got %q", combo.Id)
	}
	if combo.Rank != 1 {
		t.Errorf("expected Rank=1, got %d", combo.Rank)
	}
	if len(combo.Items) != 2 {
		t.Errorf("expected 2 items, got %d", len(combo.Items))
	}
}

func TestCombo_AttachReasoning(t *testing.T) {
	combo := entities.NewCombo("combo-1", nil, 1)
	r, _ := valueobjects.NewComboReasoning("Great pairing for summer.")
	combo.AttachReasoning(r)
	if combo.Reasoning.Text != "Great pairing for summer." {
		t.Errorf("unexpected reasoning text: %q", combo.Reasoning.Text)
	}
}

func TestCombo_Reject(t *testing.T) {
	combo := entities.NewCombo("combo-1", nil, 1)
	if combo.IsRejected() {
		t.Error("new combo should not be rejected")
	}
	combo.Reject()
	if !combo.IsRejected() {
		t.Error("combo should be rejected after Reject()")
	}
}

func TestCombo_NotRejectedByDefault(t *testing.T) {
	combo := entities.NewCombo("combo-x", nil, 2)
	if combo.IsRejected() {
		t.Error("combo should not be rejected before Reject() is called")
	}
}

// ── FallbackResult ───────────────────────────────────────────────────────────

func TestNewFallbackResult_FieldsSet(t *testing.T) {
	alts := []valueobjects.AlternativeItem{
		{ConfigSku: "CFG-ALT", Name: "Classic White Shirt", Reason: "Matches your style"},
	}
	fb := entities.NewFallbackResult("No combo could be formed.", alts)
	if fb.Message != "No combo could be formed." {
		t.Errorf("unexpected message: %q", fb.Message)
	}
	if len(fb.Alternatives) != 1 {
		t.Errorf("expected 1 alternative, got %d", len(fb.Alternatives))
	}
}

func TestNewFallbackResult_EmptyAlternatives(t *testing.T) {
	// TC-404-4 setup: FallbackPolicy should handle this without panicking.
	fb := entities.NewFallbackResult("Exhausted options.", []valueobjects.AlternativeItem{})
	if len(fb.Alternatives) != 0 {
		t.Errorf("expected 0 alternatives, got %d", len(fb.Alternatives))
	}
}

// ── Phase 3 gap tests ────────────────────────────────────────────────────────

// TC-DOM-11: Rank values are stored correctly across two combos
func TestCombo_RankStoredCorrectly(t *testing.T) {
	c1 := entities.NewCombo("combo-1", nil, 1)
	c2 := entities.NewCombo("combo-2", nil, 2)
	if c1.Rank != 1 {
		t.Errorf("expected Rank=1, got %d", c1.Rank)
	}
	if c2.Rank != 2 {
		t.Errorf("expected Rank=2, got %d", c2.Rank)
	}
}

// TC-DOM-12: FallbackResult stores alternative with empty Reason as-is (enforcement is at policy level)
func TestFallbackResult_AlternativeWithEmptyReason_StoredAsIs(t *testing.T) {
	alts := []valueobjects.AlternativeItem{
		{ConfigSku: "CFG-X", Reason: ""},
	}
	fb := entities.NewFallbackResult("No combo.", alts)
	if len(fb.Alternatives) != 1 {
		t.Fatalf("expected 1 alternative, got %d", len(fb.Alternatives))
	}
	if fb.Alternatives[0].Reason != "" {
		t.Errorf("expected empty Reason to be stored as-is, got %q", fb.Alternatives[0].Reason)
	}
}
