package events

import "ai-styling-engine/domain/valueobjects"

const EventTypePreferencesInterpreted = "PreferencesInterpreted"

type PreferencesInterpreted struct {
	Summary valueobjects.PreferenceSummary
}

func (e PreferencesInterpreted) EventType() string {
	return EventTypePreferencesInterpreted
}
