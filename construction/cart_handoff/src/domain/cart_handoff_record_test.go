package domain_test

import (
	"testing"
	"time"

	"github.com/KietVuong-ZLRVN/aws-ai-dlc-demo/cart-handoff/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

// ---------------------------------------------------------------------------
// Generators
// ---------------------------------------------------------------------------

func genNonEmptyString(t *rapid.T, label string) string {
	return rapid.StringOfN(rapid.Rune(), 1, 40, -1).Draw(t, label)
}

func genCartItem(t *rapid.T, i int) domain.CartItem {
	return domain.CartItem{
		SimpleSku: "sku-" + string(rune('a'+i)),
		Quantity:  1,
		Size:      "M",
	}
}

func genCartItems(t *rapid.T, n int) []domain.CartItem {
	items := make([]domain.CartItem, n)
	for i := 0; i < n; i++ {
		items[i] = genCartItem(t, i)
	}
	return items
}

func genSkippedItems(t *rapid.T, n int) []domain.SkippedItem {
	items := make([]domain.SkippedItem, n)
	for i := 0; i < n; i++ {
		items[i] = domain.SkippedItem{SimpleSku: "sku-skipped-" + string(rune('a'+i)), Reason: "out_of_stock"}
	}
	return items
}

// ---------------------------------------------------------------------------
// 3a. HandoffSource validation
// ---------------------------------------------------------------------------

// TestHandoffSource_TypeDiscriminantDeterminesValidation verifies that the domain uses Type
// as the discriminant. Setting extra fields beyond what the Type requires does NOT cause
// a rejection at the domain level (mutual-exclusion enforcement is in the application layer).
func TestHandoffSource_TypeDiscriminantDeterminesValidation(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		n := rapid.IntRange(1, 5).Draw(t, "n")
		// Type=saved_combo with non-empty ComboId must succeed even when Items is also set.
		source := domain.HandoffSource{
			Type:    domain.HandoffSourceTypeSavedCombo,
			ComboId: genNonEmptyString(t, "comboId"),
			Items:   genCartItems(t, n), // ignored by domain when Type=saved_combo
		}
		added := genCartItems(t, 1)
		_, err := domain.NewCartHandoffRecord("shopper1", source, added, nil)
		if err != nil {
			t.Fatalf("domain should accept valid saved_combo source regardless of Items field: %v", err)
		}
	})
}

func TestHandoffSource_NeitherComboIdNorItemsRejected(t *testing.T) {
	source := domain.HandoffSource{Type: domain.HandoffSourceTypeSavedCombo, ComboId: ""}
	_, err := domain.NewCartHandoffRecord("shopper1", source, nil, nil)
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidHandoffSource)

	source2 := domain.HandoffSource{Type: domain.HandoffSourceTypeInlineItems, Items: nil}
	_, err2 := domain.NewCartHandoffRecord("shopper1", source2, nil, nil)
	assert.Error(t, err2)
	assert.ErrorIs(t, err2, domain.ErrInvalidHandoffSource)
}

func TestHandoffSource_SavedComboOnlyAccepted(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		comboId := genNonEmptyString(t, "comboId")
		source := domain.HandoffSource{
			Type:    domain.HandoffSourceTypeSavedCombo,
			ComboId: comboId,
		}
		added := genCartItems(t, rapid.IntRange(1, 5).Draw(t, "n"))
		_, err := domain.NewCartHandoffRecord("shopper1", source, added, nil)
		if err != nil {
			t.Fatalf("expected no error for saved_combo source, got %v", err)
		}
	})
}

func TestHandoffSource_InlineItemsOnlyAccepted(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		n := rapid.IntRange(1, 5).Draw(t, "n")
		source := domain.HandoffSource{
			Type:  domain.HandoffSourceTypeInlineItems,
			Items: genCartItems(t, n),
		}
		added := genCartItems(t, n)
		_, err := domain.NewCartHandoffRecord("shopper1", source, added, nil)
		if err != nil {
			t.Fatalf("expected no error for inline_items source, got %v", err)
		}
	})
}

// ---------------------------------------------------------------------------
// 3b. HandoffStatus derivation consistency
// ---------------------------------------------------------------------------

