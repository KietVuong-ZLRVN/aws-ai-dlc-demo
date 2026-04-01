package valueobject

import (
	"strconv"
	"testing"

	"pgregory.net/rapid"
)

// priceStr formats a float as a plain decimal string without scientific notation,
// safe to use in "min-max" price range strings.
func priceStr(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

// --- PBT 4.1.1: valid construction ---

func TestPriceRange_ValidConstruction(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		min := rapid.Float64Range(0, 1000).Draw(t, "min")
		max := rapid.Float64Range(min, 2000).Draw(t, "max")

		pr, err := NewPriceRange(priceStr(min) + "-" + priceStr(max))

		if err != nil {
			t.Fatalf("expected no error for valid range min=%v max=%v, got: %v", min, max, err)
		}
		if pr.Min != min {
			t.Fatalf("Min: got %v, want %v", pr.Min, min)
		}
		if pr.Max != max {
			t.Fatalf("Max: got %v, want %v", pr.Max, max)
		}
	})
}

// --- PBT 4.1.2: rejects min > max ---

func TestPriceRange_RejectsInvalidRange(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate min strictly greater than max by ensuring min > max.
		base := rapid.Float64Range(1, 1000).Draw(t, "base")
		extra := rapid.Float64Range(0.01, 500).Draw(t, "extra")
		min := base + extra
		max := base

		_, err := NewPriceRange(priceStr(min) + "-" + priceStr(max))

		if err == nil {
			t.Fatalf("expected error for min=%v > max=%v, got nil", min, max)
		}
	})
}

// --- PBT 4.1.3: rejects negative min ---

func TestPriceRange_RejectsNegativeMin(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		min := rapid.Float64Range(-1000, -0.001).Draw(t, "negMin")
		max := rapid.Float64Range(0, 1000).Draw(t, "max")

		_, err := NewPriceRange(priceStr(min) + "-" + priceStr(max))

		if err == nil {
			t.Fatalf("expected error for negative min=%v, got nil", min)
		}
	})
}

// --- Example 4.1.4: rejects malformed strings ---

func TestPriceRange_RejectsMalformed(t *testing.T) {
	cases := []string{
		"",
		"100",
		"abc-def",
		"100-",
		"-200",
		"abc",
		"100-200-300",
	}
	for _, c := range cases {
		t.Run(c, func(t *testing.T) {
			_, err := NewPriceRange(c)
			if err == nil {
				t.Fatalf("expected error for %q, got nil", c)
			}
		})
	}
}

// --- PBT 4.1.5: round-trip ---

func TestPriceRange_RoundTrip(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		min := rapid.Float64Range(0, 500).Draw(t, "min")
		max := rapid.Float64Range(min, 1000).Draw(t, "max")

		s := priceStr(min) + "-" + priceStr(max)
		pr, err := NewPriceRange(s)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if pr.Min != min || pr.Max != max {
			t.Fatalf("round-trip failed: input (%v,%v) got (%v,%v)", min, max, pr.Min, pr.Max)
		}
	})
}
