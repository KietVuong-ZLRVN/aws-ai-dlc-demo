package acl

import "ai-styling-engine/domain/valueobjects"

// InMemoryProductCatalogRepository returns a fixed set of mock catalog items.
// In production this would call the Platform Product API GET /v1/products/list.
type InMemoryProductCatalogRepository struct{}

func NewInMemoryProductCatalogRepository() *InMemoryProductCatalogRepository {
	return &InMemoryProductCatalogRepository{}
}

func (r *InMemoryProductCatalogRepository) SearchSupplementaryItems(filters valueobjects.CatalogSearchFilters) ([]valueobjects.ComboItem, error) {
	return []valueobjects.ComboItem{
		{
			ConfigSku: "CFG-LOAFERS-BGE",
			SimpleSku: "SKU-LOAFERS-BGE-38",
			Name:      "Leather Loafers",
			Brand:     "Mango",
			Price:     79.90,
			ImageUrl:  "https://example.com/images/leather-loafers-beige.jpg",
			Source:    valueobjects.ItemSourceCatalog,
		},
		{
			ConfigSku: "CFG-TOTE-WHT",
			SimpleSku: "SKU-TOTE-WHT-OS",
			Name:      "Canvas Tote Bag",
			Brand:     "Toteme",
			Price:     59.90,
			ImageUrl:  "https://example.com/images/canvas-tote-white.jpg",
			Source:    valueobjects.ItemSourceCatalog,
		},
	}, nil
}
