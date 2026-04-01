package services

import "ai-styling-engine/domain/valueobjects"

// ScoringInput contains all data the scoring service needs to produce combos.
type ScoringInput struct {
	WishlistItems       []valueobjects.WishlistItem
	SupplementaryItems  []valueobjects.ComboItem
	CompleteLookSignals []valueobjects.ComboItem
	Preferences         *valueobjects.StylePreferences
	ExcludedComboIds    valueobjects.ExcludedComboIds
	QuickGenerate       bool
}

// ComboCandidate is an undecorated combo produced by the scoring service (no reasoning yet).
type ComboCandidate struct {
	Id    string
	Items []valueobjects.ComboItem
	Score float64
}

// ScoringResult is a sealed result type — either a list of combo candidates or a fallback.
type ScoringResult struct {
	Candidates []ComboCandidate
	Fallback   *ScoringFallback
}

type ScoringFallback struct {
	Message      string
	Alternatives []valueobjects.AlternativeItem
}

func (r ScoringResult) IsFallback() bool {
	return r.Fallback != nil
}
