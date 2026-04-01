package events

import "ai-styling-engine/domain/valueobjects"

const EventTypeWishlistFetchCompleted = "WishlistFetchCompleted"

type WishlistFetchCompleted struct {
	SessionId valueobjects.StyleSessionId
	Snapshot  valueobjects.WishlistSnapshot
}

func (e WishlistFetchCompleted) EventType() string {
	return EventTypeWishlistFetchCompleted
}
