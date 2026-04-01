package valueobject

import (
	"testing"

	"pgregory.net/rapid"
)

// 4.5.7: valid construction
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

// 4.5.8: rejects negative offset
func TestPagination_RejectsNegativeOffset(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		offset := rapid.IntRange(-10000, -1).Draw(t, "negOffset")

		_, err := NewPagination(offset, 10)

		if err == nil {
			t.Fatalf("expected error for offset=%d, got nil", offset)
		}
	})
}

// 4.5.9: rejects zero and negative limit
func TestPagination_RejectsZeroOrNegativeLimit(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		limit := rapid.IntRange(-10000, 0).Draw(t, "badLimit")

		_, err := NewPagination(0, limit)

		if err == nil {
			t.Fatalf("expected error for limit=%d, got nil", limit)
		}
	})
}
