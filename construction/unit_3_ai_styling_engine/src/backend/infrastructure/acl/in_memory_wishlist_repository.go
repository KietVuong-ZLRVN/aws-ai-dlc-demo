package acl

import "ai-styling-engine/domain/valueobjects"

// InMemoryWishlistRepository returns a fixed set of realistic mock wishlist items.
// In production this would call Unit 2's GET /api/v1/wishlist.
type InMemoryWishlistRepository struct{}

func NewInMemoryWishlistRepository() *InMemoryWishlistRepository {
	return &InMemoryWishlistRepository{}
}

func (r *InMemoryWishlistRepository) FetchForSession(session valueobjects.ShopperSession) (valueobjects.WishlistSnapshot, error) {
	items := []valueobjects.WishlistItem{
		{
			ItemId:    "item-001",
			SimpleSku: "SKU-BLAZER-BLK-M",
			ConfigSku: "CFG-BLAZER-BLK",
			Name:      "Linen Blazer",
			Brand:     "Arket",
			Price:     129.90,
			ImageUrl:  "https://example.com/images/linen-blazer-black.jpg",
			Color:     "black",
			Size:      "M",
			InStock:   true,
		},
		{
			ItemId:    "item-002",
			SimpleSku: "SKU-TROUSERS-BGE-M",
			ConfigSku: "CFG-TROUSERS-BGE",
			Name:      "Wide-Leg Trousers",
			Brand:     "& Other Stories",
			Price:     89.90,
			ImageUrl:  "https://example.com/images/wide-leg-trousers-beige.jpg",
			Color:     "beige",
			Size:      "M",
			InStock:   true,
		},
		{
			ItemId:    "item-003",
			SimpleSku: "SKU-SCARF-WHT-OS",
			ConfigSku: "CFG-SCARF-WHT",
			Name:      "Silk Scarf",
			Brand:     "COS",
			Price:     49.90,
			ImageUrl:  "https://example.com/images/silk-scarf-white.jpg",
			Color:     "white",
			Size:      "One Size",
			InStock:   false,
		},
	}
	return valueobjects.WishlistSnapshot{Items: items, TotalCount: len(items)}, nil
}
