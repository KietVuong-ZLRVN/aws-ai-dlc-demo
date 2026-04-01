package repositories

import "ai-styling-engine/domain/valueobjects"

// ProductCatalogRepository is the ACL port for searching supplementary catalog items
// from the Platform Product API.
type ProductCatalogRepository interface {
	SearchSupplementaryItems(filters valueobjects.CatalogSearchFilters) ([]valueobjects.ComboItem, error)
}
