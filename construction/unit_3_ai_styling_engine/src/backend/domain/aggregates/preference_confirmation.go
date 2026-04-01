package aggregates

import (
	"ai-styling-engine/domain/events"
	"ai-styling-engine/domain/services"
	"ai-styling-engine/domain/valueobjects"
)

// PreferenceConfirmation is a transient aggregate representing one round-trip
// where the shopper submits preferences and receives an AI-generated summary.
type PreferenceConfirmation struct {
	preferences valueobjects.StylePreferences
	summary     *valueobjects.PreferenceSummary
	dispatcher  events.EventDispatcher
}

func NewPreferenceConfirmation(
	preferences valueobjects.StylePreferences,
	dispatcher events.EventDispatcher,
) *PreferenceConfirmation {
	return &PreferenceConfirmation{
		preferences: preferences,
		dispatcher:  dispatcher,
	}
}

// Interpret calls the PreferenceInterpretationService and stores the resulting summary.
func (pc *PreferenceConfirmation) Interpret(svc services.PreferenceInterpretationService) (valueobjects.PreferenceSummary, error) {
	summary, err := svc.Interpret(pc.preferences)
	if err != nil {
		return valueobjects.PreferenceSummary{}, err
	}
	pc.summary = &summary
	pc.dispatcher.Dispatch(events.PreferencesInterpreted{Summary: summary})
	return summary, nil
}
