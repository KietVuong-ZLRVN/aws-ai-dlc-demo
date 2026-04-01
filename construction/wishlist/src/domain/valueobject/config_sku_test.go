package valueobject

import (
	"testing"

	"pgregory.net/rapid"
)

// 4.5.3: valid construction for any non-empty string
func TestConfigSku_ValidConstruction(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		raw := rapid.StringMatching(`[A-Za-z0-9_-]{1,20}`).Draw(t, "raw")

		sku, err := NewConfigSku(raw)

		if err != nil {
			t.Fatalf("expected no error for %q, got: %v", raw, err)
		}
		if sku.String() != raw {
			t.Fatalf("String(): got %q, want %q", sku.String(), raw)
		}
	})
}

// 4.5.4: empty string is rejected
func TestConfigSku_RejectsEmpty(t *testing.T) {
	_, err := NewConfigSku("")
	if err == nil {
		t.Fatal("expected error for empty ConfigSku, got nil")
	}
}
