package entity

import "wishlist/domain/valueobject"

type WishlistItem struct {
	ItemId    valueobject.WishlistItemId
	SimpleSku valueobject.SimpleSku
	ConfigSku valueobject.ConfigSku
	Name      string
	Brand     string
	Price     valueobject.Money
	ImageUrl  string
	Color     string
	Size      string
	InStock   bool
}
