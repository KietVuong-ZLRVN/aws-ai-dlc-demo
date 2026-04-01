package acl

import "ai-styling-engine/domain/valueobjects"

// InMemoryCompleteLookRepository returns fixed mock "complete the look" signals.
// In production this would call the Platform Recommendation API
// GET /v1/recommendation/completethelook/{config_sku}.
type InMemoryCompleteLookRepository struct{}

func NewInMemoryCompleteLookRepository() *InMemoryCompleteLookRepository {
	return &InMemoryCompleteLookRepository{}
}

func (r *InMemoryCompleteLookRepository) FetchCompleteLookSignals(configSku valueobjects.Sku) ([]valueobjects.ComboItem, error) {
	return []valueobjects.ComboItem{
		{
			ConfigSku: "CFG-BELT-BLK",
			SimpleSku: "SKU-BELT-BLK-OS",
			Name:      "Woven Leather Belt",
			Brand:     "Zara",
			Price:     35.90,
			ImageUrl:  "https://example.com/images/leather-belt-black.jpg",
			Source:    valueobjects.ItemSourceCatalog,
		},
	}, nil
}
