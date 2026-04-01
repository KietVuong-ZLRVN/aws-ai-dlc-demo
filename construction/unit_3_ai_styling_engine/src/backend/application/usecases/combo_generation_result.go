package usecases

import "ai-styling-engine/domain/entities"

// ComboGenerationResult is the sealed output of GenerateCombosUseCase.
// Exactly one of Success or Fallback will be non-nil.
type ComboGenerationResult struct {
	Success  *ComboGenerationSuccess
	Fallback *ComboGenerationFallback
}

type ComboGenerationSuccess struct {
	Combos    []entities.Combo
	Exhausted bool
}

type ComboGenerationFallback struct {
	FallbackResult entities.FallbackResult
}

func (r ComboGenerationResult) IsSuccess() bool {
	return r.Success != nil
}
