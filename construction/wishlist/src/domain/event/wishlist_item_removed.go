package event

import "time"

type WishlistItemRemoved struct {
	ShopperID  string
	ConfigSku  string
	OccurredAt time.Time
}

func (e WishlistItemRemoved) EventName() string {
	return "WishlistItemRemoved"
}
