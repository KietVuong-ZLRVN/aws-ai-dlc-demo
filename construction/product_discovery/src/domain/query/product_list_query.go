package query

import "product_discovery/domain/valueobject"

// ProductListQuery carries the parameters for a product list / search request.
type ProductListQuery struct {
	Query      string
	CategoryID string
	Colors     []string
	PriceRange *valueobject.PriceRange
	Pagination valueobject.Pagination
}
