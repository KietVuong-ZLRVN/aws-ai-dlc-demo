package events

import "ai-styling-engine/domain/valueobjects"

const EventTypeComboGenerationRequested = "ComboGenerationRequested"

type ComboGenerationRequested struct {
	SessionId    valueobjects.StyleSessionId
	Preferences  *valueobjects.StylePreferences
	ExcludedIds  valueobjects.ExcludedComboIds
	QuickGenerate bool
}

func (e ComboGenerationRequested) EventType() string {
	return EventTypeComboGenerationRequested
}
