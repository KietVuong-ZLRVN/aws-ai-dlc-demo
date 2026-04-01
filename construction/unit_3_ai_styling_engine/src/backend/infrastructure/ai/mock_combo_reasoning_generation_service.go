package ai

import (
	"ai-styling-engine/domain/services"
	"ai-styling-engine/domain/valueobjects"
	"fmt"
)

// MockComboReasoningGenerationService returns realistic hardcoded reasoning text.
// Replaces AWS Bedrock LLM calls for local development.
type MockComboReasoningGenerationService struct{}

func NewMockComboReasoningGenerationService() *MockComboReasoningGenerationService {
	return &MockComboReasoningGenerationService{}
}

func (s *MockComboReasoningGenerationService) GenerateReasoning(
	candidate services.ComboCandidate,
	preferences *valueobjects.StylePreferences,
) (valueobjects.ComboReasoning, error) {
	if len(candidate.Items) == 0 {
		return valueobjects.NewComboReasoning("A well-balanced combination of pieces.")
	}

	first := candidate.Items[0].Name
	var second string
	if len(candidate.Items) > 1 {
		second = candidate.Items[1].Name
	}

	var text string
	if preferences != nil && len(preferences.Occasions) > 0 {
		text = fmt.Sprintf(
			"The %s and %s work beautifully together for a %s look — balanced proportions with a cohesive colour story.",
			first, second, preferences.Occasions[0],
		)
	} else {
		text = fmt.Sprintf(
			"The %s pairs effortlessly with the %s, creating a polished everyday outfit with complementary tones.",
			first, second,
		)
	}

	return valueobjects.NewComboReasoning(text)
}
