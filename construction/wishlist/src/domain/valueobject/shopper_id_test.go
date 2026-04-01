package valueobject

import (
	"testing"

	"pgregory.net/rapid"
)

func TestShopperId_ValidConstruction(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		raw := rapid.StringMatching(`[A-Za-z0-9_-]{1,20}`).Draw(t, "raw")
		sid, err := NewShopperId(raw)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if sid.String() != raw {
			t.Fatalf("String(): got %q, want %q", sid.String(), raw)
		}
	})
}

func TestShopperId_RejectsEmpty(t *testing.T) {
	_, err := NewShopperId("")
	if err == nil {
		t.Fatal("expected error for empty ShopperId, got nil")
	}
}
