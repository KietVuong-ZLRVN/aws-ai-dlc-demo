package services

// ComboCompatibilityScoringService evaluates fashion compatibility between items
// and produces ranked combo candidates. It does not generate reasoning text.
type ComboCompatibilityScoringService interface {
	Score(input ScoringInput) (ScoringResult, error)
}
