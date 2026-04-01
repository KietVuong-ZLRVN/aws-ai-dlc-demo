package events

import "ai-styling-engine/domain/valueobjects"

const EventTypeCatalogSupplementationRequired = "CatalogSupplementationRequired"

type CatalogSupplementationRequired struct {
	SessionId   valueobjects.StyleSessionId
	Filters     valueobjects.CatalogSearchFilters
	WishlistSkus []valueobjects.Sku
}

func (e CatalogSupplementationRequired) EventType() string {
	return EventTypeCatalogSupplementationRequired
}