func TestHandoffStatus_OkWhenAllAdded(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		n := rapid.IntRange(1, 10).Draw(t, "n")
		source := domain.HandoffSource{Type: domain.HandoffSourceTypeSavedCombo, ComboId: "combo-1"}
		added := genCartItems(t, n)
		rec, err := domain.NewCartHandoffRecord("shopper1", source, added, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rec.Status() != domain.HandoffStatusOk {
			t.Fatalf("expected ok, got %v", rec.Status())
		}
	})
}

func TestHandoffStatus_PartialWhenSomeSkipped(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		nAdded := rapid.IntRange(1, 5).Draw(t, "nAdded")
		nSkipped := rapid.IntRange(1, 5).Draw(t, "nSkipped")
		source := domain.HandoffSource{Type: domain.HandoffSourceTypeSavedCombo, ComboId: "combo-1"}
		added := genCartItems(t, nAdded)
		skipped := genSkippedItems(t, nSkipped)
		rec, err := domain.NewCartHandoffRecord("shopper1", source, added, skipped)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rec.Status() != domain.HandoffStatusPartial {
			t.Fatalf("expected partial, got %v", rec.Status())
		}
	})
}

func TestHandoffStatus_FailedWhenNoItemsAdded(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		source := domain.HandoffSource{Type: domain.HandoffSourceTypeSavedCombo, ComboId: "combo-1"}
		nSkipped := rapid.IntRange(0, 5).Draw(t, "nSkipped")
		skipped := genSkippedItems(t, nSkipped)
		rec, err := domain.NewCartHandoffRecord("shopper1", source, nil, skipped)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rec.Status() != domain.HandoffStatusFailed {
			t.Fatalf("expected failed, got %v (added=0, skipped=%d)", rec.Status(), nSkipped)
		}
	})
}

func TestHandoffStatus_AlwaysOneOfThreeValues(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		source := domain.HandoffSource{Type: domain.HandoffSourceTypeSavedCombo, ComboId: "combo-1"}
		nAdded := rapid.IntRange(0, 5).Draw(t, "nAdded")
		nSkipped := rapid.IntRange(0, 5).Draw(t, "nSkipped")
		added := genCartItems(t, nAdded)
		skipped := genSkippedItems(t, nSkipped)
		rec, err := domain.NewCartHandoffRecord("shopper1", source, added, skipped)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		valid := rec.Status() == domain.HandoffStatusOk ||
			rec.Status() == domain.HandoffStatusPartial ||
			rec.Status() == domain.HandoffStatusFailed
		if !valid {
			t.Fatalf("unexpected status value: %q", rec.Status())
		}
	})
}

// ---------------------------------------------------------------------------
// 3c. Record immutability
// ---------------------------------------------------------------------------

func TestReconstitueCartHandoffRecord_PreservesAllFields(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	id := domain.CartHandoffRecordId("test-id-001")
	shopperID := domain.ShopperId("shopper-x")
	source := domain.HandoffSource{Type: domain.HandoffSourceTypeSavedCombo, ComboId: "combo-abc"}
	added := []domain.CartItem{{SimpleSku: "sku-1", Quantity: 1, Size: "M"}}
	skipped := []domain.SkippedItem{{SimpleSku: "sku-2", Reason: "out_of_stock"}}

	rec := domain.ReconstitueCartHandoffRecord(id, shopperID, source, domain.HandoffStatusPartial, added, skipped, now)

	require.NotNil(t, rec)
	assert.Equal(t, id, rec.ID())
	assert.Equal(t, shopperID, rec.ShopperID())
	assert.Equal(t, domain.HandoffStatusPartial, rec.Status())
	assert.Equal(t, added, rec.AddedItems())
	assert.Equal(t, skipped, rec.SkippedItems())
	assert.Equal(t, now, rec.RecordedAt())
}

func TestCartHandoffRecord_PopEventsIsDestructive(t *testing.T) {
	source := domain.HandoffSource{Type: domain.HandoffSourceTypeSavedCombo, ComboId: "c1"}
	rec, err := domain.NewCartHandoffRecord("s1", source, []domain.CartItem{{SimpleSku: "sku-1", Quantity: 1, Size: "M"}}, nil)
	require.NoError(t, err)

	first := rec.PopEvents()
	second := rec.PopEvents()
	assert.Len(t, first, 1, "first PopEvents should have 1 event")
	assert.Empty(t, second, "second PopEvents should be empty")
}
