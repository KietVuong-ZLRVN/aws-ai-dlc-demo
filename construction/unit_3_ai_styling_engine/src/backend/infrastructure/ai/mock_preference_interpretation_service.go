package ai

import (
	"ai-styling-engine/domain/valueobjects"
	"fmt"
	"strings"
)

// MockPreferenceInterpretationService returns a realistic hardcoded preference summary.
// Replaces AWS Bedrock LLM calls for local development.
type MockPreferenceInterpretationService struct{}

func NewMockPreferenceInterpretationService() *MockPreferenceInterpretationService {
	return &MockPreferenceInterpretationService{}
}

func (s *MockPreferenceInterpretationService) Interpret(
	prefs valueobjects.StylePreferences,
) (valueobjects.PreferenceSummary, error) {
	if prefs.IsEmpty() {
		return valueobjects.PreferenceSummary{
			Text:        "You're open to any style — we'll pick the best combo from your wishlist.",
			Preferences: prefs,
		}, nil
	}

	var parts []string

	if len(prefs.Styles) > 0 {
		styles := make([]string, len(prefs.Styles))
		for i, s := range prefs.Styles {
			styles[i] = string(s)
		}
		parts = append(parts, strings.Join(styles, ", "))
	}

	if len(prefs.Occasions) > 0 {
		occasions := make([]string, len(prefs.Occasions))
		for i, o := range prefs.Occasions {
			occasions[i] = string(o)
		}
		parts = append(parts, strings.Join(occasions, " or ")+" look")
	}

	if prefs.Budget != nil {
		parts = append(parts, fmt.Sprintf("between $%.0f–$%.0f", float64(prefs.Budget.Min), float64(prefs.Budget.Max)))
	}

	if len(prefs.Colors.Preferred) > 0 {
		colors := make([]string, len(prefs.Colors.Preferred))
		for i, c := range prefs.Colors.Preferred {
			colors[i] = string(c)
		}
		parts = append(parts, "in "+strings.Join(colors, " and ")+" tones")
	}

	summary := "You're looking for a " + strings.Join(parts, ", ") + "."
	if prefs.FreeText != "" {
		summary = strings.TrimSuffix(summary, ".") + " — " + prefs.FreeText + "."
	}
	return valueobjects.PreferenceSummary{Text: summary, Preferences: prefs}, nil
}
