package valueobjects

type StylePreferences struct {
	Occasions []Occasion
	Styles    []StyleDirection
	Budget    *BudgetRange
	Colors    ColorPalette
	FreeText  string
}

// IsEmpty returns true when no preference constraints have been set.
// Used by PreferenceDefaultPolicy to detect quick-generate mode.
func (sp StylePreferences) IsEmpty() bool {
	return len(sp.Occasions) == 0 &&
		len(sp.Styles) == 0 &&
		sp.Budget == nil &&
		len(sp.Colors.Preferred) == 0 &&
		len(sp.Colors.Excluded) == 0 &&
		sp.FreeText == ""
}
