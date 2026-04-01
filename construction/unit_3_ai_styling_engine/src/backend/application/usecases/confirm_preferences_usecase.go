package usecases

import (
	"ai-styling-engine/application/commands"
	"ai-styling-engine/domain/aggregates"
	"ai-styling-engine/domain/events"
	"ai-styling-engine/domain/services"
	"ai-styling-engine/domain/valueobjects"
)

// ConfirmPreferencesUseCase handles POST /api/v1/style/preferences/confirm.
type ConfirmPreferencesUseCase struct {
	interpretationSvc services.PreferenceInterpretationService
	dispatcher        events.EventDispatcher
}

func NewConfirmPreferencesUseCase(
	interpretationSvc services.PreferenceInterpretationService,
	dispatcher events.EventDispatcher,
) *ConfirmPreferencesUseCase {
	return &ConfirmPreferencesUseCase{
		interpretationSvc: interpretationSvc,
		dispatcher:        dispatcher,
	}
}

func (uc *ConfirmPreferencesUseCase) Execute(cmd commands.ConfirmPreferencesCommand) (valueobjects.PreferenceSummary, error) {
	confirmation := aggregates.NewPreferenceConfirmation(cmd.Preferences, uc.dispatcher)
	return confirmation.Interpret(uc.interpretationSvc)
}
