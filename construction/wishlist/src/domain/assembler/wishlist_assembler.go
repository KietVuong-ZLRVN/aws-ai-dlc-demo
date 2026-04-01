package assembler

import (
	"fmt"
	"wishlist/domain/aggregate"
	"wishlist/domain/entity"
	"wishlist/domain/valueobject"
)

// RawWishlistItem is the raw struct used by the in-memory repository to store items.
type RawWishlistItem struct {
	ItemId    string
	SimpleSku string
	ConfigSku string
	Name      string
	Brand     string
	Price     float64
	Currency  string
	ImageUrl  string
	Color     string
	Size      string
	InStock   bool
}

type WishlistAssembler struct{}

// Assemble creates a Wishlist aggregate from raw items.
// WishlistId = "wl-" + shopperID.String()
func (a *WishlistAssembler) Assemble(shopperID valueobject.ShopperId, items []RawWishlistItem) (*aggregate.Wishlist, error) {
	wishlistId, err := valueobject.NewWishlistId("wl-" + shopperID.String())
	if err != nil {
		return nil, fmt.Errorf("assembling wishlist id: %w", err)
	}

	wishlistItems := make([]entity.WishlistItem, 0, len(items))
	for _, raw := range items {
		itemId, err := valueobject.NewWishlistItemId(raw.ItemId)
		if err != nil {
			return nil, fmt.Errorf("assembling item id %q: %w", raw.ItemId, err)
		}
		simpleSku, err := valueobject.NewSimpleSku(raw.SimpleSku)
		if err != nil {
			return nil, fmt.Errorf("assembling simple sku %q: %w", raw.SimpleSku, err)
		}
		configSku, err := valueobject.NewConfigSku(raw.ConfigSku)
		if err != nil {
			return nil, fmt.Errorf("assembling config sku %q: %w", raw.ConfigSku, err)
		}
		price, err := valueobject.NewMoney(raw.Price, raw.Currency)
		if err != nil {
			return nil, fmt.Errorf("assembling price for item %q: %w", raw.ItemId, err)
		}

		wishlistItems = append(wishlistItems, entity.WishlistItem{
			ItemId:    itemId,
			SimpleSku: simpleSku,
			ConfigSku: configSku,
			Name:      raw.Name,
			Brand:     raw.Brand,
			Price:     price,
			ImageUrl:  raw.ImageUrl,
			Color:     raw.Color,
			Size:      raw.Size,
			InStock:   raw.InStock,
		})
	}

	return &aggregate.Wishlist{
		ID:         wishlistId,
		ShopperID:  shopperID,
		Items:      wishlistItems,
		TotalCount: len(wishlistItems),
	}, nil
}
