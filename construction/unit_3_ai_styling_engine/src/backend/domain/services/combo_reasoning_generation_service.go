package services

import "ai-styling-engine/domain/valueobjects"

// ComboReasoningGenerationService generates human-readable 1–3 sentence explanations
// for each scored combo. It does not rank or filter — reasoning only.
type ComboReasoningGenerationService interface {
	GenerateReasoning(candidate ComboCandidate, preferences *valueobjects.StylePreferences) (valueobjects.ComboReasoning, error)
}
