package valueobjects

// StyleOptionsCatalogue holds the static predefined options for the preference input form.
// This is a configuration value object — not fetched from a database.
type StyleOptionsCatalogue struct {
	Occasions []Occasion
	Styles    []StyleDirection
	Colors    []Color
}

// DefaultStyleOptionsCatalogue returns the canonical set of style options.
func DefaultStyleOptionsCatalogue() StyleOptionsCatalogue {
	return StyleOptionsCatalogue{
		Occasions: ValidOccasions,
		Styles:    ValidStyleDirections,
		Colors:    ValidColors,
	}
}
