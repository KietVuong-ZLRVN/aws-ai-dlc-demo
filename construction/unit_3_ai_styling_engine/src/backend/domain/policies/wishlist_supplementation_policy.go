package policies

import (
	"ai-styling-engine/domain/events"
	"ai-styling-engine/domain/valueobjects"
)

// WishlistSupplementationPolicy handles WishlistFetchCompleted.
// It determines whether the wishlist alone is sufficient to form a combo.
// If not, it raises CatalogSupplementationRequired via the dispatcher.
type WishlistSupplementationPolicy struct {
	dispatcher events.EventDispatcher
}

func NewWishlistSupplementationPolicy(dispatcher events.EventDispatcher) *WishlistSupplementationPolicy {
	return &WishlistSupplementationPolicy{dispatcher: dispatcher}
}

func (p *WishlistSupplementationPolicy) Handle(event events.DomainEvent) {
	e := event.(events.WishlistFetchCompleted)

	inStock := e.Snapshot.InStockItems()
	if len(inStock) >= 2 {
		// Wishlist has enough items; no supplementation needed.
		return
	}

	// Collect configSkus for Complete-the-Look calls.
	skus := make([]valueobjects.Sku, len(inStock))
	for i, item := range inStock {
		skus[i] = item.ConfigSku
	}

	p.dispatcher.Dispatch(events.CatalogSupplementationRequired{
		SessionId:    e.SessionId,
		Filters:      valueobjects.CatalogSearchFilters{Limit: 20},
		WishlistSkus: skus,
	})
}
