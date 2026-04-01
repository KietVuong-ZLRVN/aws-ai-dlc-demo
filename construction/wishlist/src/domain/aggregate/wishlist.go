package aggregate

import (
	"errors"
	"strings"
	"wishlist/domain/entity"
	"wishlist/domain/valueobject"
)

var ErrWishlistItemAlreadyPresent = errors.New("item already in wishlist")

type AddItemIntent struct {
	SimpleSku valueobject.SimpleSku
	ConfigSku valueobject.ConfigSku
}

type RemoveItemIntent struct {
	ConfigSku valueobject.ConfigSku
}

type Wishlist struct {
	ID         valueobject.WishlistId
	ShopperID  valueobject.ShopperId
	Items      []entity.WishlistItem
	TotalCount int
}

// deriveConfigSku extracts the config SKU from a simple SKU by taking the first 2 dash-separated parts.
// e.g., "PD-005-S-NVY" -> "PD-005". If fewer than 2 parts, uses the whole value.
func deriveConfigSku(simpleSku valueobject.SimpleSku) (valueobject.ConfigSku, error) {
	parts := strings.Split(simpleSku.String(), "-")
	var configSkuValue string
	if len(parts) >= 2 {
		configSkuValue = strings.Join(parts[:2], "-")
	} else {
		configSkuValue = simpleSku.String()
	}
	return valueobject.NewConfigSku(configSkuValue)
}

// AddItem checks for duplicate by ConfigSku.
// Returns (AddItemIntent, nil) or (zero, ErrWishlistItemAlreadyPresent).
func (w *Wishlist) AddItem(simpleSku valueobject.SimpleSku) (AddItemIntent, error) {
	configSku, err := deriveConfigSku(simpleSku)
	if err != nil {
		return AddItemIntent{}, err
	}

	for _, item := range w.Items {
		if item.ConfigSku.String() == configSku.String() {
			return AddItemIntent{}, ErrWishlistItemAlreadyPresent
		}
	}

	return AddItemIntent{
		SimpleSku: simpleSku,
		ConfigSku: configSku,
	}, nil
}

// RemoveItem always returns RemoveItemIntent. If item not found, it's a no-op.
func (w *Wishlist) RemoveItem(configSku valueobject.ConfigSku) RemoveItemIntent {
	return RemoveItemIntent{ConfigSku: configSku}
}

// ToggleItem checks if configSku exists; if yes calls RemoveItem, if no calls AddItem.
// Returns (interface{}, error) where interface{} is either AddItemIntent or RemoveItemIntent.
func (w *Wishlist) ToggleItem(simpleSku valueobject.SimpleSku, configSku valueobject.ConfigSku) (interface{}, error) {
	for _, item := range w.Items {
		if item.ConfigSku.String() == configSku.String() {
			return w.RemoveItem(configSku), nil
		}
	}
	return w.AddItem(simpleSku)
}
