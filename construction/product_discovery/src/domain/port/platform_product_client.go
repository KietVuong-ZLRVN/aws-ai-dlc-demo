package port

import "context"

// PlatformProduct is the raw product struct exchanged between the platform client
// and the domain assemblers.
type PlatformProduct struct {
	ConfigSku string
	Name      string
	Brand     string
	Category  string
	Price     float64
	Currency  string
	Colors    []string
	Simples   []PlatformSimple
	ImageUrl  string
	UrlKey    string
	Occasions []string
}

// PlatformSimple is the raw variant (simple product) struct.
type PlatformSimple struct {
	SimpleSku string
	Size      string
	Color     string
	Quantity  int
}

// ProductListParams carries filter and pagination parameters for a product list request.
type ProductListParams struct {
	Query      string
	CategoryID string
	Colors     []string
	MinPrice   float64
	MaxPrice   float64
	Offset     int
	Limit      int
}

// RawProductListPayload wraps a page of products together with the total count.
type RawProductListPayload struct {
	Products   []PlatformProduct
	TotalCount int
}

// RawFilterPayload wraps the available filter facets for the full catalogue.
type RawFilterPayload struct {
	Colors     []string
	Categories []string
	MinPrice   float64
	MaxPrice   float64
}

// RawProductDetailPayload wraps the raw detail for a single configurable product.
type RawProductDetailPayload struct {
	Product PlatformProduct
}

// PlatformProductClient is the port through which the application layer
// communicates with the upstream product platform.
type PlatformProductClient interface {
	// FetchProductList returns a paginated, filtered product list.
	FetchProductList(ctx context.Context, params ProductListParams) (*RawProductListPayload, error)

	// FetchProductFilters returns the available filter facets for the full catalogue.
	FetchProductFilters(ctx context.Context, params ProductListParams) (*RawFilterPayload, error)

	// FetchProductDetail returns the detail for a single configurable product.
	// Returns nil, nil when the product is not found.
	FetchProductDetail(ctx context.Context, configSku string) (*RawProductDetailPayload, error)
}
