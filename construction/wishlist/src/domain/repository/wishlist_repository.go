package repository

import (
	"context"
	"wishlist/domain/aggregate"
	"wishlist/domain/valueobject"
)

type WishlistRepository interface {
	GetByShopperId(ctx context.Context, shopperID valueobject.ShopperId, pagination valueobject.Pagination) (*aggregate.Wishlist, error)
	AddItem(ctx context.Context, shopperID valueobject.ShopperId, simpleSku valueobject.SimpleSku, configSku valueobject.ConfigSku) (valueobject.WishlistItemId, error)
	RemoveItemByConfigSku(ctx context.Context, shopperID valueobject.ShopperId, configSku valueobject.ConfigSku) error
}
