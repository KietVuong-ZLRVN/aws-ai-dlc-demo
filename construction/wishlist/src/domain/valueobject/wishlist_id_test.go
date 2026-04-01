package valueobject

import (
	"testing"

	"pgregory.net/rapid"
)

func TestWishlistId_ValidConstruction(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		raw := rapid.StringMatching(`[A-Za-z0-9_-]{1,20}`).Draw(t, "raw")
		wid, err := NewWishlistId(raw)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if wid.String() != raw {
			t.Fatalf("String(): got %q, want %q", wid.String(), raw)
		}
	})
}

func TestWishlistId_RejectsEmpty(t *testing.T) {
	_, err := NewWishlistId("")
	if err == nil {
		t.Fatal("expected error for empty WishlistId, got nil")
	}
}
