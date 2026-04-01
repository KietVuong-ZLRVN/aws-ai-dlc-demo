package usecases

import "ai-styling-engine/domain/valueobjects"

// GetPreferenceOptionsUseCase returns the static style options catalogue.
// No external calls, no command needed.
type GetPreferenceOptionsUseCase struct{}

func NewGetPreferenceOptionsUseCase() *GetPreferenceOptionsUseCase {
	return &GetPreferenceOptionsUseCase{}
}

func (uc *GetPreferenceOptionsUseCase) Execute() valueobjects.StyleOptionsCatalogue {
	return valueobjects.DefaultStyleOptionsCatalogue()
}
