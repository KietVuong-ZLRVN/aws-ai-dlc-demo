package platform

import (
	"context"
	"math"
	"strings"

	"product_discovery/domain/port"
)

// InMemoryProductClient is a fully in-memory implementation of port.PlatformProductClient.
// It seeds a fixed catalogue of 8 products and never makes any external calls.
type InMemoryProductClient struct {
	catalogue []platformProduct
}

// NewInMemoryProductClient returns an InMemoryProductClient pre-seeded with the
// product catalogue.
func NewInMemoryProductClient() *InMemoryProductClient {
	return &InMemoryProductClient{
		catalogue: seedCatalogue(),
	}
}

// seedCatalogue returns the fixed set of 8 products.
func seedCatalogue() []platformProduct {
	return []platformProduct{
		{
			ConfigSku: "PD-001",
			Name:      "Classic White Tee",
			Brand:     "Uniqlo",
			Category:  "tops",
			Price:     29.90,
			Currency:  "SGD",
			Colors:    []string{"white"},
			Occasions: []string{"casual"},
			ImageUrl:  "/images/pd-001.jpg",
			UrlKey:    "classic-white-tee",
			Simples: []platformSimple{
				{SimpleSku: "PD-001-S", Size: "S", Color: "white", Quantity: 10},
				{SimpleSku: "PD-001-M", Size: "M", Color: "white", Quantity: 5},
				{SimpleSku: "PD-001-L", Size: "L", Color: "white", Quantity: 0},
			},
		},
		{
			ConfigSku: "PD-002",
			Name:      "Slim Fit Jeans",
			Brand:     "Levi's",
			Category:  "bottoms",
			Price:     89.90,
			Currency:  "SGD",
			Colors:    []string{"blue"},
			Occasions: []string{"casual", "smart casual"},
			ImageUrl:  "/images/pd-002.jpg",
			UrlKey:    "slim-fit-jeans",
			Simples: []platformSimple{
				{SimpleSku: "PD-002-28", Size: "28", Color: "blue", Quantity: 3},
				{SimpleSku: "PD-002-30", Size: "30", Color: "blue", Quantity: 8},
				{SimpleSku: "PD-002-32", Size: "32", Color: "blue", Quantity: 0},
			},
		},
		{
			ConfigSku: "PD-003",
			Name:      "Floral Summer Dress",
			Brand:     "Zalora",
			Category:  "dresses",
			Price:     59.90,
			Currency:  "SGD",
			Colors:    []string{"pink"},
			Occasions: []string{"casual", "beach"},
			ImageUrl:  "/images/pd-003.jpg",
			UrlKey:    "floral-summer-dress",
			Simples: []platformSimple{
				{SimpleSku: "PD-003-XS", Size: "XS", Color: "pink", Quantity: 2},
				{SimpleSku: "PD-003-S", Size: "S", Color: "pink", Quantity: 6},
				{SimpleSku: "PD-003-M", Size: "M", Color: "pink", Quantity: 0},
			},
		},
		{
			ConfigSku: "PD-004",
			Name:      "Black Blazer",
			Brand:     "Mango",
			Category:  "outerwear",
			Price:     129.90,
			Currency:  "SGD",
			Colors:    []string{"black"},
			Occasions: []string{"work", "smart casual"},
			ImageUrl:  "/images/pd-004.jpg",
			UrlKey:    "black-blazer",
			Simples: []platformSimple{
				{SimpleSku: "PD-004-S", Size: "S", Color: "black", Quantity: 4},
				{SimpleSku: "PD-004-M", Size: "M", Color: "black", Quantity: 7},
				{SimpleSku: "PD-004-L", Size: "L", Color: "black", Quantity: 1},
			},
		},
		{
			ConfigSku: "PD-005",
			Name:      "Casual Polo Shirt",
			Brand:     "Ralph Lauren",
			Category:  "tops",
			Price:     79.90,
			Currency:  "SGD",
			Colors:    []string{"navy", "white"},
			Occasions: []string{"casual", "smart casual"},
			ImageUrl:  "/images/pd-005.jpg",
			UrlKey:    "casual-polo-shirt",
			Simples: []platformSimple{
				{SimpleSku: "PD-005-S-NVY", Size: "S", Color: "navy", Quantity: 5},
				{SimpleSku: "PD-005-M-NVY", Size: "M", Color: "navy", Quantity: 3},
				{SimpleSku: "PD-005-S-WHT", Size: "S", Color: "white", Quantity: 2},
				{SimpleSku: "PD-005-M-WHT", Size: "M", Color: "white", Quantity: 4},
			},
		},
		{
			ConfigSku: "PD-006",
			Name:      "High-Waist Skirt",
			Brand:     "H&M",
			Category:  "bottoms",
			Price:     39.90,
			Currency:  "SGD",
			Colors:    []string{"black"},
			Occasions: []string{"casual", "work"},
			ImageUrl:  "/images/pd-006.jpg",
			UrlKey:    "high-waist-skirt",
			Simples: []platformSimple{
				{SimpleSku: "PD-006-XS", Size: "XS", Color: "black", Quantity: 0},
				{SimpleSku: "PD-006-S", Size: "S", Color: "black", Quantity: 9},
				{SimpleSku: "PD-006-M", Size: "M", Color: "black", Quantity: 6},
			},
		},
		{
			ConfigSku: "PD-007",
			Name:      "Striped Linen Shirt",
			Brand:     "Topshop",
			Category:  "tops",
			Price:     49.90,
			Currency:  "SGD",
			Colors:    []string{"blue", "white"},
			Occasions: []string{"casual", "beach"},
			ImageUrl:  "/images/pd-007.jpg",
			UrlKey:    "striped-linen-shirt",
			Simples: []platformSimple{
				{SimpleSku: "PD-007-S", Size: "S", Color: "blue", Quantity: 7},
				{SimpleSku: "PD-007-M", Size: "M", Color: "blue", Quantity: 3},
				{SimpleSku: "PD-007-L", Size: "L", Color: "blue", Quantity: 0},
			},
		},
		{
			ConfigSku: "PD-008",
			Name:      "Leopard Print Blouse",
			Brand:     "Zara",
			Category:  "tops",
			Price:     55.90,
			Currency:  "SGD",
			Colors:    []string{"beige"},
			Occasions: []string{"party", "casual"},
			ImageUrl:  "/images/pd-008.jpg",
			UrlKey:    "leopard-print-blouse",
			Simples: []platformSimple{
				{SimpleSku: "PD-008-XS", Size: "XS", Color: "beige", Quantity: 1},
				{SimpleSku: "PD-008-S", Size: "S", Color: "beige", Quantity: 4},
				{SimpleSku: "PD-008-M", Size: "M", Color: "beige", Quantity: 2},
			},
		},
	}
}

