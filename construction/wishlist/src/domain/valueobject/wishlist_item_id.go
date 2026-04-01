package valueobject

import "errors"

type WishlistItemId struct {
	value string
}

func NewWishlistItemId(v string) (WishlistItemId, error) {
	if v == "" {
		return WishlistItemId{}, errors.New("wishlist item id cannot be empty")
	}
	return WishlistItemId{value: v}, nil
}

func (w WishlistItemId) String() string {
	return w.value
}
