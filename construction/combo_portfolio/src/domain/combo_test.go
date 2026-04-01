package domain_test

import (
	"strings"
	"testing"

	"github.com/KietVuong-ZLRVN/aws-ai-dlc-demo/combo-portfolio/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

// ---------------------------------------------------------------------------
// Generators
// ---------------------------------------------------------------------------

func genValidName(t *rapid.T) string {
	return rapid.StringOfN(rapid.Rune(), 1, 100, -1).Draw(t, "name")
}

func genShopperID(t *rapid.T) domain.ShopperId {
	s := rapid.StringOfN(rapid.Rune(), 1, 40, -1).Draw(t, "shopperID")
	return domain.ShopperId(s)
}

func genUniqueItems(t *rapid.T, n int) []domain.ComboItem {
	items := make([]domain.ComboItem, n)
	for i := 0; i < n; i++ {
		items[i] = domain.ComboItem{
			ConfigSku: "cfg",
			SimpleSku: strings.Repeat("x", i+1),
			Name:      "product",
			ImageUrl:  "https://example.com/img.jpg",
			Price:     10.0,
		}
	}
	return items
}

func genValidItemCount(t *rapid.T) int {
	return rapid.IntRange(2, 10).Draw(t, "itemCount")
}

func newValidCombo(t interface {
	Helper()
	Fatal(...interface{})
	Error(...interface{})
}, shopperID domain.ShopperId, n int) *domain.Combo {
	name, err := domain.NewComboName("My Combo")
	if err != nil {
		t.Fatal(err)
	}
	items := make([]domain.ComboItem, n)
	for i := 0; i < n; i++ {
		items[i] = domain.ComboItem{
			ConfigSku: "cfg",
			SimpleSku: strings.Repeat("x", i+1),
			Name:      "product",
			ImageUrl:  "https://example.com/img.jpg",
			Price:     10.0,
		}
	}
	combo, err := domain.NewCombo(domain.NewComboId(), shopperID, name, items, domain.VisibilityPrivate)
	if err != nil {
		t.Fatal(err)
	}
	return combo
}

// ---------------------------------------------------------------------------
// 2a. ComboName value-object
// ---------------------------------------------------------------------------

func TestComboName_ValidLength(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		s := genValidName(t)
		name, err := domain.NewComboName(s)
		if err != nil {
			t.Fatalf("expected valid for len=%d, got %v", len(s), err)
		}
		if name.String() != s {
			t.Fatalf("String() mismatch: got %q, want %q", name.String(), s)
		}
	})
}

func TestComboName_EmptyAlwaysRejected(t *testing.T) {
	_, err := domain.NewComboName("")
	assert.ErrorIs(t, err, domain.ErrInvalidComboName)
}

func TestComboName_OverLimitAlwaysRejected(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		n := rapid.IntRange(101, 500).Draw(t, "len")
		s := strings.Repeat("a", n)
		_, err := domain.NewComboName(s)
		if err == nil {
			t.Fatalf("expected error for len=%d, got nil", n)
		}
	})
}

// ---------------------------------------------------------------------------
// 2b. NewCombo item-count invariant
// ---------------------------------------------------------------------------

func TestNewCombo_TooFewItemsRejected(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		n := rapid.IntRange(0, 1).Draw(t, "count")
		items := genUniqueItems(t, n)
		name, _ := domain.NewComboName("c")
		_, err := domain.NewCombo(domain.NewComboId(), "s1", name, items, domain.VisibilityPrivate)
		if err != domain.ErrInvalidItemCount {
			t.Fatalf("expected ErrInvalidItemCount for %d items, got %v", n, err)
		}
	})
}

func TestNewCombo_TooManyItemsRejected(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		n := rapid.IntRange(11, 50).Draw(t, "count")
		items := genUniqueItems(t, n)
		name, _ := domain.NewComboName("c")
		_, err := domain.NewCombo(domain.NewComboId(), "s1", name, items, domain.VisibilityPrivate)
		if err != domain.ErrInvalidItemCount {
			t.Fatalf("expected ErrInvalidItemCount for %d items, got %v", n, err)
		}
	})
}

func TestNewCombo_ValidItemCountAccepted(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		n := genValidItemCount(t)
		items := genUniqueItems(t, n)
		name, _ := domain.NewComboName("c")
		combo, err := domain.NewCombo(domain.NewComboId(), "s1", name, items, domain.VisibilityPrivate)
		if err != nil {
			t.Fatalf("expected no error for %d items, got %v", n, err)
		}
		if len(combo.Items()) != n {
			t.Fatalf("expected %d items, got %d", n, len(combo.Items()))
		}
	})
}

func TestNewCombo_DuplicateSkuRejected(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		n := rapid.IntRange(2, 10).Draw(t, "count")
		items := genUniqueItems(t, n)
		dupIdx := rapid.IntRange(1, n-1).Draw(t, "dupIdx")
		items[dupIdx].SimpleSku = items[0].SimpleSku
		name, _ := domain.NewComboName("c")
		_, err := domain.NewCombo(domain.NewComboId(), "s1", name, items, domain.VisibilityPrivate)
		if err != domain.ErrDuplicateItem {
			t.Fatalf("expected ErrDuplicateItem, got %v", err)
		}
	})
}

