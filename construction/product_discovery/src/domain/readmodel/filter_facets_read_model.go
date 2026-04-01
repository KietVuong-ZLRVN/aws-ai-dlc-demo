package readmodel

// Money is a monetary amount embedded in read models.
type Money struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

// PriceRangeFacet holds the min and max price across the catalogue.
type PriceRangeFacet struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

// FilterFacetsReadModel holds the available filter options for a product list.
type FilterFacetsReadModel struct {
	Colors     []string        `json:"colors"`
	Categories []string        `json:"categories"`
	PriceRange PriceRangeFacet `json:"priceRange"`
}
