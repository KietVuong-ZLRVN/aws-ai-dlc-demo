package valueobject

import (
	"testing"

	"pgregory.net/rapid"
)

// 4.5.1: valid construction for any non-empty string
func TestSimpleSku_ValidConstruction(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		raw := rapid.StringMatching(`[A-Za-z0-9_-]{1,30}`).Draw(t, "raw")

		sku, err := NewSimpleSku(raw)

		if err != nil {
			t.Fatalf("expected no error for %q, got: %v", raw, err)
		}
		if sku.String() != raw {
			t.Fatalf("String(): got %q, want %q", sku.String(), raw)
		}
	})
}

// 4.5.2: empty string is rejected
func TestSimpleSku_RejectsEmpty(t *testing.T) {
	_, err := NewSimpleSku("")
	if err == nil {
		t.Fatal("expected error for empty SimpleSku, got nil")
	}
}
