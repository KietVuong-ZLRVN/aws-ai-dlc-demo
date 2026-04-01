package valueobject

import (
	"testing"

	"pgregory.net/rapid"
)

// --- PBT 4.1.6: valid construction ---

func TestPagination_ValidConstruction(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		offset := rapid.IntRange(0, 10000).Draw(t, "offset")
		limit := rapid.IntRange(1, 500).Draw(t, "limit")

		p, err := NewPagination(offset, limit)

		if err != nil {
			t.Fatalf("expected no error for offset=%d limit=%d, got: %v", offset, limit, err)
		}
		if p.Offset != offset {
			t.Fatalf("Offset: got %d, want %d", p.Offset, offset)
		}
		if p.Limit != limit {
			t.Fatalf("Limit: got %d, want %d", p.Limit, limit)
		}
	})
}

// --- PBT 4.1.7: rejects negative offset ---

func TestPagination_RejectsNegativeOffset(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		offset := rapid.IntRange(-10000, -1).Draw(t, "negOffset")
		limit := rapid.IntRange(1, 100).Draw(t, "limit")

		_, err := NewPagination(offset, limit)

		if err == nil {
			t.Fatalf("expected error for offset=%d, got nil", offset)
		}
	})
}

// --- PBT 4.1.8: limit=0 falls back to default (20), negative limit returns error ---
// Note: Unit 1 Pagination treats limit=0 as "use default (20)", not an error.

func TestPagination_ZeroLimitDefaultsTo20(t *testing.T) {
	p, err := NewPagination(0, 0)
	if err != nil {
		t.Fatalf("expected no error for limit=0 (default), got: %v", err)
	}
	if p.Limit != defaultLimit {
		t.Fatalf("expected default limit %d, got %d", defaultLimit, p.Limit)
	}
}

func TestPagination_RejectsNegativeLimit(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		offset := rapid.IntRange(0, 100).Draw(t, "offset")
		limit := rapid.IntRange(-10000, -1).Draw(t, "negLimit")

		_, err := NewPagination(offset, limit)

		if err == nil {
			t.Fatalf("expected error for limit=%d, got nil", limit)
		}
	})
}