// ---------------------------------------------------------------------------
// 2c. Share / Visibility invariants
// ---------------------------------------------------------------------------

func TestNewCombo_DefaultIsPrivateWithNoToken(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		n := genValidItemCount(t)
		items := genUniqueItems(t, n)
		name, _ := domain.NewComboName("c")
		combo, err := domain.NewCombo(domain.NewComboId(), "s1", name, items, domain.VisibilityPrivate)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if combo.Visibility() != domain.VisibilityPrivate {
			t.Fatalf("expected private, got %v", combo.Visibility())
		}
		if combo.ShareToken() != nil {
			t.Fatalf("expected nil share token, got non-nil")
		}
	})
}

func TestCombo_ShareSetsPublicAndToken(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		combo := newValidCombo(t, "shopper1", 3)
		tokenStr := rapid.StringOfN(rapid.Rune(), 1, 40, -1).Draw(t, "token")
		token := domain.ShareToken(tokenStr)
		combo.Share(token)
		if combo.Visibility() != domain.VisibilityPublic {
			t.Fatalf("expected public, got %v", combo.Visibility())
		}
		if combo.ShareToken() == nil || *combo.ShareToken() != token {
			t.Fatalf("expected token %q, got %v", token, combo.ShareToken())
		}
	})
}

func TestCombo_MakePrivateAtomicallyRevokesToken(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		combo := newValidCombo(t, "shopper1", 3)
		combo.Share("some-token")
		combo.MakePrivate()
		if combo.Visibility() != domain.VisibilityPrivate {
			t.Fatalf("expected private, got %v", combo.Visibility())
		}
		if combo.ShareToken() != nil {
			t.Fatal("expected nil token after MakePrivate")
		}
	})
}

func TestCombo_MakePrivateIsIdempotent(t *testing.T) {
	combo := newValidCombo(t, "shopper1", 3)
	combo.Share("token-abc")
	combo.MakePrivate()
	combo.MakePrivate()
	assert.Equal(t, domain.VisibilityPrivate, combo.Visibility())
	assert.Nil(t, combo.ShareToken())
}

func TestCombo_ShareMakePrivateReshareIsCycleConsistent(t *testing.T) {
	combo := newValidCombo(t, "shopper1", 3)
	combo.Share("token-1")
	combo.MakePrivate()
	combo.Share("token-2")
	assert.Equal(t, domain.VisibilityPublic, combo.Visibility())
	require.NotNil(t, combo.ShareToken())
	assert.Equal(t, domain.ShareToken("token-2"), *combo.ShareToken())
}

// ---------------------------------------------------------------------------
// 2d. Rename invariant
// ---------------------------------------------------------------------------

func TestCombo_RenameUpdatesName(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		combo := newValidCombo(t, "shopper1", 3)
		newNameStr := genValidName(t)
		newName, err := domain.NewComboName(newNameStr)
		if err != nil {
			return // skip invalid generated name
		}
		combo.Rename(newName)
		if combo.Name() != newName {
			t.Fatalf("expected name %q, got %q", newName, combo.Name())
		}
	})
}

func TestCombo_RenamePreservesIdentityAndItems(t *testing.T) {
	combo := newValidCombo(t, "shopper1", 4)
	originalID := combo.ID()
	originalShopperID := combo.ShopperID()
	originalItemCount := len(combo.Items())
	newName, _ := domain.NewComboName("New Name")
	combo.Rename(newName)
	assert.Equal(t, originalID, combo.ID())
	assert.Equal(t, originalShopperID, combo.ShopperID())
	assert.Equal(t, originalItemCount, len(combo.Items()))
}

// ---------------------------------------------------------------------------
// 2e. Domain events
// ---------------------------------------------------------------------------

func TestNewCombo_EmitsExactlyOneComboCreatedEvent(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		n := genValidItemCount(t)
		items := genUniqueItems(t, n)
		name, _ := domain.NewComboName("c")
		combo, err := domain.NewCombo(domain.NewComboId(), "s1", name, items, domain.VisibilityPrivate)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		events := combo.PopEvents()
		if len(events) != 1 {
			t.Fatalf("expected 1 event, got %d", len(events))
		}
		if events[0].EventName() != "ComboCreated" {
			t.Fatalf("expected ComboCreated, got %q", events[0].EventName())
		}
	})
}

func TestCombo_PopEventsIsDestructive(t *testing.T) {
	combo := newValidCombo(t, "shopper1", 3)
	first := combo.PopEvents()
	second := combo.PopEvents()
	assert.Len(t, first, 1, "first PopEvents should have 1 event")
	assert.Empty(t, second, "second PopEvents should be empty")
}

func TestCombo_OwnedByIsCorrect(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		ownerID := genShopperID(t)
		combo := newValidCombo(t, ownerID, 2)
		if !combo.OwnedBy(ownerID) {
			t.Fatal("OwnedBy should be true for the owner")
		}
		otherID := domain.ShopperId(ownerID.String() + "_other")
		if combo.OwnedBy(otherID) {
			t.Fatalf("OwnedBy should be false for different shopperID %q", otherID)
		}
	})
}
