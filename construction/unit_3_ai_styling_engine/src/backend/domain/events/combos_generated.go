package events

import "ai-styling-engine/domain/valueobjects"

const EventTypeCombosGenerated = "CombosGenerated"

type CombosGenerated struct {
	SessionId  valueobjects.StyleSessionId
	ComboCount int
}

func (e CombosGenerated) EventType() string {
	return EventTypeCombosGenerated
}
