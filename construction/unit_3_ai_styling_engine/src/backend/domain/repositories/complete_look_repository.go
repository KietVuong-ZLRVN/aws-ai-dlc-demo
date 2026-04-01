package repositories

import "ai-styling-engine/domain/valueobjects"

// CompleteLookRepository is the ACL port for fetching platform-computed
// "complete the look" styling signals for a given product.
type CompleteLookRepository interface {
	FetchCompleteLookSignals(configSku valueobjects.Sku) ([]valueobjects.ComboItem, error)
}
