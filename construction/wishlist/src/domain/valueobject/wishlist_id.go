package valueobject

import "errors"

type WishlistId struct {
	value string
}

func NewWishlistId(v string) (WishlistId, error) {
	if v == "" {
		return WishlistId{}, errors.New("wishlist id cannot be empty")
	}
	return WishlistId{value: v}, nil
}

func (w WishlistId) String() string {
	return w.value
}
