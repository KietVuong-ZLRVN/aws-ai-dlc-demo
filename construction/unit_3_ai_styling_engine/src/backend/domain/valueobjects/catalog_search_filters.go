package valueobjects

// CatalogSearchFilters is constructed from StylePreferences and passed to
// ProductCatalogRepository when supplementary catalog items are needed.
type CatalogSearchFilters struct {
	Colors     []Color
	Occasion   *Occasion
	PriceRange *BudgetRange
	CategoryId *int
	Limit      int
}

// FromPreferences derives catalog search filters from the shopper's style preferences.
func CatalogSearchFiltersFromPreferences(prefs *StylePreferences) CatalogSearchFilters {
	f := CatalogSearchFilters{Limit: 20}
	if prefs == nil {
		return f
	}
	f.Colors = prefs.Colors.Preferred
	if len(prefs.Occasions) > 0 {
		occ := prefs.Occasions[0]
		f.Occasion = &occ
	}
	f.PriceRange = prefs.Budget
	return f
}
