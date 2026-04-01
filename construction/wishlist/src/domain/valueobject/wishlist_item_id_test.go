package valueobject

import (
	"testing"

	"pgregory.net/rapid"
)

func TestWishlistItemId_ValidConstruction(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		raw := rapid.StringMatching(`[A-Za-z0-9_-]{1,20}`).Draw(t, "raw")
		iid, err := NewWishlistItemId(raw)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if iid.String() != raw {
			t.Fatalf("String(): got %q, want %q", iid.String(), raw)
		}
	})
}

func TestWishlistItemId_RejectsEmpty(t *testing.T) {
	_, err := NewWishlistItemId("")
	if err == nil {
		t.Fatal("expected error for empty WishlistItemId, got nil")
	}
}
