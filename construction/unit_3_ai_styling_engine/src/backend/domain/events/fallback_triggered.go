package events

import "ai-styling-engine/domain/valueobjects"

const EventTypeFallbackTriggered = "FallbackTriggered"

type FallbackTriggered struct {
	SessionId    valueobjects.StyleSessionId
	Alternatives []valueobjects.AlternativeItem
}

func (e FallbackTriggered) EventType() string {
	return EventTypeFallbackTriggered
}
