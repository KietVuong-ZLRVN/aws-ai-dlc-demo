package services

import "ai-styling-engine/domain/valueobjects"

// PreferenceInterpretationService converts structured StylePreferences into
// a natural-language PreferenceSummary for the shopper to review.
type PreferenceInterpretationService interface {
	Interpret(preferences valueobjects.StylePreferences) (valueobjects.PreferenceSummary, error)
}
