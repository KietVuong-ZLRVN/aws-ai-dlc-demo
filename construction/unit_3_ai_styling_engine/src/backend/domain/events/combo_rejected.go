package events

import "ai-styling-engine/domain/valueobjects"

const EventTypeComboRejected = "ComboRejected"

type ComboRejected struct {
	SessionId valueobjects.StyleSessionId
	ComboId   string
}

func (e ComboRejected) EventType() string {
	return EventTypeComboRejected
}
