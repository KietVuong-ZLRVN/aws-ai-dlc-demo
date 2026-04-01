package readmodel

// ProductSummaryReadModel is a single product card in a list response.
type ProductSummaryReadModel struct {
	ConfigSku string   `json:"configSku"`
	Name      string   `json:"name"`
	Brand     string   `json:"brand"`
	Price     Money    `json:"price"`
	ImageUrl  string   `json:"imageUrl"`
	Slug      string   `json:"slug"`
	InStock   bool     `json:"inStock"`
	Colors    []string `json:"colors"`
}

// ProductListReadModel is the full list / search response.
type ProductListReadModel struct {
	Products []ProductSummaryReadModel `json:"products"`
	Total    int                       `json:"total"`
	Filters  FilterFacetsReadModel     `json:"filters"`
}
