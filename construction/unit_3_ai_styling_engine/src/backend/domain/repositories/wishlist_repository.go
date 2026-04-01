package repositories

import "ai-styling-engine/domain/valueobjects"

// WishlistRepository is the ACL port for fetching the shopper's wishlist from Unit 2.
type WishlistRepository interface {
	FetchForSession(session valueobjects.ShopperSession) (valueobjects.WishlistSnapshot, error)
}
