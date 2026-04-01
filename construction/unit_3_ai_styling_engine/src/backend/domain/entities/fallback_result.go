package entities

import "ai-styling-engine/domain/valueobjects"

// FallbackResult is produced when the StyleSession cannot form any valid Combo.
type FallbackResult struct {
	Message      string
	Alternatives []valueobjects.AlternativeItem
}

func NewFallbackResult(message string, alternatives []valueobjects.AlternativeItem) FallbackResult {
	return FallbackResult{Message: message, Alternatives: alternatives}
}
