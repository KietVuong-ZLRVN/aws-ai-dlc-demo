package valueobject

import (
	"fmt"
	"strconv"
	"strings"
)

// PriceRange represents a min/max price filter.
type PriceRange struct {
	Min float64
	Max float64
}

// NewPriceRange parses a "min-max" string into a PriceRange value object.
// Returns an error if the format is wrong or if min > max.
func NewPriceRange(s string) (PriceRange, error) {
	parts := strings.SplitN(s, "-", 2)
	if len(parts) != 2 {
		return PriceRange{}, fmt.Errorf("invalid price range format %q: expected \"min-max\"", s)
	}
	min, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	if err != nil {
		return PriceRange{}, fmt.Errorf("invalid price range min %q: %w", parts[0], err)
	}
	max, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err != nil {
		return PriceRange{}, fmt.Errorf("invalid price range max %q: %w", parts[1], err)
	}
	if min > max {
		return PriceRange{}, fmt.Errorf("price range min (%f) must not be greater than max (%f)", min, max)
	}
	return PriceRange{Min: min, Max: max}, nil
}