// FetchProductList applies all optional filters (AND logic) and returns a paginated result.
func (c *InMemoryProductClient) FetchProductList(ctx context.Context, params port.ProductListParams) (*port.RawProductListPayload, error) {
	var filtered []platformProduct

	for _, p := range c.catalogue {
		if !matchesParams(p, params) {
			continue
		}
		filtered = append(filtered, p)
	}

	total := len(filtered)

	// Apply pagination.
	start := params.Offset
	if start > total {
		start = total
	}
	end := start + params.Limit
	if end > total {
		end = total
	}
	page := filtered[start:end]

	portProducts := make([]port.PlatformProduct, 0, len(page))
	for _, p := range page {
		portProducts = append(portProducts, toPortProduct(p))
	}

	return &port.RawProductListPayload{
		Products:   portProducts,
		TotalCount: total,
	}, nil
}

// FetchProductFilters returns all distinct colors, categories, and the global
// price min/max from the full catalogue (unfiltered).
func (c *InMemoryProductClient) FetchProductFilters(_ context.Context, _ port.ProductListParams) (*port.RawFilterPayload, error) {
	colorSet := make(map[string]struct{})
	categorySet := make(map[string]struct{})
	minPrice := math.MaxFloat64
	maxPrice := -math.MaxFloat64

	for _, p := range c.catalogue {
		for _, col := range p.Colors {
			colorSet[col] = struct{}{}
		}
		categorySet[p.Category] = struct{}{}
		if p.Price < minPrice {
			minPrice = p.Price
		}
		if p.Price > maxPrice {
			maxPrice = p.Price
		}
	}

	colors := make([]string, 0, len(colorSet))
	for col := range colorSet {
		colors = append(colors, col)
	}
	categories := make([]string, 0, len(categorySet))
	for cat := range categorySet {
		categories = append(categories, cat)
	}

	if len(c.catalogue) == 0 {
		minPrice = 0
		maxPrice = 0
	}

	return &port.RawFilterPayload{
		Colors:     colors,
		Categories: categories,
		MinPrice:   minPrice,
		MaxPrice:   maxPrice,
	}, nil
}

// FetchProductDetail finds a product by configSku.
// Returns nil, nil if the product is not found.
func (c *InMemoryProductClient) FetchProductDetail(_ context.Context, configSku string) (*port.RawProductDetailPayload, error) {
	for _, p := range c.catalogue {
		if p.ConfigSku == configSku {
			return &port.RawProductDetailPayload{Product: toPortProduct(p)}, nil
		}
	}
	return nil, nil
}

// toPortProduct converts an internal platformProduct to port.PlatformProduct.
func toPortProduct(p platformProduct) port.PlatformProduct {
	simples := make([]port.PlatformSimple, 0, len(p.Simples))
	for _, s := range p.Simples {
		simples = append(simples, port.PlatformSimple{
			SimpleSku: s.SimpleSku,
			Size:      s.Size,
			Color:     s.Color,
			Quantity:  s.Quantity,
		})
	}
	return port.PlatformProduct{
		ConfigSku: p.ConfigSku,
		Name:      p.Name,
		Brand:     p.Brand,
		Category:  p.Category,
		Price:     p.Price,
		Currency:  p.Currency,
		Colors:    p.Colors,
		Simples:   simples,
		ImageUrl:  p.ImageUrl,
		UrlKey:    p.UrlKey,
		Occasions: p.Occasions,
	}
}

// matchesParams returns true if the product satisfies all active filter criteria.
func matchesParams(p platformProduct, params port.ProductListParams) bool {
	// Query: case-insensitive substring match on Name or Brand.
	if params.Query != "" {
		q := strings.ToLower(params.Query)
		if !strings.Contains(strings.ToLower(p.Name), q) &&
			!strings.Contains(strings.ToLower(p.Brand), q) {
			return false
		}
	}

	// CategoryID: exact match.
	if params.CategoryID != "" && p.Category != params.CategoryID {
		return false
	}

	// Colors: at least one product color must be in the filter list.
	if len(params.Colors) > 0 {
		colorFilter := make(map[string]struct{}, len(params.Colors))
		for _, col := range params.Colors {
			colorFilter[col] = struct{}{}
		}
		found := false
		for _, col := range p.Colors {
			if _, ok := colorFilter[col]; ok {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// MinPrice: 0 means no lower bound.
	if params.MinPrice > 0 && p.Price < params.MinPrice {
		return false
	}

	// MaxPrice: 0 means no upper bound.
	if params.MaxPrice > 0 && p.Price > params.MaxPrice {
		return false
	}

	return true
}
