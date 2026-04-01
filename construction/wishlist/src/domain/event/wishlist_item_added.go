package event

import "time"

// Event is the base interface all domain events implement.
type Event interface {
	EventName() string
}

type WishlistItemAdded struct {
	ShopperID  string
	SimpleSku  string
	ConfigSku  string
	ItemId     string
	OccurredAt time.Time
}

func (e WishlistItemAdded) EventName() string {
	return "WishlistItemAdded"
}
